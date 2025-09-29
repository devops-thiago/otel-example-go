# Go REST API with Grafana Observability Stack

This project demonstrates a well-structured Go REST API with MySQL database and a complete Grafana observability stack (Tempo, Mimir, Loki, Alloy), following Go best practices and clean architecture patterns.

## 🎉 **COMPLETE OpenTelemetry Integration**

The application now has **full OpenTelemetry support** for traces, metrics, and structured logging with the complete Grafana observability stack!

**✅ What's Working:**
- **🔍 Distributed Tracing** - Full trace collection with OTLP export to Tempo
- **📊 Metrics Collection** - Custom HTTP metrics exported to Mimir
- **📝 Structured Logging** - JSON logs with trace correlation using Logrus
- **🗄️ REST API** - Complete CRUD operations with full instrumentation
- **🐳 Docker Stack** - Grafana, Tempo, Mimir, Loki, Alloy all configured
- **⚡ Auto-Instrumentation** - HTTP requests, custom spans, and error tracking
- **🗃️ Database Tracing** - Complete SQL query instrumentation with XSAM/otelsql
- **📈 Database Metrics** - Connection pool stats and query performance metrics

**🔄 Future Enhancements:**
- Log export to Loki (currently stdout/structured JSON)
- Custom business metrics and dashboards
- Grafana dashboard templates

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go         # Application entry point
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── database/          # Database connection
│   ├── handlers/          # HTTP handlers and routes
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Data models and structs
│   └── repository/        # Data access layer
├── pkg/                   # Public packages
│   └── utils/             # Utility functions
├── Dockerfile             # Multi-stage Docker build
├── docker-compose.yml     # Docker Compose configuration
├── init.sql              # Database initialization
├── env.example           # Environment variables template
├── go.mod                # Go module dependencies
├── main.go               # Application wrapper
└── README.md             # Documentation
```

## Prerequisites

- Docker
- Docker Compose

## Getting Started

### Option 1: Production Build (Recommended)

1. **Clone and navigate to the project directory:**
   ```bash
   cd /path/to/otel-example-go
   ```

2. **Start the services:**
   ```bash
   docker-compose up --build
   ```

3. **If you encounter network issues during build, try the resilient build script:**
   ```bash
   ./build.sh
   docker-compose up
   ```

### Option 2: Development Mode (Network Issues Workaround)

If you're experiencing network connectivity issues during Docker builds:

```bash
# Use the development compose file that runs Go directly in container
docker-compose -f docker-compose.dev.yml up
```

### Option 3: Manual Build with Network Troubleshooting

```bash
# Build with host network (helps with corporate firewalls)
docker build --network=host -t otel-example-go .

# Or build without cache
docker build --no-cache -t otel-example-go .

# Then run with docker-compose
docker-compose up
```

### Access the Application
- Go app: http://localhost:8080
- MySQL: localhost:3306

## Services

### Core Application
- **Go API** (port 8080) - REST API with CRUD operations
- **MySQL** (port 3306) - Database with sample data

### Grafana Observability Stack
- **Grafana** (port 3000) - Visualization dashboard (admin/admin)
- **Tempo** (port 3200) - Distributed tracing backend
- **Mimir** (port 9009) - Metrics storage (Prometheus-compatible)
- **Loki** (port 3100) - Log aggregation
- **Alloy** (port 12345) - Telemetry collection and processing

### Storage
- **MinIO** (ports 9000/9001) - Object storage for Tempo, Mimir, and Loki

## Database Schema

The MySQL container is initialized with sample tables:
- `users` - Sample user data
- `products` - Sample product data

## Environment Variables

The Go application uses these environment variables:
- `DB_HOST` - Database host (default: mysql)
- `DB_PORT` - Database port (default: 3306)
- `DB_USER` - Database user (default: appuser)
- `DB_PASSWORD` - Database password (default: apppassword)
- `DB_NAME` - Database name (default: otel_example)

## Development

### Stop the services:
```bash
docker-compose down
```

### Stop and remove volumes (deletes database data):
```bash
docker-compose down -v
```

### View logs:
```bash
# All services
docker-compose logs

# Specific service
docker-compose logs app
docker-compose logs mysql
```

### Rebuild after code changes:
```bash
docker-compose up --build
```

## API Endpoints

### Health Checks
- `GET /health` - Service health check
- `GET /ready` - Service readiness check

### API Info
- `GET /api/` - API information

### Users
- `GET /api/users` - Get all users (with pagination)
  - Query params: `page` (default: 1), `limit` (default: 10, max: 100)
- `POST /api/users` - Create a new user
- `GET /api/users/:id` - Get user by ID
- `PUT /api/users/:id` - Update user by ID
- `DELETE /api/users/:id` - Delete user by ID

### Example API Calls

```bash
# Get all users
curl http://localhost:8080/api/users

# Get users with pagination
curl "http://localhost:8080/api/users?page=1&limit=5"

# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com", "bio": "Software Developer"}'

# Get user by ID
curl http://localhost:8080/api/users/1

# Update user
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "John Smith", "bio": "Senior Software Developer"}'

# Delete user
curl -X DELETE http://localhost:8080/api/users/1
```

## Architecture & Best Practices

This project follows Go best practices and clean architecture principles:

### Directory Structure
- `cmd/` - Application entry points
- `internal/` - Private application code (cannot be imported by other projects)
- `pkg/` - Public packages that can be imported
- Repository pattern for data access
- Separation of concerns with handlers, services, and repositories

### Features
- **RESTful API** with proper HTTP methods and status codes
- **Middleware** for logging, CORS, error handling, and recovery
- **Environment-based configuration** with sensible defaults
- **Database connection pooling** and health checks
- **Structured logging** and error handling
- **Input validation** and sanitization
- **Pagination** support for list endpoints
- **Graceful shutdown** handling
- **Docker multi-stage builds** for optimized images

### Database
- MySQL with connection pooling and OpenTelemetry instrumentation
- Repository pattern for clean data access with custom tracing
- Complete SQL query tracing using XSAM/otelsql
- Database connection pool metrics and performance monitoring
- Proper error handling and transaction support
- Database health checks

## Troubleshooting

### Network Issues During Docker Build

If you encounter network connectivity issues when building the Docker image:

#### Common Error Messages:
- `connection reset by peer`
- `EOF` errors from Go proxy
- `failed to solve: process "/bin/sh -c go mod download" did not complete successfully`

#### Solutions:

1. **Use the resilient build script:**
   ```bash
   ./build.sh
   ```

2. **Use development mode (bypasses build issues):**
   ```bash
   docker-compose -f docker-compose.dev.yml up
   ```

3. **Build with host networking:**
   ```bash
   docker build --network=host -t otel-example-go .
   ```

4. **Configure Docker daemon DNS:**
   Add to `/etc/docker/daemon.json`:
   ```json
   {
     "dns": ["8.8.8.8", "8.8.4.4"]
   }
   ```

5. **Corporate firewall/proxy:**
   ```bash
   docker build --build-arg HTTP_PROXY=http://your-proxy:port \
                --build-arg HTTPS_PROXY=http://your-proxy:port \
                -t otel-example-go .
   ```

6. **Use offline build (with vendor directory):**
   ```bash
   docker build -f Dockerfile.offline -t otel-example-go .
   ```

### Database Connection Issues

- Ensure MySQL container is healthy: `docker-compose logs mysql`
- Check if port 3306 is available: `netstat -an | grep 3306`
- Verify environment variables in docker-compose.yml

### Application Issues

- Check application logs: `docker-compose logs app`
- Verify the application is binding to 0.0.0.0:8080
- Test health endpoint: `curl http://localhost:8080/health`

## Notes

- The MySQL container includes a health check to ensure the database is ready before starting the Go application
- Database data is persisted using Docker volumes
- The Go application waits for MySQL to be healthy before starting
- All API responses follow a consistent JSON structure
- Environment variables are used for configuration with Docker-friendly defaults
- The project includes vendor directory for offline builds
- Multiple Dockerfile options available for different network scenarios
