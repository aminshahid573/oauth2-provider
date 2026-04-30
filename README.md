<p align="center">
  <img src="https://readme-typing-svg.herokuapp.com?font=Fira+Code&size=22&pause=1000&color=00F7FF&center=true&vCenter=true&width=500&lines=Hi%2C+I'm+Shahid+Amin+👋;OAuth2+Provider+in+Go+🚀;Secure.+Scalable.+Production+Ready." alt="Typing SVG" />
</p>



<h1 align="center">🔐 Authexa in Go</h1>

<p align="center">
  <b>A blazing fast, production-ready OAuth2 Authorization Server built with 💙 Go, 🛢 MongoDB, and ⚡ Redis.</b>
</p>

<p align="center">
  <a href="https://golang.org/"><img alt="Go Version" src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white"></a>
  <a href="#"><img alt="License" src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square"></a>
  <a href="#"><img alt="Platform" src="https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20Mac-blue?style=flat-square"></a>
  <a href="#"><img alt="Maintainer" src="https://img.shields.io/badge/Maintained%20by-Shahid%20Amin-blueviolet?style=flat-square"></a>
</p>

---
## Why This Project?

Authentication is the front door of every application. Most developers 
outsource it to Auth0, Okta, or Firebase — and pay thousands of dollars 
a month as they scale, while handing over their most sensitive user data 
to a third party.

This project gives you a **self-hosted alternative**. You own the server, 
you own the data, you own the keys.

Built from scratch against official IETF specifications — not wrapped 
around a library — this is a full OAuth2 Authorization Server you can 
understand, audit, and extend yourself.

**Built for:**
- Startups who want auth infrastructure without the Auth0 bill
- Teams with compliance requirements (HIPAA, GDPR) who can't 
     send user data to third parties  
- Developers who want to understand OAuth2 by reading 
     real working code
- Engineers who need a lightweight alternative to Keycloak


## What Is This?

This is a complete OAuth2 and OpenID Connect Authorization Server — 
the same kind of system that powers "Login with Google" — except you 
run it yourself, on your own infrastructure.

It handles login, consent, token issuance, and session management for 
all your applications from a single, central auth server.

### Supported Features

-   ✅ **Authorization Code Flow** (with PKCE)
-   ✅ **Client Credentials Flow**
-   ✅ **Refresh Token Flow**
-   ✅ **Device Authorization Flow**
-   ✅ **JWT Bearer Token Flow**
-   ✅ **Token Introspection & Revocation**
-   ✅ **JWKS & Discovery Endpoints**
-   ✅ **Admin Dashboard & API**
-   ✅ **Metrics & Health Checks**

```mermaid
sequenceDiagram
    actor 👤 User
    participant 📱 Client App
    participant 🌐 Browser
    participant 🔐 OAuth Server
    participant 💾 Database

    %% Phase 1: Initiate Login
    👤 User->>📱 Client App: Click "Login"
    📱 Client App->>🌐 Browser: Redirect to /authorize + PKCE challenge
    
    %% Phase 2: User Authentication
    🌐 Browser->>🔐 OAuth Server: GET /authorize
    🔐 OAuth Server->>🌐 Browser: Show login page
    👤 User->>🌐 Browser: Enter credentials
    🌐 Browser->>🔐 OAuth Server: POST /login
    🔐 OAuth Server->>💾 Database: Validate user
    💾 Database-->>🔐 OAuth Server: User valid
    🔐 OAuth Server->>🌐 Browser: Set session + show consent
    
    %% Phase 3: User Consent
    👤 User->>🌐 Browser: Click "Allow"
    🌐 Browser->>🔐 OAuth Server: POST /authorize (consent)
    🔐 OAuth Server->>💾 Database: Store auth code + PKCE
    🔐 OAuth Server->>🌐 Browser: Redirect with auth code
    
    %% Phase 4: Token Exchange
    🌐 Browser->>📱 Client App: Deliver auth code
    📱 Client App->>🔐 OAuth Server: POST /token (code + PKCE verifier)
    🔐 OAuth Server->>💾 Database: Validate code + PKCE
    💾 Database-->>🔐 OAuth Server: Valid
    🔐 OAuth Server->>💾 Database: Store refresh token
    🔐 OAuth Server-->>📱 Client App: Return access + refresh tokens
    📱 Client App-->>👤 User: Login successful
```
For a deep dive into the project's design, please see the **[Architecture Documentation](./docs/ARCHITECTURE.md)**.

## Quick Start (Docker Compose)

This is the fastest and recommended way to get the entire stack running.

### Prerequisites

-   [Docker](https://www.docker.com/get-started) & [Docker Compose](https://docs.docker.com/compose/install/)
-   [Go](https://go.dev/doc/install) (1.22+)
-   `openssl` (for generating secrets)

### 1. Configure Environment

Copy the example environment file. The default values are configured to work with Docker Compose out of the box.
```bash
cp .env.example .env
```
**Important**: For a real deployment, you must generate your own secrets in the `.env` file. See the [Setup Guide](./docs/SETUP.md) for details.

### 2. Build and Run the Stack

This single command will build the Go application, start the databases, and run the server.
```bash
docker-compose -f docker/docker-compose.yml up --build -d
```

### 3. Verify and Access

-   **Check container status:**
    ```bash
    docker-compose -f docker/docker-compose.yml ps
    ```
-   **Authexa is running at:** `http://localhost:8080`
-   **Mongo Express (DB Admin) is at:** `http://localhost:8081`

### 4. Seed Initial Data

The database is currently empty. To log in and test the flows, you need to create an admin user and a test client.

For a complete script to do this, please follow the **[Database Seeding section in the Setup Guide](./docs/SETUP.md#5-database-seeding)**.

## Documentation

This project includes comprehensive documentation for developers, administrators, and API consumers.

-   **[SETUP.md](./docs/SETUP.md)**: A detailed guide for setting up a local development environment.
-   **[ARCHITECTURE.md](./docs/ARCHITECTURE.md)**: An explanation of the project's structure and design patterns.
-   **[API.md](./docs/API.md)**: A complete technical reference for every API endpoint.
-   **[FLOWS.md](./docs/FLOWS.md)**: Practical walkthroughs of each supported OAuth2 flow.
-   **[DEPLOYMENT.md](./docs/DEPLOYMENT.md)**: Instructions for deploying the application using Docker.
-   **[REFERENCES.md](./REFERENCES.md)**: A complete list of all IETF RFCs and official specifications this project is implemented against.

## Technology Stack

-   **Backend**: Go 1.22+ (`net/http`)
-   **Database**: MongoDB
-   **Cache/Sessions**: Redis
-   **API Endpoints**: Standard Library `net/http` with `ServeMux`
-   **Frontend Templates**: Go `html/template`
-   **Observability**: Prometheus Metrics
