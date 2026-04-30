# Contributing to Authexa

First off, thank you for considering contributing to Authexa! It's people like you that make open source such a great community.

## Where do I go from here?

If you've noticed a bug or have a feature request, make sure to check our [Issues](../../issues) to see if someone else has already created a ticket. If not, go ahead and [make one](../../issues/new/choose)!

## Fork & create a branch

If this is something you think you can fix, then [fork Authexa](https://help.github.com/articles/fork-a-repo) and create a branch with a descriptive name.

A good branch name would be (where issue #325 is the ticket you're working on):

```sh
git checkout -b 325-add-pkce-support
```

## Local Development

### Prerequisites

- Go 1.22 or higher
- Docker & Docker Compose
- `make`

### Setting up the environment

1. Clone your fork and navigate into it:
   ```sh
   git clone https://github.com/YOUR_USERNAME/authexa.git
   cd authexa
   ```
2. Copy the example environment file:
   ```sh
   cp .env.example .env
   ```
3. Start the dependent services (MongoDB and Redis):
   ```sh
   make docker-up
   ```
4. Build and run the server locally:
   ```sh
   make run
   ```

## Running Tests

We use standard Go testing tools. Please ensure your tests pass before submitting a PR.

```sh
make test
```

## Pull Request Process

1. Ensure any install or build dependencies are removed before the end of the layer when doing a build.
2. Update the README.md or docs/ folder with details of changes to the interface, this includes new environment variables, exposed ports, useful file locations and container parameters.
3. Once you're finished with your changes, create a pull request (PR).
4. Provide a detailed description of the changes you made in the PR description. Include the issue number if applicable (e.g., `Fixes #123`).

## Code Style

- Use `go fmt` to format your code.
- We follow standard Go conventions (Effective Go).
- Add comments for any exported types and functions, as well as complex logic blocks.

Thank you for your contribution!
