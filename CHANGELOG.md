# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Complete OAuth2 Authorization Server implementation in Go.
- Support for Authorization Code Flow (with PKCE), Client Credentials Flow, Refresh Token Flow, Device Authorization Flow, and JWT Bearer Token Flow.
- MongoDB integration for persistent storage of clients, users, and tokens.
- Redis integration for session management and rate limiting.
- Admin dashboard UI for managing OAuth2 clients and users.
- Prometheus metrics endpoint for monitoring.
- Docker Compose setup for easy local development and deployment.
- Initial project documentation (API, Architecture, Deployment, Flows).
- Open Source community files (Code of Conduct, Contributing guide, Security policy).

### Security
- HSTS header (`Strict-Transport-Security`) is now enabled in production with a 2-year max-age and `includeSubDomains` (RFC 6797).
- Session cookies now enforce `Secure=true`, `HttpOnly=true`, and `SameSite=Lax` in production (RFC 6265).
- Cookie-clearing operations (logout, invalid session) now include matching security attributes so browsers correctly delete the cookie.
- Added `Referrer-Policy: strict-origin-when-cross-origin` header to prevent leaking OAuth parameters in referrer URLs.
- Added `Content-Security-Policy` header with restrictive defaults suitable for an authorization server.
- CORS debug logging is now disabled in production to prevent leaking internal routing information (RFC 6454).

### Changed
- Health check endpoint split into two separate endpoints for container orchestration:
  - `GET /health/live` returns 200 if the process is running (Kubernetes `livenessProbe`).
  - `GET /health/ready` checks MongoDB and Redis connectivity, returns 200 or 503 with per-component status (Kubernetes `readinessProbe`).
- Health check response format now follows the [draft-inadarei-api-health-check](https://inadarei.github.io/rfc-healthcheck/) convention with `status`, `components`, and `time` fields.
- Security headers middleware is now environment-aware, accepting `appEnv` to conditionally apply production-only headers.

### Fixed
- Layout rendering bug causing 500 errors in the device authorization consent flow.
