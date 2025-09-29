# OpenTelemetry Go Example

[![CI](https://img.shields.io/github/actions/workflow/status/devops-thiago/otel-example-go/ci.yml?branch=main&label=CI)](https://github.com/devops-thiago/otel-example-go/actions)
[![Go Version](https://img.shields.io/badge/go-1.24.x-00ADD8?logo=go)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/devops-thiago/otel-example-go)](https://goreportcard.com/report/github.com/devops-thiago/otel-example-go)
[![License](https://img.shields.io/github/license/devops-thiago/otel-example-go)](LICENSE)
[![Codecov](https://img.shields.io/codecov/c/github/devops-thiago/otel-example-go?label=coverage)](https://app.codecov.io/gh/devops-thiago/otel-example-go)
[![Sonar Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=devops-thiago_otel-example-go&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=devops-thiago_otel-example-go)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=devops-thiago_otel-example-go&metric=coverage)](https://sonarcloud.io/summary/new_code?id=devops-thiago_otel-example-go)
[![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-enabled-blue?logo=opentelemetry)](https://opentelemetry.io)
[![Docker](https://img.shields.io/badge/Docker-ready-blue?logo=docker)](https://www.docker.com)
[![Docker Hub](https://img.shields.io/docker/v/thiagosg/otel-crud-api-go?logo=docker&label=Docker%20Hub)](https://hub.docker.com/r/thiagosg/otel-crud-api-go)
[![Docker Pulls](https://img.shields.io/docker/pulls/thiagosg/otel-crud-api-go)](https://hub.docker.com/r/thiagosg/otel-crud-api-go)

A production-ready Go REST API with comprehensive OpenTelemetry instrumentation, featuring distributed tracing, metrics collection, and structured logging. Built with clean architecture principles and designed for cloud-native deployments.

## 📋 Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Deployment Options](#deployment-options)
- [API Documentation](#api-documentation)
- [Configuration](#configuration)
- [Observability](#observability)
- [Development](#development)
- [Testing](#testing)
- [Contributing](#contributing)

## ✨ Features

- **🚀 RESTful API** - Clean REST API with user CRUD operations
- **📊 Full Observability** - Distributed tracing, metrics, and structured logging
- **🔌 OpenTelemetry Native** - Built-in OTLP exporter support
- **🏗️ Clean Architecture** - Modular design with separation of concerns
- **🐳 Docker Ready** - Multi-stage Dockerfile with security best practices
- **🔒 Security First** - Non-root user, minimal attack surface, vulnerability scanning
- **🧪 Well Tested** - Comprehensive test coverage with mocking
- **📝 API Documentation** - Clear endpoint documentation
- **🔧 Configuration** - Environment-based configuration
- **💾 MySQL Integration** - Database connection with proper instrumentation

## 📚 Prerequisites

- Go 1.22+ (for local development)
- Docker & Docker Compose
- MySQL 8.0+ (or use the provided docker-compose)
- OpenTelemetry Collector (optional - included in full setup)

## 🚀 Quick Start

### Option 1: Full Stack (App + Database + Observability)

```bash
# Clone the repository
git clone https://github.com/devops-thiago/otel-example-go.git
cd otel-example-go

# Start everything with docker-compose
docker-compose up -d

# Check if services are running
docker-compose ps
```

**Access points:**
- API: http://localhost:8080
- Health: http://localhost:8080/health
- Metrics: http://localhost:8080/metrics

### Option 2: Run Locally

```bash
# Install dependencies
go mod download

# Set up environment variables
cp .env.example .env
# Edit .env with your configuration

# Run the application
go run cmd/api/main.go
```

## 🚢 Deployment Options

### Using Your Own OpenTelemetry Collector

If you already have an OpenTelemetry infrastructure:

```bash
# Use the app-only compose file
docker-compose -f docker-compose.app-only.yml up -d
```

**Required environment variables:**
```bash
# OpenTelemetry Configuration
OTEL_EXPORTER_OTLP_ENDPOINT=your-collector:4317
OTEL_SERVICE_NAME=otel-example-go

# Database Configuration
DB_HOST=your-mysql-host
DB_PORT=3306
DB_USER=your-db-user
DB_PASSWORD=your-db-password
DB_NAME=your-db-name
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: otel-example-go
spec:
  replicas: 3
  selector:
    matchLabels:
      app: otel-example-go
  template:
    metadata:
      labels:
        app: otel-example-go
    spec:
      containers:
      - name: api
        image: otel-example-go:latest
        env:
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "otel-collector.observability.svc.cluster.local:4317"
        # Add other environment variables
```

### Using Pre-built Docker Image

```bash
# Pull from Docker Hub
docker pull thiagosg/otel-crud-api-go:latest

# Run with custom environment
docker run -d \
  -p 8080:8080 \
  -e OTEL_EXPORTER_OTLP_ENDPOINT=your-collector:4317 \
  -e DB_HOST=your-db-host \
  -e DB_PORT=3306 \
  -e DB_USER=your-user \
  -e DB_PASSWORD=your-password \
  -e DB_NAME=your-database \
  thiagosg/otel-crud-api-go:latest
```

### Building Docker Image

```bash
# Build the image locally
docker build -t otel-example-go:latest .

# Build with version information
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  --build-arg VCS_REF=$(git rev-parse --short HEAD) \
  -t otel-example-go:latest .

# Build multi-platform image
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t otel-example-go:latest .
```

## 📖 API Documentation

### Health Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check endpoint |
| GET | `/ready` | Readiness check endpoint |
| GET | `/metrics` | Prometheus-compatible metrics |

### User API

| Method | Endpoint | Description | Request Body |
|--------|----------|-------------|---------------|
| GET | `/api/users` | List all users | - |
| GET | `/api/users/:id` | Get user by ID | - |
| POST | `/api/users` | Create new user | `{"name": "John", "email": "john@example.com", "bio": "Developer"}` |
| PUT | `/api/users/:id` | Update user | `{"name": "John Updated"}` |
| DELETE | `/api/users/:id` | Delete user | - |

### Example Requests

```bash
# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com", "bio": "Software Engineer"}'

# Get all users
curl http://localhost:8080/api/users

# Get user by ID
curl http://localhost:8080/api/users/1

# Update user
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "John Updated"}'

# Delete user
curl -X DELETE http://localhost:8080/api/users/1
```

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| **OpenTelemetry** | | |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP collector endpoint | `otel-collector:4317` |
| `OTEL_SERVICE_NAME` | Service name for telemetry | `otel-example-go` |
| `OTEL_ENABLE_TRACING` | Enable distributed tracing | `true` |
| `OTEL_ENABLE_METRICS` | Enable metrics collection | `true` |
| `OTEL_ENABLE_LOGGING` | Enable OTLP log export | `true` |
| **Database** | | |
| `DB_HOST` | MySQL host | `localhost` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_USER` | MySQL user | `root` |
| `DB_PASSWORD` | MySQL password | `password` |
| `DB_NAME` | MySQL database name | `otel_example` |
| **Server** | | |
| `SERVER_HOST` | API server host | `0.0.0.0` |
| `SERVER_PORT` | API server port | `8080` |
| `APP_ENV` | Application environment | `development` |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |

### Configuration File

Create a `.env` file in the project root:

```env
# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
OTEL_SERVICE_NAME=otel-example-go
OTEL_ENABLE_TRACING=true
OTEL_ENABLE_METRICS=true
OTEL_ENABLE_LOGGING=true

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=otel_example

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
APP_ENV=development
LOG_LEVEL=info
```

## 🔍 Observability

### OpenTelemetry Integration

This application exports telemetry data in OTLP format:

- **Traces**: Distributed tracing for all HTTP requests and database operations
- **Metrics**: Request duration, database connection pool, custom business metrics
- **Logs**: Structured logs with trace correlation

### Viewing Telemetry Data

#### Using Jaeger (included in docker-compose)

```bash
# Access Jaeger UI
open http://localhost:16686
```

#### Using Grafana (included in docker-compose)

```bash
# Access Grafana
open http://localhost:3000
# Default credentials: admin/admin
```

#### Custom Collectors

The application supports any OTLP-compatible collector:
- AWS X-Ray
- Google Cloud Trace
- Azure Monitor
- Datadog
- New Relic
- Elastic APM

## 🏗️ Project Structure

```
.
├── cmd/
│   └── api/              # Application entrypoints
│       └── main.go       # Main application
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and utilities
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   ├── repository/      # Data access layer
│   └── logging/         # Structured logging
├── pkg/                 # Public packages
│   └── utils/           # Utility functions
├── scripts/             # Utility scripts
│   ├── verify-docker-security.sh  # Docker security verification
│   └── analyze-docker-context.sh  # Docker build context analyzer
├── docker-compose.yml   # Full stack deployment
├── docker-compose.app-only.yml  # App-only deployment
├── Dockerfile           # Multi-stage Docker build
├── Makefile            # Build automation
├── go.mod              # Go dependencies
└── README.md           # This file
```

## 🔒 Security

### Docker Image Security

Our Docker images are built with security best practices:

- **Non-root user**: Runs as `appuser` (UID/GID 1000)
- **Minimal base image**: Uses Alpine Linux
- **No shell**: Reduces attack surface
- **Read-only filesystem compatible**: Can run with `--read-only`
- **No unnecessary packages**: Only required runtime dependencies
- **No test files**: Test files excluded from production image
- **Optimized build context**: Only necessary files copied
- **Security scanning**: Automated Trivy scans on every build
- **SBOM generation**: Software Bill of Materials for supply chain security

#### Verify Security

```bash
# Run security verification script
./scripts/verify-docker-security.sh

# Run with security options
docker run --rm \
  --security-opt=no-new-privileges:true \
  --read-only \
  --cap-drop=ALL \
  -p 8080:8080 \
  thiagosg/otel-crud-api-go:latest
```

## 🛠️ Development

### Local Development Setup

```bash
# Clone the repository
git clone https://github.com/thiagorb/otel-example-go.git
cd otel-example-go

# Install dependencies
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run locally with hot reload
go install github.com/cosmtrek/air@latest
air
```

### Building from Source

```bash
# Build binary
go build -o bin/api ./cmd/api

# Build with specific version
go build -ldflags "-X main.version=1.0.0" -o bin/api ./cmd/api
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make cover

# Generate HTML coverage report
make coverhtml

# Run specific package tests
go test ./internal/handlers -v

# Run tests with race detection
go test -race ./...
```

### Test Structure

- Unit tests: `*_test.go` files alongside source code
- Mocks: Using interfaces for dependency injection
- Integration tests: Testing HTTP endpoints with httptest
- Test coverage: Aiming for >80% coverage

## 🤝 Contributing

### Code Quality Standards

```bash
# Format code
make fmt

# Run linters
make lint

# Check formatting without changes
make fmt-check

# Run go vet
make vet

# Remove trailing whitespaces
make trim-whitespace
```

### Pre-commit Hook

```bash
# Enable pre-commit checks
ln -s ../../.githooks/pre-commit .git/hooks/pre-commit
```

### Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linters (`make test && make lint`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### CI/CD Pipeline

All PRs must pass:
- ✅ Linting (golangci-lint)
- ✅ Formatting (gofmt)
- ✅ Tests with coverage (sent to SonarCloud)
- ✅ Code quality analysis (SonarCloud)
- ✅ Build on multiple Go versions
- ✅ Docker build verification
- ✅ Security scanning (Trivy)

#### Automated Docker Builds

When code is merged to main:
1. Docker image is automatically built
2. Multi-platform images (linux/amd64, linux/arm64)
3. Pushed to Docker Hub: `thiagosg/otel-crud-api-go`
4. Tagged as:
   - `latest` - Latest main branch build
   - `main-<sha>` - Specific commit
   - `v*.*.*` - Semantic version tags
5. Security scanning with Trivy
6. SBOM (Software Bill of Materials) generation

## License

This project is licensed under the MIT License - see the LICENSE file for details.
