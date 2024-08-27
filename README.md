# AliceTraINT

CERN ALICE Training Interface.

AliceTraINT is a full-stack web application built using Go (Golang) with PostgreSQL, GORM, and HTMX. This application connects to CERN SSO for authentication and uses a PostgreSQL database to manage user data. Its purpose is to be training interface for CERN ALICE experiment machine learning projects especially PIDML.

## Project Structure

- **cmd/AliceTraINT/**: Contains the main application entry point.
- **internal/auth/**: Handles authentication and session management using CERN SSO.
- **internal/db/migrate/migrations/**: Database migrations.
- **internal/db/models/**: Defines the database models.
- **internal/db/repository/**: Provides database access methods.
- **internal/handler/**: Contains HTTP handlers for managing user requests.
- **web/templates/**: HTML templates used for rendering views.
- **test/**: Contains integration tests.
- **Dockerfile**: Dockerfile for containerizing the application.
- **docker-compose.yml**: Docker Compose configuration for running the application and PostgreSQL in containers.

## Prerequisites

- Go 1.22 or higher
- Docker
- Docker Compose
- PostgreSQL

## Setup

### Using Docker and Docker Compose

1. **Build and Run Docker Containers**

   Ensure Docker and Docker Compose are installed. Build and start the containers:

   ```bash
   docker-compose up --build
   ```
   or
   ```bash
   make docker
   ```

   This command builds the Docker images and starts the containers for the application and PostgreSQL.

2. **Access the Application**

   Once the containers are up, access the application at `http://localhost:8080`.

### Local Development

1. **Clone the Repository**

   ```bash
   git clone https://github.com/mytkom/AliceTraINT.git
   cd AliceTraINT
   ```

2. **Create a `.env` File**

   Create a `.env` file in the root directory of the project with the following content:

   ```dotenv
   CERN_REALM_URL=https://example.cern.ch/auth/realms/your-realm
   CERN_CLIENT_ID=your-client-id
   CERN_CLIENT_SECRET=your-client-secret
   CERN_REDIRECT_URL=http://localhost:8080/callback
   ```

   Replace the placeholder values with your actual CERN SSO details.

3. **Install Dependencies**

   Ensure you have Go 1.22 or higher installed. Install Go dependencies:

   ```bash
   go mod tidy
   ```

4. **Run Migrations**

   Run the database migrations to set up the schema:

   ```bash
   migrate -path db/migrations -database "postgres://user:password@localhost:5432/alice-train?sslmode=disable" up
   ```

5. **Start the Application**

   Run the application locally:

   ```bash
   go run cmd/AliceTraINT/main.go
   ```
   or
   ```bash
   make run
   ```

   Access the application at `http://localhost:8088`.

### Nix Development Environment

1. **Setup Nix Environment**

   Ensure Nix is installed. Enter the development shell:

   ```bash
   nix develop
   ```

   This command sets up the development environment as defined in the Nix flake.

## Makefile

The project includes a `Makefile` to simplify common development tasks. Below are some of the available commands:

- **`make build`**: Build the application binary.
- **`make run`**: Run the application locally.
- **`make test`**: Run unit and integration tests.
- **`make lint`**: Run linters.
- **`make docker`**: Docker compose up.

### Available Linters

- **`golangci-lint`**: A Go linter aggregator that runs multiple linters in parallel to check for issues in Go code. It combines several linters including `golint`, `govet`, `errcheck`, `staticcheck`, and more.

### Setup and Usage

1. **Install `golangci-lint`**

   You can install `golangci-lint` using the following command:

   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

2. **Run Linters**

   To run all configured linters on your codebase, execute:

   ```bash
   golangci-lint run
   ```
    or
   ```bash
   make lint
   ```

## Configuration

The application configuration is managed through environment variables defined in the `.env` file.

### Required Environment Variables

- `CERN_REALM_URL`: The URL for CERN's OIDC provider.
- `CERN_CLIENT_ID`: Client ID for the CERN OIDC application.
- `CERN_CLIENT_SECRET`: Client secret for the CERN OIDC application.
- `CERN_REDIRECT_URL`: Redirect URL for the CERN OIDC application.

## Testing

1. **Run All Tests**

   ```bash
   go test ./...
   ```
   or
   ```bash
   make test
   ```

2. **Run Integration Tests**

   ```bash
   go test -tags=integration ./test/integration
   ```

## Dockerfile

The `Dockerfile` defines the container setup for the application. It uses a multi-stage build to create a lightweight production image.

## Docker Compose

The `docker-compose.yml` file defines the services for the application and PostgreSQL, allowing you to easily start both services together.

## Nix Flake

The `flake.nix` file provides a NixOS configuration for a reproducible development environment. It defines the necessary packages and configurations.

## Contribution

Contributions are welcome! Please submit issues and pull requests via GitHub.

## License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0). See the [LICENSE](LICENSE) file for details.

## Acknowledgements

- [GORM](https://gorm.io) - ORM for Go
- [HTMX](https://htmx.org) - HTML extensions for enhancing user experience
- [CERN SSO](https://cern.ch) - Authentication via CERN's SSO
- [Docker](https://docker.com) - Containerization
- [Nix](https://nixos.org) - Package management and build tool


