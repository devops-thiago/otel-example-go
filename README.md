### otel-example-go

[![CI](https://img.shields.io/github/actions/workflow/status/thiagorb/otel-example-go/ci.yml?branch=main)](https://github.com/thiagorb/otel-example-go/actions)
[![Go Version](https://img.shields.io/badge/go-1.24.x-00ADD8?logo=go)](#)
[![Codecov](https://img.shields.io/codecov/c/github/thiagorb/otel-example-go?label=coverage)](https://app.codecov.io/gh/thiagorb/otel-example-go)
[![Sonar Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=devops-thiago_otel-example-go&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=devops-thiago_otel-example-go)
[![Sonar Coverage](https://sonarcloud.io/api/project_badges/measure?project=devops-thiago_otel-example-go&metric=coverage)](https://sonarcloud.io/summary/new_code?id=devops-thiago_otel-example-go)

Go REST API with MySQL and OpenTelemetry integration (traces, metrics, logs) designed to run with Grafana stack or any OTLP-compatible collector.

### Quick start

```bash
docker-compose up --build -d
```

App: `http://localhost:8080`.

### App-only (use your own OpenTelemetry collector)

If you already have an OTLP collector, you can deploy only the app:

```bash
docker-compose -f docker-compose.app-only.yml up --build -d
```

Configure `OTEL_EXPORTER_OTLP_ENDPOINT` to point to your collector (default `otel-collector:4317`). Database variables (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`) must point to your DB.

### Test and coverage

```bash
make test
make cover   # writes coverage.out and prints summary
make coverhtml  # opens coverage.html
```

### API endpoints

- GET `/health`, GET `/ready`
- GET `/api/`
- CRUD under `/api/users`

### Configuration

- `OTEL_EXPORTER_OTLP_ENDPOINT` (e.g., `collector:4317`)
- `OTEL_SERVICE_NAME`, `OTEL_ENABLE_TRACING`, `OTEL_ENABLE_METRICS`, `OTEL_ENABLE_LOGGING`
- DB envs: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`

### Project structure

```
cmd/, internal/, pkg/, Dockerfile, docker-compose.yml
```

### Troubleshooting

- If Docker build fails due to network, retry build or use your environment’s proxy.
