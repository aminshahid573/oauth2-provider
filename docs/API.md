# API Reference Guide

This document provides a complete technical reference for all machine-to-machine (M2M) and administrative API endpoints of **Authexa**.

## Common Conventions

### Base URL
All endpoint paths in this document are relative to the server's base URL. For local development, this is `http://localhost:8080`.

### Authentication Methods
Different endpoints require different authentication methods:
- **Client Authentication (Basic Auth)**: Used by clients to authenticate themselves. Send the `client_id` and `client_secret` as an HTTP Basic Auth header.
- **Bearer Token**: Used by clients to access protected resources (like `/oauth2/userinfo`) on behalf of a user. The token is sent via the `Authorization: Bearer <access_token>` header.
- **Session Cookie**: Used by the Admin API to authenticate administrative users via the browser. Requests must include the `session_id` cookie.

---

## Category 1: Core OAuth2 & OIDC Endpoints

### Endpoint: `POST /oauth2/token`
The central endpoint for exchanging codes or credentials for access tokens, refresh tokens, and OIDC ID tokens.
- **Content-Type**: `application/x-www-form-urlencoded`

#### Grant Type: `authorization_code` (with PKCE)
**Request Body:**
| Parameter | Required | Description |
|---|---|---|
| `grant_type` | **Yes** | Must be `authorization_code`. |
| `code` | **Yes** | The authorization code from the `/oauth2/authorize` redirect. |
| `redirect_uri` | **Yes** | Must exactly match the `redirect_uri` used in the initial request. |
| `client_id` | **Yes** | The client's unique identifier. |
| `client_secret` | **Yes** | The client's secret. |
| `code_verifier` | **Yes** | The PKCE secret generated at the start of the flow. (Authexa mandates S256 PKCE). |

**Success Response (`200 OK`):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6Im...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_token": "a_very_long_and_secure_refresh_token_string...",
  "scope": "openid profile",
  "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI..." // Included ONLY if 'openid' scope was granted
}
```

#### Grant Type: `client_credentials`
**Request Body:**
| Parameter | Required | Description |
|---|---|---|
| `grant_type` | **Yes** | Must be `client_credentials`. |
| `client_id` | **Yes** | The client's ID. |
| `client_secret` | **Yes** | The client's secret. |
| `scope` | No | A space-delimited list of scopes. |

#### Grant Type: `refresh_token`
**Request Body:**
| Parameter | Required | Description |
|---|---|---|
| `grant_type` | **Yes** | Must be `refresh_token`. |
| `refresh_token` | **Yes** | The refresh token obtained from a previous flow. |
| `client_id` | **Yes** | The client's ID. |
| `client_secret` | **Yes** | The client's secret. |

#### Grant Type: `urn:ietf:params:oauth:grant-type:device_code`
Used by devices to poll for tokens after displaying a code.
**Request Body:**
| Parameter | Required | Description |
|---|---|---|
| `grant_type` | **Yes** | Must be `urn:ietf:params:oauth:grant-type:device_code`. |
| `device_code` | **Yes** | The `device_code` obtained from `/oauth2/device_authorization`. |
| `client_id` | **Yes** | The client's ID. |

*(Returns `428 Precondition Required` with `authorization_pending` while waiting for user approval).*

---

### Endpoint: `POST /oauth2/device_authorization`
Starts the Device Authorization Flow.
- **Content-Type**: `application/x-www-form-urlencoded`

**Request Body:**
| Parameter | Required | Description |
|---|---|---|
| `client_id` | **Yes** | The client's ID. |
| `scope` | No | A space-delimited list of scopes (e.g. `openid profile`). |

**Success Response (`200 OK`):**
```json
{
  "device_code": "a_long_secret_string_for_the_device",
  "user_code": "ABCDEFGH",
  "verification_uri": "http://localhost:8080/device",
  "expires_in": 900,
  "interval": 5
}
```

---

### Endpoint: `POST /oauth2/introspect`
Allows a resource server to validate an access token OR an opaque refresh token.
- **Authentication**: HTTP Basic Auth (`-u client_id:client_secret`).
- **Content-Type**: `application/x-www-form-urlencoded`

**Request Body:**
| Parameter | Required | Description |
|---|---|---|
| `token` | **Yes** | The token to validate (Access or Refresh token). |

**Success Response (`200 OK`, Active Token):**
```json
{
  "active": true,
  "scope": "openid profile",
  "client_id": "test-client",
  "sub": "6675d3a...",
  "exp": 1755855000,
  "iat": 1755854100,
  "token_type": "Bearer"
}
```

---

### Endpoint: `POST /oauth2/revoke`
Invalidates a refresh token so it can no longer be used.
- **Authentication**: HTTP Basic Auth (`-u client_id:client_secret`).

**Request Body:**
| Parameter | Required | Description |
|---|---|---|
| `token` | **Yes** | The `refresh_token` to revoke. |

---

### Endpoint: `GET /oauth2/userinfo`
Retrieves information about the user associated with an access token.
- **Authentication**: Bearer Token (`Authorization: Bearer <access_token>`).

**Success Response (`200 OK`):**
```json
{
  "sub": "6675d3a...",
  "name": "testuser",
  "preferred_username": "testuser",
  "email": "testuser@example.com",
  "email_verified": false
}
```

---

### Endpoints: Metadata & Discovery
- **`GET /.well-known/oauth-authorization-server`**: Returns the OAuth2 Authorization Server Metadata configuration.
- **`GET /.well-known/jwks.json`**: Returns the JSON Web Key Set (JWKS) containing the public keys used to verify JWT signatures issued by Authexa.

---

## Category 2: Admin API Endpoints

These endpoints manage the Authexa provider itself and are protected by an admin user's session cookie.
- **Authentication**: Session Cookie (`Cookie: session_id=...`).
- **CSRF Protection**: All `POST`, `PUT`, `DELETE` requests require an `X-CSRF-Token` header.

### Endpoints overview:
- **`GET /api/admin/stats`**: Returns overall system statistics (counts of users, clients, active tokens).
- **`GET /api/admin/audit-logs`**: Returns recent audit log events (logins, client creation, etc.).

### Client Management (`/api/admin/clients`)
- **`GET /api/admin/clients`**: List all clients.
- **`POST /api/admin/clients`**: Create a new client (Returns the generated `client_secret` exactly once).
- **`GET /api/admin/clients/{clientID}`**: Get a specific client.
- **`PUT /api/admin/clients/{clientID}`**: Update a client.
- **`DELETE /api/admin/clients/{clientID}`**: Delete a client.

### User Management (`/api/admin/users`)
- **`GET /api/admin/users`**: List all users.
- **`POST /api/admin/users`**: Create a new user.
- **`GET /api/admin/users/{userID}`**: Get a specific user.
- **`PUT /api/admin/users/{userID}`**: Update a user.
- **`DELETE /api/admin/users/{userID}`**: Delete a user.