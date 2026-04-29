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
- Promethus metrics endpoint for monitoring.
- Docker Compose setup for easy local development and deployment.
- Initial project documentation (API, Architecture, Deployment, Flows).
- Open Source community files (Code of Conduct, Contributing guide, Security policy).

### Fixed
- Layout rendering bug causing 500 errors in the device authorization consent flow.
