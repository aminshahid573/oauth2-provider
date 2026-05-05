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
