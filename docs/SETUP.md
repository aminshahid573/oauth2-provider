# Project Setup & Configuration Guide

This document provides a detailed, step-by-step guide to setting up the OAuth2 Provider for local development.

## 1. Prerequisites

Ensure you have the following tools installed on your system:

-   **Go**: Version 1.22 or higher.
-   **Docker & Docker Compose**: To run the required database and cache services.
-   **Git**: For cloning the repository.
-   **`openssl`**: For generating secrets (usually pre-installed on Linux/macOS, available with Git Bash on Windows).
-   **`mongosh`** (Optional but Recommended): The MongoDB shell for interacting with the database.
-   **`curl` & `jq`**: For testing API endpoints from the command line.

## 2. Installation

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/aminshahid573/oauth2-provider
    cd oauth2-provider
    ```

2.  **Install Go Dependencies:**
    This command will download all the necessary libraries defined in `go.mod`.
    ```bash
    go mod tidy
    ```

## 3. Configuration (`.env` file)

The application is configured via a `.env` file in the project root.

1.  **Create the `.env` file:**
    ```bash
    cp .env.example .env
    ```

2.  **Understand and Set Environment Variables:**
    Open the `.env` file and configure the following variables.

| Variable                   | Description                                                                                                                                                           | Example Value                                                                                             |
| -------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `APP_ENV`                  | The application environment. Set to `development` to disable some security features like CSRF origin checks and secure cookies. Set to `production` for deployment.      | `development`                                                                                             |
| `SERVER_HOST`              | The host address the server binds to. `0.0.0.0` listens on all available network interfaces.                                                                            | `0.0.0.0`                                                                                                 |
| `SERVER_PORT`              | The port the server listens on.                                                                                                                                       | `8080`                                                                                                    |
| `LOG_LEVEL`                | The minimum level of logs to display. Can be `debug`, `info`, `warn`, or `error`.                                                                                       | `debug`                                                                                                   |
| `MONGO_URI`                | The connection string for MongoDB. For local development (not in Docker), use `localhost`. For Docker Compose, use the service name `mongodb`.                          | `mongodb://root:password@localhost:27017/oauth2_provider?authSource=admin`                                |
| `REDIS_ADDR`               | The address for Redis. For local development, use `localhost:6379`. For Docker Compose, use `redis:6379`.                                                               | `localhost:6379`                                                                                          |
| `JWT_SECRET_KEY`           | **(Legacy)** A secret key for HS256 signing. Still required by config validation but not used for RS256.                                                                | `your-super-secret-key-change-me`                                                                         |
| `JWT_PRIVATE_KEY_BASE64`   | **(Critical Secret)** The Base64-encoded RSA private key used for signing all access tokens (RS256). Generate with `openssl genpkey ...` and `base64 -w 0 ...`.          | `MII...` (a very long string)                                                                             |
| `CSRF_AUTH_KEY`            | **(Critical Secret)** A 32-byte random key for signing CSRF tokens. Generate with `openssl rand -base64 32`.                                                              | `your-32-byte-long-csrf-auth-key...`                                                                      |
| `BASE_URL`                 | The public-facing base URL of the server. Used for constructing redirect URIs and discovery documents.                                                                  | `http://localhost:8080`                                                                                   |
| `CORS_ALLOWED_ORIGINS`     | A comma-separated list of domains allowed to make cross-origin requests to the API.                                                                                     | `http://localhost:3000,http://localhost:8080`                                                             |

## 4. Running Dependencies

Start the required MongoDB and Redis instances using Docker Compose.

```bash
docker-compose -f docker/docker-compose.yml up -d
```

## 5. Database Seeding

For the application to function, you need to create an admin user and at least one client.

1.  **Generate Hashes:**
    Create a temporary file `hasher.go` and run `go run hasher.go` to get unique bcrypt hashes for your secrets. or just go to bcrypt generator in any browser

2.  **Connect to MongoDB:**
    ```bash
    docker exec -it oauth2-mongo mongosh -u root -p password
    ```

3.  **Run Seed Script:**
    Inside the `mongosh` shell, run the following, replacing the placeholders with your generated hashes.
    ```javascript
    use oauth2_provider;

    // Create an admin user (password: password123)
    db.users.insertOne({
        username: "testuser",
        hashed_password: "PASTE_HASH_FOR_password123_HERE",
        role: "admin",
        created_at: new Date(),
        updated_at: new Date()
    });

    // Create a client for user-based flows
    db.clients.insertOne({
        client_id: "test-client",
        client_secret: "PASTE_HASH_FOR_test-secret_HERE",
        name: "My Awesome App",
        redirect_uris: ["http://localhost:3000/callback"],
        grant_types: ["authorization_code", "refresh_token", "urn:ietf:params:oauth:grant-type:device_code"],
        response_types: ["code"],
        scopes: ["openid", "profile"],
        created_at: new Date(),
        updated_at: new Date()
    });
    ```

## 6. Running the Application

Once all setup is complete, you can run the Go server from the project root.

```bash
go run ./cmd/server
```

The application will be running and accessible at `http://localhost:8080`.

---
| [![Previous](https://img.shields.io/badge/←_Previous-1f6feb?style=for-the-badge&logo=none&logoColor=white&labelColor=1f6feb&color=1f6feb)](README.md) <br> <sub>README.md</sub> | [![Next](https://img.shields.io/badge/Next_→-1f6feb?style=for-the-badge&logo=none&logoColor=white&labelColor=1f6feb&color=1f6feb)](ARCHITECTURE.md) <br> <sub>ARCHITECTURE.md</sub> |
|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
