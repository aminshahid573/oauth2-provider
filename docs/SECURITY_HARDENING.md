# Security Hardening and Health Probes

This document covers the production security defaults, HTTP security headers, cookie policies, and Kubernetes-compatible health check endpoints provided by Authexa.

All security behaviors are environment-aware. Settings that could interfere with local development (HSTS, forced secure cookies) are only activated when `APP_ENV=production`.

---

## HTTP Security Headers

Every response from the server includes the following headers:

| Header | Value | When | Standard |
| :--- | :--- | :--- | :--- |
| `X-Frame-Options` | `DENY` | Always | [OWASP Secure Headers](https://owasp.org/www-project-secure-headers/) |
| `X-Content-Type-Options` | `nosniff` | Always | [OWASP Secure Headers](https://owasp.org/www-project-secure-headers/) |
| `X-XSS-Protection` | `1; mode=block` | Always | Legacy browser compatibility |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Always | [W3C Referrer Policy](https://www.w3.org/TR/referrer-policy/) |
| `Content-Security-Policy` | `default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; frame-ancestors 'none'; form-action 'self'; base-uri 'self'` | Always | [W3C CSP Level 3](https://www.w3.org/TR/CSP3/) |
| `Strict-Transport-Security` | `max-age=63072000; includeSubDomains` | Production only | [RFC 6797](https://www.rfc-editor.org/rfc/rfc6797) |

### Why these matter for an authorization server

- **HSTS** prevents SSL-stripping attacks. An authorization server issues tokens over HTTPS; downgrade attacks would expose credentials and tokens in transit.
- **Referrer-Policy** prevents OAuth parameters (authorization codes, tokens in fragment URIs) from leaking via the `Referer` header during redirects.
- **CSP with `frame-ancestors 'none'`** provides clickjacking protection equivalent to `X-Frame-Options: DENY` but with broader browser support for CSP Level 3.
- **`form-action 'self'`** prevents form data (login credentials, consent approvals) from being submitted to third-party origins.

---

## Session Cookie Policy

Session cookies (`session_id`) are set with the following attributes:

| Attribute | Value | Standard |
| :--- | :--- | :--- |
| `HttpOnly` | `true` (always) | [RFC 6265 section 4.1.2.6](https://www.rfc-editor.org/rfc/rfc6265#section-4.1.2.6) |
| `Secure` | `true` in production, TLS-derived in development | [RFC 6265 section 4.1.2.5](https://www.rfc-editor.org/rfc/rfc6265#section-4.1.2.5) |
| `SameSite` | `Lax` (always) | [RFC 6265bis](https://httpwg.org/http-extensions/draft-ietf-httpbis-rfc6265bis.html) |
| `Path` | `/` | [RFC 6265 section 4.1.2.4](https://www.rfc-editor.org/rfc/rfc6265#section-4.1.2.4) |

Cookie-clearing operations (logout, invalid session detection) set the same attributes on the deletion cookie. Browsers require attribute matching to correctly remove cookies (RFC 6265 section 4.1.2).

### Production vs Development

In development (`APP_ENV=development`), the `Secure` attribute is derived from `r.TLS != nil`, allowing local HTTP setups to work without TLS. In production, `Secure=true` is always forced regardless of the TLS termination point, which correctly handles reverse-proxy deployments where TLS is terminated at a load balancer.

---

## CORS Configuration

CORS debug logging is disabled in production to prevent internal routing information from appearing in server logs. In development, debug logging remains active to assist with troubleshooting cross-origin requests.

The CORS middleware supports configurable allowed origins via the `CORS_ALLOWED_ORIGINS` environment variable (comma-separated list).

**Relevant standards:** [RFC 6454 (Web Origin Concept)](https://www.rfc-editor.org/rfc/rfc6454), [Fetch Living Standard (CORS Protocol)](https://fetch.spec.whatwg.org/#http-cors-protocol).

---

## Health Check Endpoints

Two separate endpoints support Kubernetes-style container orchestration, following the [draft-inadarei-api-health-check](https://inadarei.github.io/rfc-healthcheck/) response format.

### `GET /health/live`

Returns `200 OK` if the Go process is running. No dependency checks are performed. This endpoint is intended for the Kubernetes `livenessProbe`.

```json
{
  "status": "pass",
  "time": "2025-01-15T10:30:00Z"
}
```

### `GET /health/ready`

Checks MongoDB and Redis connectivity (each with a 2-second timeout). Returns `200 OK` with per-component status if all dependencies are reachable. This endpoint is intended for the Kubernetes `readinessProbe`.

**Healthy response (200 OK):**
```json
{
  "status": "pass",
  "components": {
    "mongodb": { "status": "pass" },
    "redis": { "status": "pass" }
  },
  "time": "2025-01-15T10:30:00Z"
}
```

**Unhealthy response (503 Service Unavailable):**
```json
{
  "status": "fail",
  "components": {
    "mongodb": { "status": "pass" },
    "redis": { "status": "fail", "detail": "dial tcp 127.0.0.1:6379: connect: connection refused" }
  },
  "time": "2025-01-15T10:30:00Z"
}
```

The `503` response includes a `Retry-After: 5` header per [RFC 9110 section 15.6.4](https://www.rfc-editor.org/rfc/rfc9110#section-15.6.4).

Both endpoints set `Cache-Control: no-store` to prevent health data from being cached.

---

## Access Token Revocation (JTI Revocation List)

JWT access tokens are stateless by design. Once issued, a valid JWT cannot be invalidated before its `exp` claim without a server-side check. Authexa addresses this with a Redis-backed JTI (JWT ID) revocation list.

### How it works

1. Every access token includes a unique `jti` claim (UUID, per [RFC 7519 section 4.1.7](https://www.rfc-editor.org/rfc/rfc7519#section-4.1.7)).
2. When `POST /oauth2/revoke` receives a JWT access token, the handler writes the token's `jti` to Redis with a TTL equal to `exp - now` (the token's remaining lifetime).
3. When `POST /oauth2/introspect` receives a JWT access token, the handler checks Redis for the `jti` before returning `active: true`.
4. Redis entries expire automatically once the original token would have expired, keeping the revocation list small.

### Fail-closed behavior

If the Redis revocation store is unreachable during introspection, the token is reported as `active: false`. This fail-closed approach prevents revoked tokens from being accepted during infrastructure outages.

### Revocation store interface

```go
type RevocationStore interface {
    Revoke(ctx context.Context, jti string, ttl time.Duration) error
    IsRevoked(ctx context.Context, jti string) (bool, error)
}
```

The Redis implementation uses keys prefixed with `revoked_jti:` and leverages Redis TTL for automatic cleanup.

**Relevant standards:** [RFC 7009 (Token Revocation)](https://www.rfc-editor.org/rfc/rfc7009), [RFC 7662 section 2.2 (Introspection)](https://www.rfc-editor.org/rfc/rfc7662#section-2.2), [RFC 9068 section 2.2 (JWT Access Token Profile)](https://www.rfc-editor.org/rfc/rfc9068#section-2.2).

---

## Deterministic Key IDs (RFC 7638 JWK Thumbprint)

The `kid` (Key ID) in the JWKS and JWT headers is derived from the public key using the [RFC 7638 JWK Thumbprint](https://www.rfc-editor.org/rfc/rfc7638) algorithm (SHA-256). This replaces the previous approach of generating a random UUID on every server startup.

### Why this matters

- **JWKS stability:** Clients and resource servers cache the JWKS endpoint (per its `Cache-Control: public, max-age=3600` header). A random `kid` that changes on every restart invalidates all cached keys, causing token verification failures until the cache refreshes.
- **Determinism:** The same RSA key always produces the same `kid`, whether loaded from disk, a secrets manager, or re-deployed to a different server instance.
- **Key rotation readiness:** When key rotation is implemented, the deterministic `kid` ensures that tokens signed with the old key can still be verified by looking up the correct key in the JWKS set.

**Relevant standards:** [RFC 7638 (JWK Thumbprint)](https://www.rfc-editor.org/rfc/rfc7638), [RFC 7517 section 4.5 (JWK kid parameter)](https://www.rfc-editor.org/rfc/rfc7517#section-4.5).

---

## Per-Client Rate Limiting

The per-client rate limiter resolves `client_id` from two sources, checked in order:

1. **HTTP Basic Auth header** (RFC 6749 section 2.3.1) -- the username component of `Authorization: Basic base64(client_id:client_secret)`.
2. **Form body parameter** -- `client_id` in the `application/x-www-form-urlencoded` request body, used by public clients.

This ensures that confidential clients authenticating via Basic Auth are correctly identified for rate limiting, rather than being bucketed as anonymous.

**Relevant standards:** [RFC 6749 section 2.3.1](https://www.rfc-editor.org/rfc/rfc6749#section-2.3.1), [RFC 7617 (Basic Auth)](https://www.rfc-editor.org/rfc/rfc7617), [RFC 6819 section 4.4.1.1 (brute-force countermeasures)](https://www.rfc-editor.org/rfc/rfc6819#section-4.4.1.1).

### Kubernetes Configuration Example

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  failureThreshold: 3
```

---

| [![Previous](https://img.shields.io/badge/←_Previous-1f6feb?style=for-the-badge&logo=none&logoColor=white&labelColor=1f6feb&color=1f6feb)](DEPLOYMENT.md) <br> <sub>DEPLOYMENT.md</sub> | [![Next](https://img.shields.io/badge/Next_→-1f6feb?style=for-the-badge&logo=none&logoColor=white&labelColor=1f6feb&color=1f6feb)](REFERENCES.md) <br> <sub>REFERENCES.md</sub> |
|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
