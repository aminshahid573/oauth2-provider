# Deployment Guide

This guide provides instructions for building and running the OAuth2 Provider using Docker and Docker Compose. This is the recommended method for both development and production environments.

## Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Configuration

The application is configured using environment variables.

1.  **Create a `.env` file:**
    Copy the example configuration file to create your own local version.
    ```bash
    cp .env.example .env
    ```

2.  **Edit `.env` for Production:**
    Before deploying, you **must** change the following values in your `.env` file:

    -   `APP_ENV`: Set to `production`. This will enable secure cookies and other production settings.
    -   `JWT_SECRET_KEY`: A randomly generated 32-byte string.
    -   `CSRF_AUTH_KEY`: A randomly generated 32-byte string.
    -   `JWT_PRIVATE_KEY_BASE64`: A base64-encoded RSA private key.
        -   Generate the key: `openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048`
        -   Encode it: `base64 -w 0 private.pem`
    -   `MONGO_ROOT_PASSWORD`: A strong password for the MongoDB root user.
    -   `CORS_ALLOWED_ORIGINS`: A comma-separated list of the domains of your frontend applications that will interact with this service (e.g., `https://my-app.com,https://another-app.com`).

3.  **Database Host Configuration:**
    When running inside Docker Compose, the application needs to connect to the other containers by their service name, not `localhost`. Update your `.env` file accordingly:
    ```dotenv
    # .env for Docker Compose
    MONGO_URI=mongodb://root:password@mongodb:27017/oauth2_provider?authSource=admin
    REDIS_ADDR=redis:6379
    ```

## Building and Running with Docker Compose

This is the simplest way to run the entire stack (Go application, MongoDB, Redis).

1.  **Build the Images and Start the Services:**
    From the root of the project directory, run:
    ```bash
    docker-compose -f docker/docker-compose.yml up --build -d
    ```
    -   `--build`: Forces Docker to rebuild the `app` image using the `Dockerfile`.
    -   `-d`: Runs the containers in detached mode (in the background).

2.  **Verify the Services are Running:**
    ```bash
    docker-compose -f docker/docker-compose.yml ps
    ```
    You should see three services (`app`, `mongodb`, `redis`) with a `STATUS` of `Up`.

3.  **Accessing the Application:**
    The OAuth2 provider will be available at `http://localhost:8080`.

## Seeding Initial Data

For a new deployment, you will need to seed the database with initial clients and an admin user.

1.  **Connect to the MongoDB Container:**
    ```bash
    docker exec -it oauth2-mongo mongosh -u root -p your_mongo_password
    ```

2.  **Run Seed Scripts:**
    Inside the `mongosh` shell, you can create your first admin user and clients. Refer to the main `README.md` for example scripts.

## Stopping the Services

To stop all running containers:
```bash
docker-compose -f docker/docker-compose.yml down