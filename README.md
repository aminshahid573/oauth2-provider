<p align="center">
  <img src="https://readme-typing-svg.herokuapp.com?font=Fira+Code&size=22&pause=1000&color=00F7FF&center=true&vCenter=true&width=500&lines=Hi%2C+I'm+Shahid+Amin+👋;OAuth2+Provider+in+Go+🚀;Secure.+Scalable.+Production+Ready." alt="Typing SVG" />
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
## Project Overview

This project is a complete, standards-compliant OAuth2 and OpenID Connect provider written in Go. It is designed to be a secure, performant, and scalable central authentication authority for your entire application ecosystem.

It includes a full suite of OAuth2 flows, a user-facing frontend for login and consent, and a complete admin panel for managing clients and users.

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
-   **OAuth2 Provider is running at:** `http://localhost:8080`
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

## Technology Stack

-   **Backend**: Go 1.22+ (`net/http`)
-   **Database**: MongoDB
-   **Cache/Sessions**: Redis
-   **API Endpoints**: Standard Library `net/http` with `ServeMux`
-   **Frontend Templates**: Go `html/template`
-   **Observability**: Prometheus Metrics
