<p align="center">
  <img src="https://readme-typing-svg.demolab.com?font=Fira+Code&weight=500&size=22&pause=1000&color=F7F7F7&background=1E1E1E00&center=true&vCenter=true&width=500&lines=Hi%2C+I'm+Shahid+Amin+👋;OAuth2+Provider+in+Go+%F0%9F%9A%80;Secure.+Scalable.+Production+Ready.">
</p>

<h1 align="center">🔐 OAuth2 Provider in Go</h1>

<p align="center">
  <b>A blazing fast, production-ready OAuth2 Authorization Server built with 💙 Go, 🛢 MongoDB, and ⚡ Redis.</b>
</p>

<p align="center">
  <a href="https://golang.org/"><img alt="Go Version" src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white"></a>
  <a href="#"><img alt="License" src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square"></a>
  <a href="#"><img alt="Platform" src="https://img.shields.io/badge/Platform-Linux%20%7C%20Windows-blue?style=flat-square"></a>
  <a href="#"><img alt="Maintainer" src="https://img.shields.io/badge/Maintained%20by-Shahid%20Amin-blueviolet?style=flat-square"></a>
</p>

---
## Table of Contents

- [Project Overview](#project-overview)
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation & Setup](#installation--setup)
- [Running the Server](#running-the-server)
- [API Endpoints](#api-endpoints)
  - [Authorization Endpoint](#1-authorization-endpoint-oauth2authorize)
  - [Token Endpoint](#2-token-endpoint-oauth2token)
  - [Token Introspection Endpoint](#3-token-introspection-endpoint-oauth2introspect)
- [Testing with `curl`](#testing-with-curl)
  - [Flow 1: Authorization Code Flow](#flow-1-authorization-code-flow)
  - [Flow 2: Client Credentials Flow](#flow-2-client-credentials-flow)
  - [Flow 3: Refresh Token Flow](#flow-3-refresh-token-flow)
  - [Flow 4: Device Authorization Flow](#flow-4-device-authorization-flow)
  - [Flow 5: JWT Bearer Token Flow](#flow-5-jwt-bearer-token-flow)
  - [Endpoint Test: Token Introspection](#endpoint-test-token-introspection)
- [Project Structure](#project-structure)
- [Technology Stack](#technology-stack)

## Project Overview

This project implements a full-featured OAuth2 server that can act as a central authentication authority for your applications. It handles user login, consent management, and the issuance of secure access tokens.

## Features

- **OAuth2 Standards Compliant**: Implements core RFCs for modern authentication.
- **Secure by Default**: Includes CSRF protection, password hashing (bcrypt), and secure token generation.
- **Scalable Backend**: Uses MongoDB for persistent storage and Redis for caching and session management.
- **Clean Architecture**: A well-defined separation of concerns between handlers, services, and storage layers.
- **User-Facing Frontend**: Includes built-in pages for user login and application consent.

### Supported OAuth2 Flows & Features

- [x] **Authorization Code Flow**: For user-based authorization in web and mobile apps.
- [x] **Client Credentials Flow**: For machine-to-machine (M2M) authentication.
- [x] **Refresh Token Flow**: To obtain new access tokens without user interaction.
- [x] **Device Authorization Flow**: For input-constrained devices like Smart TVs.
- [x] **JWT Bearer Token Flow**: For high-security M2M authentication using public/private keys.
- [x] **Token Introspection**: Allows resource servers to validate access tokens.

## Getting Started

### Prerequisites

- **Go**: Version 1.22+
- **Docker** & **Docker Compose**: To run MongoDB and Redis.
- **`make`** (Optional): For convenience commands.
- **`mongosh`** (Optional): For direct database interaction.
- **`curl`** and **`jq`**: For testing API endpoints.

### Installation & Setup

1.  **Clone the repository:**
    ```bash
    git clone <not-published-yet>
    cd oauth2-provider
    ```

2.  **Configure Environment Variables:**
    ```bash
    cp .env.example .env
    ```
    Open `.env` and set `JWT_SECRET_KEY` and `CSRF_AUTH_KEY` to strong, random values.

3.  **Install Go Dependencies:**
    ```bash
    go mod tidy
    ```

4.  **Start Backend Services:**
    ```bash
    docker-compose -f docker/docker-compose.yml up -d
    ```

5.  **Seed the Database:**
    Connect to MongoDB (`mongosh "mongodb://root:password@localhost:27017"`) and run a script to create a test user and several clients. **You must generate your own unique bcrypt hashes for the secrets.**

    *(See the [Testing Guide](#testing-with-curl) for the full seed script and instructions.)*

## Running the Server

```bash
go run ./cmd/server
```
The server will be available at `http://localhost:8080`.

## API Endpoints

### 1. Authorization Endpoint: `/oauth2/authorize`

- **Method**: `GET`, `POST`
- **Description**: The user-facing starting point for the Authorization Code Flow. Handles user login and consent.

### 2. Token Endpoint: `/oauth2/token`

- **Method**: `POST`
- **Description**: The central backend endpoint for clients to exchange codes or credentials for tokens. Supports `authorization_code`, `client_credentials`, `refresh_token`, `urn:ietf:params:oauth:grant-type:device_code`, and `urn:ietf:params:oauth:grant-type:jwt-bearer` grant types.

### 3. Token Introspection Endpoint: `/oauth2/introspect`

- **Method**: `POST`
- **Authentication**: HTTP Basic Auth (using the resource server's `client_id` and `client_secret`).
- **Description**: Allows a client (acting as a resource server) to check the validity of an access token.

#### Request Parameters (Form Body)

| Parameter | Required | Description              |
| --------- | -------- | ------------------------ |
| `token`   | Yes      | The access token to check. |

#### Success Response (`active: true`)

```json
{
  "active": true,
  "scope": "openid profile",
  "client_id": "test-client",
  "sub": "6675d3a...",
  "exp": 1755855000,
  "iat": 1755854100,
  "nbf": 1755854100,
  "iss": "oauth2-provider",
  "aud": ["test-client"],
  "jti": "...",
  "token_type": "Bearer"
}
```

#### Failure Response (`active: false`)

```json
{
  "active": false
}
```

## Testing with `curl`

*(This section assumes you have followed the database seeding instructions in the [Getting Started](#getting-started) guide.)*

### Flow 1: Authorization Code Flow

1.  **Start in browser:** `http://localhost:8080/oauth2/authorize?response_type=code&client_id=test-client&...`
2.  Log in, grant consent, and copy the `code` from the redirect URL.
3.  **Exchange code for tokens:**
    ```bash
    curl -X POST http://localhost:8080/oauth2/token \
    -d "grant_type=authorization_code" \
    -d "code=YOUR_CODE_HERE" \
    -d "client_id=test-client" \
    -d "client_secret=test-secret" | jq
    ```
    **Expected:** `access_token` and `refresh_token`.

### Flow 2: Client Credentials Flow

```bash
curl -X POST http://localhost:8080/oauth2/token \
-d "grant_type=client_credentials" \
-d "client_id=m2m-client" \
-d "client_secret=m2m-secret" | jq
```
**Expected:** `access_token`.

### Flow 3: Refresh Token Flow

Use the `refresh_token` from Flow 1.
```bash
curl -X POST http://localhost:8080/oauth2/token \
-d "grant_type=refresh_token" \
-d "refresh_token=YOUR_REFRESH_TOKEN_HERE" \
-d "client_id=test-client" \
-d "client_secret=test-secret" | jq
```
**Expected:** A new `access_token`.

### Flow 4: Device Authorization Flow

1.  **Get codes:**
    ```bash
    curl -X POST http://localhost:8080/oauth2/device_authorization \
    -d "client_id=test-client" | jq
    ```
    Copy the `user_code` and `device_code`.
2.  **In browser:** Go to `http://localhost:8080/device`, log in, enter the `user_code`, and grant consent.
3.  **Poll for tokens:**
    ```bash
    curl -X POST http://localhost:8080/oauth2/token \
    -d "grant_type=urn:ietf:params:oauth:grant-type:device_code" \
    -d "device_code=YOUR_DEVICE_CODE_HERE" \
    -d "client_id=test-client" | jq
    ```
    **Expected:** `access_token` and `refresh_token`.

### Flow 5: JWT Bearer Token Flow

Requires running the separate `jwt-client-simulator` tool. Follow the instructions printed by that tool to seed the client and run the `curl` command.
**Expected:** `access_token`.

### Endpoint Test: Token Introspection

1.  **Seed a Resource Server Client:** Create a client in MongoDB to represent your API.
    ```javascript
    // In mongosh
    db.clients.insertOne({
        client_id: "resource-server",
        client_secret: "PASTE_HASH_FOR_resource-server-secret_HERE",
        name: "My Protected API"
    });
    ```
2.  **Get any valid `access_token`** from one of the flows above.
3.  **Call the endpoint:**
    ```bash
    curl -X POST http://localhost:8080/oauth2/introspect \
    -u "resource-server:resource-server-secret" \
    -d "token=YOUR_ACCESS_TOKEN_HERE" | jq
    ```
    **Expected:** A JSON response with `active: true` and token details.

## Project Structure (not updated)

```
oauth2-provider/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── handlers/
│   │   ├── auth.go                # Authorization endpoints
│   │   ├── token.go               # Token endpoints
│   │   ├── introspection.go       # Token introspection
│   │   ├── revocation.go          # Token revocation
│   │   ├── jwks.go                # JWKS endpoint
│   │   ├── admin.go               # Admin panel API
│   │   └── frontend.go            # Frontend page handlers
│   ├── middleware/
│   │   ├── auth.go                # Authentication middleware
│   │   ├── cors.go                # CORS middleware
│   │   ├── logging.go             # Request logging
│   │   ├── ratelimit.go           # Rate limiting
│   │   ├── security.go            # Security headers
│   │   └── session.go             # Session management
│   ├── models/
│   │   ├── client.go              # OAuth2 client model
│   │   ├── user.go                # User model
│   │   ├── token.go               # Token models
│   │   ├── scope.go               # Scope definitions
│   │   └── session.go             # Session model
│   ├── services/
│   │   ├── auth.go                # Authentication service
│   │   ├── token.go               # Token management
│   │   ├── client.go              # Client management
│   │   ├── user.go                # User management
│   │   ├── pkce.go                # PKCE implementation
│   │   └── session.go             # Session management
│   ├── storage/
│   │   ├── interfaces.go          # Storage interfaces
│   │   ├── mongodb/               # MongoDB implementation
│   │   │   ├── client.go
│   │   │   ├── user.go
│   │   │   ├── token.go
│   │   │   └── connection.go
│   │   ├── redis/                 # Redis implementation
│   │   │   ├── session.go
│   │   │   ├── token.go
│   │   │   └── connection.go
│   │   └── memory/                # In-memory implementation (dev)
│   └── utils/
│       ├── crypto.go              # Cryptographic utilities
│       ├── jwt.go                 # JWT utilities
│       ├── validation.go          # Input validation
│       ├── errors.go              # Error handling
│       └── template.go            # Template utilities
├── pkg/
│   └── oauth2/
│       ├── types.go               # Public types
│       └── errors.go              # Public error types
├── web/
│   ├── static/                    # Static assets (CSS, JS, images)
│   │   ├── css/
│   │   │   ├── auth.css
│   │   │   ├── admin.css
│   │   │   └── common.css
│   │   ├── js/
│   │   │   ├── auth.js
│   │   │   ├── admin.js
│   │   │   └── common.js
│   │   └── images/
│   └── templates/                 # HTML templates
│       ├── layouts/
│       │   ├── base.html
│       │   ├── auth.html
│       │   └── admin.html
│       ├── auth/
│       │   ├── login.html         # User login page
│       │   ├── consent.html       # OAuth2 consent page
│       │   ├── device.html        # Device authorization page
│       │   └── error.html         # Error pages
│       └── admin/
│           ├── dashboard.html     # Admin dashboard
│           ├── clients.html       # Client management
│           ├── users.html         # User management
│           ├── tokens.html        # Token inspection
│           └── logs.html          # Audit logs
├── api/
│   └── openapi.yaml               # OpenAPI specification
├── migrations/
│   └── mongodb/
│       ├── init.js                # MongoDB initialization scripts
│       └── indexes.js             # Index creation
├── docker/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── docker-compose.dev.yml
├── scripts/
│   ├── setup.sh
│   ├── dev.sh
│   └── build.sh
├── docs/
│   ├── README.md
│   ├── DEPLOYMENT.md
│   ├── API.md
│   └── FRONTEND.md
├── go.mod
├── go.sum
├── .env.example
├── .gitignore
└── Makefile
```


## Technology Stack

- **Backend**: Go 1.22+ (`net/http`), `gorilla/csrf`
- **Database**: MongoDB
- **Cache/Sessions**: Redis
- **Configuration**: `spf13/viper`
- **JWTs**: `golang-jwt/jwt/v5`, `lestrrat-go/jwx/v2`
- **Password Hashing**: `golang.org/x/crypto/bcrypt`
