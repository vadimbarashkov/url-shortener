# URL Shortener Service

A scalable URL Shortening service built with Golang, PostgreSQL, and Docker. This service allows users to shorten URLs, resolve them, retrieve statistics, and manage shortened URLs.

## Table of Contents

- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Installation](#installation)
- [Running the Application](#running-the-application)
  - [Using Docker](#using-docker)
  - [Without Docker](#without-docker)
- [API Documentation](#api-documentation)
- [Running Tests](#running-tests)
  - [Unit Tests](#unit-tests)
  - [Integration Tests](#integration-tests)
  - [E2E Tests](#e2e-tests)
- [Database Migrations](#database-migrations)
- [Application Configuration](#application-configuration)
- [Contributing](#contributing)
- [License](#license)

## Tech Stack

- Golang
- PostgreSQL
- Docker

## Project Structure

Here are the main components of the application:

```bash
.
├── cmd                     # Application entrypoints
│   └── url-shortener
├── docs                    # API documentation
├── internal
│   ├── adapter
│   │   ├── delivery        # Data delivery layer
│   │   │   └── http
│   │   └── repository      # Database repositories
│   │       └── postgres
│   ├── app                 # Application initialization logic
│   ├── config              # Configuration loading logic
│   ├── entity              # Core domain entities
│   └── usecase             # Business logic
├── migrations
├── mocks
│   ├── http
│   └── usecase
├── pkg
│   └── postgres            # PostgreSQL connection and migration setup
└── tests
    ├── e2e
    └── integration
```

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/vadimbarashkov/url-shortener.git
    cd url-shortener
    ```

2. Install dependencies:

    ```bash
    go mod tidy
    ```

## Running the Application

### Using Docker

1. Build and start the services:

    ```bash
    docker-compose up --build
    ```

### Without Docker

1. Setup PostgreSQL.

2. Run the application:

    ```bash
    CONFIG_PATH=./configs/dev.yml make all
    ```

## API Documentation

The application is documented using Swagger and you can explore it using various UI tools.

## Running Tests

### Unit Tests

To run unit tests:

```bash
make test/unit
```

### Integration Tests

To run integration tests:

```bash
make test/integration
```

### E2E Tests

First, you need to launch the application, after which you can run the E2E tests:

```bash
CONFIG_PATH=./configs/stage.yml make test/e2e
```

You need to specify the path to the configuration file that you used to run the application.

## Database Migrations

The application automatically applies migrations from the `/migrations` directory, but you can run them manually using the `Makefile`:

```bash
# Create migration
make migrations/create $(MIGRATION_NAME)

# Run migrations
make migrations/up $(DATABASE_DSN)

# Rollback migrations
make migrations/down $(DATABASE_DSN)
```

## Application Configuration

The application is configured via YAML files. Application uses `CONFIG_PATH` to load configuration from YAML file. You need to set or pass environment variable when starting application.

```bash
# Set environment variable
export CONFIG_PATH=./configs/dev.yml

# Pass environment variable when starting application
CONFIG_PATH=./configs/dev.yml make all
```

Here is the basic structure of the configuration file:

```yaml
# dev | stage | prod
# default: dev
env: dev

# default: 7
short_code_length: 7

http_server:
  # default: 8080
  port: 8443
  # default: 5s
  read_timeout: 5s
  # default: 10s
  write_timeout: 10s
  # default: 1m
  idle_timeout: 1m
  # default: 1048576
  max_header_bytes: 1048576
  cert_file: ./crts/example.pem
  key_file: ./crts/example-key.pem

postgres:
  user: postgres
  password: postgres
  # default: localhost
  host: localhost
  # default: 5432
  port: 5432
  db: url_shortener
  # default: disable
  sslmode: disable
  # default: 5m
  conn_max_idle_time: 5m
  # default: 30m
  conn_max_lifetime: 30m
  # default: 5
  max_idle_conns: 5
  # default: 25
  max_open_conns: 25
```

The behavior of the application depends on the environment passed in the configuration file:

1. `dev` - http server doesn't use TLS certificates and logging is structured without JSON.
2. `stage` - http server doesn't use TLS certificates and logging is structured with JSON.
3. `prod` - http server uses TLS certificates and logging is structured with JSON.

## Contributing

Contributions are welcome! Suggest your ideas in issues or pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
