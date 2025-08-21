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
- [Testing with `curl`](#testing-with-curl)
  - [Flow 1: Authorization Code Flow](#flow-1-authorization-code-flow)
  - [Flow 2: Client Credentials Flow](#flow-2-client-credentials-flow)
  - [Flow 3: Refresh Token Flow](#flow-3-refresh-token-flow)
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

### Supported OAuth2 Flows (Implemented so far)

- [x] **Authorization Code Flow**: The standard flow for user-based authorization in web and mobile apps.
- [x] **Client Credentials Flow**: For machine-to-machine (M2M) authentication.
- [x] **Refresh Token Flow**: Allows clients to obtain new access tokens without user interaction.

## Getting Started

### Prerequisites

- **Go**: Version 1.22+
- **Docker** & **Docker Compose**: To run the required MongoDB and Redis services.
- **`make`** (Optional, for convenience): On Windows, you can install `make` via Chocolatey (`choco install make`) or Winget.
- **`mongosh`** (Optional): For direct interaction with the MongoDB database.
- **`curl`** and **`jq`**: For testing the API endpoints from the command line.

### Installation & Setup

1.  **Clone the repository:**
    ```bash
    git clone <not-publishe-yet>
    cd oauth2-provider
    ```

2.  **Configure Environment Variables:**
    Copy the example environment file and customize it.
    ```bash
    cp .env.example .env
    ```
    Open the `.env` file and ensure the `JWT_SECRET_KEY` and `CSRF_AUTH_KEY` are set to strong, randomly generated 32-byte strings.

3.  **Install Go Dependencies:**
    ```bash
    go mod tidy
    ```

4.  **Start Backend Services:**
    Use Docker Compose to start MongoDB and Redis in the background.
    ```bash
    docker-compose -f docker/docker-compose.yml up -d
    ```
    *(Or `make docker-up` if you have `make` installed)*

5.  **Seed the Database:**
    You need to create at least one user and two clients to test the different flows. Connect to MongoDB using `mongosh` or a GUI like MongoDB Compass.

    ```bash
    # Connect to the database
    mongosh "mongodb://root:password@localhost:27017"
    ```

    Inside the `mongosh` shell, run the following commands. **Remember to generate your own bcrypt hashes for the secrets!**

    ```javascript
    use oauth2_provider;

    // Create a test user (password: password123)
    // Generate hash with: bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
    db.users.insertOne({
        username: "testuser",
        hashed_password: "$2a$10$your/generated/bcrypt/hash/for/password123",
        created_at: new Date(),
        updated_at: new Date()
    });

    // Create a client for the Authorization Code Flow
    // Generate hash for "test-secret"
    db.clients.insertOne({
        client_id: "test-client",
        client_secret: "$2a$10$your/generated/bcrypt/hash/for/test-secret",
        name: "My Awesome App",
        redirect_uris: ["http://localhost:3000/callback"],
        grant_types: ["authorization_code", "refresh_token"],
        response_types: ["code"],
        scopes: ["openid", "profile"],
        created_at: new Date(),
        updated_at: new Date()
    });

    // Create a client for the Client Credentials Flow
    // Generate hash for "m2m-secret"
    db.clients.insertOne({
        client_id: "m2m-client",
        client_secret: "$2a$10$your/generated/bcrypt/hash/for/m2m-secret",
        name: "Internal API Service",
        redirect_uris: [],
        grant_types: ["client_credentials"],
        response_types: [],
        scopes: ["api:read", "api:write"],
        created_at: new Date(),
        updated_at: new Date()
    });
    ```

## Running the Server

Once the setup is complete, you can run the application server:

```bash
go run ./cmd/server
```

The server will be available at `http://localhost:8080`

## API Endpoints

### 1. Authorization Endpoint: `/oauth2/authorize`

This endpoint is the starting point for user-centric flows. It is browser-based and requires a logged-in user.

- **Method**: `GET`, `POST`
- **Authentication**: User session cookie (redirects to `/login` if not present).
- **Description**:
  - `GET`: Validates the client and scopes, then displays the consent page to the user.
  - `POST`: Handles the user's "Allow" or "Deny" decision from the consent form. If allowed, it generates an authorization code and redirects to the client's `redirect_uri`.

#### `GET` Request Parameters (Query String)

| Parameter      | Required | Description                                          |
| -------------- | -------- | ---------------------------------------------------- |
| `response_type`| Yes      | Must be `code`.                                      |
| `client_id`    | Yes      | The ID of the client application.                    |
| `redirect_uri` | Yes      | The URL to redirect to after authorization.          |
| `scope`        | No       | A space-delimited list of requested scopes.          |
| `state`        | Yes      | An opaque value used to prevent CSRF attacks.        |

### 2. Token Endpoint: `/oauth2/token`

This is a backend API endpoint used by client applications to exchange codes or credentials for tokens.

- **Method**: `POST`
- **Content-Type**: `application/x-www-form-urlencoded`
- **Description**: Issues `access_token` and optionally `refresh_token` based on the `grant_type`.

#### `POST` Request Parameters (Form Body)

| Parameter       | Required | Description                                                              |
| --------------- | -------- | ------------------------------------------------------------------------ |
| `grant_type`    | Yes      | The type of grant being requested (`authorization_code`, `client_credentials`, `refresh_token`). |
| `code`          | If `authorization_code` | The authorization code received from the `/authorize` endpoint. |
| `redirect_uri`  | If `authorization_code` | Must match the `redirect_uri` from the authorization request. |
| `refresh_token` | If `refresh_token` | The refresh token used to obtain a new access token. |
| `client_id`     | Yes      | The client's ID.                                                         |
| `client_secret` | Yes      | The client's secret.                                                     |
| `scope`         | No       | For `client_credentials`, a space-delimited list of scopes.              |

## Testing with `curl`

### Flow 1: Authorization Code Flow

1.  **Start the flow in your browser:**
    ```
    http://localhost:8080/oauth2/authorize?response_type=code&client_id=test-client&redirect_uri=http://localhost:3000/callback&scope=openid%20profile&state=xyz123
    ```
2.  Log in as `testuser` / `password123`.
3.  Click "Allow" on the consent page.
4.  You will be redirected to a URL like `http://localhost:3000/callback?code=YOUR_CODE_HERE&state=xyz123`. Copy `YOUR_CODE_HERE`.

5.  **Exchange the code for a token:**
    ```bash
    curl -X POST http://localhost:8080/oauth2/token \
    -d "grant_type=authorization_code" \
    -d "code=YOUR_CODE_HERE" \
    -d "redirect_uri=http://localhost:3000/callback" \
    -d "client_id=test-client" \
    -d "client_secret=test-secret" | jq
    ```
    You will receive an `access_token` and a `refresh_token`.

### Flow 2: Client Credentials Flow

```bash
curl -X POST http://localhost:8080/oauth2/token \
-d "grant_type=client_credentials" \
-d "client_id=m2m-client" \
-d "client_secret=m2m-secret" \
-d "scope=api:read" | jq
```
You will receive an `access_token` valid for the requested scope.

### Flow 3: Refresh Token Flow

Use the `refresh_token` you received from the Authorization Code Flow.

```bash
curl -X POST http://localhost:8080/oauth2/token \
-d "grant_type=refresh_token" \
-d "refresh_token=YOUR_REFRESH_TOKEN_HERE" \
-d "client_id=test-client" \
-d "client_secret=test-secret" | jq
```
You will receive a new `access_token`.

## Project Structure

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
- **JWTs**: `golang-jwt/jwt/v5`
- **Password Hashing**: `golang.org/x/crypto/bcrypt`
