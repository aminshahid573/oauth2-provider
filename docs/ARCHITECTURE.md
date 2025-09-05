# Project Architecture

This document provides an overview of the architectural decisions and patterns used in the OAuth2 Provider project. The primary goal is to create a system that is maintainable, testable, and easy to reason about.

## Core Philosophy: Clean Architecture

The project follows the principles of **Clean Architecture**, which emphasizes a clear separation of concerns between different layers of the application.

-   **Dependencies flow inwards:** The HTTP layer depends on services, and services depend on storage interfaces, but never the other way around.
-   **Business logic is independent:** The core business rules (in the `services` package) have no knowledge of the database or the web framework.
-   **Interfaces are key:** We define contracts (`storage` interfaces) that decouple the business logic from the data storage implementation (MongoDB).

## Project Structure Overview

```
oauth2-provider/
├── cmd/                # Main application entry points
├── internal/           # All private application code
│   ├── config/         # Configuration loading and validation
│   ├── handlers/       # HTTP handlers (the "Controller" layer)
│   ├── middleware/     # HTTP middleware (logging, auth, CORS, etc.)
│   ├── models/         # Core data structures (User, Client, etc.)
│   ├── services/       # Business logic and orchestration
│   ├── storage/        # Data access layer (interfaces and implementations)
│   └── utils/          # Shared utilities (crypto, templates, errors)
├── web/                # Embedded frontend assets (templates, CSS, JS)
├── docs/               # Project documentation
└── ...                 # Docker, scripts, and other config files
```

### The Layers Explained

#### 1. `handlers` (The HTTP Layer)

-   **Responsibility**: To handle incoming HTTP requests and produce HTTP responses.
-   **Details**: This layer is responsible for parsing request bodies, query parameters, and headers. It validates input and calls the appropriate service methods. It should contain minimal business logic, acting primarily as a translator between the HTTP world and the service layer. It is also responsible for calling the central error handler.

#### 2. `services` (The Business Logic Layer)

-   **Responsibility**: To orchestrate the core business rules of the application.
-   **Details**: This is where the main logic of the OAuth2 flows resides. For example, the `TokenService` knows how to generate and store different types of tokens, and the `AuthService` knows how to validate a user's password. Services depend on `storage` interfaces, not concrete database types.

#### 3. `storage` (The Data Access Layer)

-   **Responsibility**: To define how the application interacts with data persistence.
-   **`interfaces.go`**: This is the most important file in the package. It defines the contracts (e.g., `UserStore`, `ClientStore`) that the service layer will use.
-   **`mongodb/` & `redis/`**: These directories contain the **concrete implementations** of the storage interfaces. If we wanted to switch from MongoDB to PostgreSQL, we would only need to create a new `postgres/` implementation without changing any of the service-layer code.

### Dependency Injection

The application is wired together in `cmd/server/main.go`. This function acts as the **Dependency Injection (DI) Container**. It follows these steps on startup:

1.  Loads the configuration.
2.  Establishes database and cache connections.
3.  Initializes the storage repositories (e.g., `mongodb.NewUserRepository`).
4.  Initializes the services, "injecting" the storage repositories they depend on (e.g., `services.NewAuthService(userStore)`).
5.  Initializes the handlers, injecting the services they depend on.
6.  Builds the router and middleware chain, injecting the handlers.
7.  Starts the HTTP server.

This pattern ensures that components are loosely coupled and can be easily tested by injecting mock dependencies.
