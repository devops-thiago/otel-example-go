# Stage 1: Dependencies
FROM golang:1.24-alpine AS deps

# Install certificates
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Configure Go environment
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies with retry mechanism
RUN go mod download || (sleep 5 && go mod download) || go mod download

# Stage 2: Builder
FROM golang:1.24-alpine AS builder

# Build arguments
ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF

# Install git for version information
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Configure Go environment for minimal binary
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Copy dependencies from previous stage
COPY --from=deps /go/pkg /go/pkg
COPY go.mod go.sum ./

# Copy only necessary source code
# This ensures we don't accidentally include test files or other unnecessary files
COPY cmd/api/main.go ./cmd/api/
COPY internal/config/*.go ./internal/config/
COPY internal/database/*.go ./internal/database/
COPY internal/handlers/*.go ./internal/handlers/
COPY internal/logging/*.go ./internal/logging/
COPY internal/middleware/*.go ./internal/middleware/
COPY internal/models/*.go ./internal/models/
COPY internal/repository/*.go ./internal/repository/
COPY pkg/utils/*.go ./pkg/utils/

# Remove any test files that might have been copied (safety measure)
RUN find . -name '*_test.go' -type f -delete

# Build the application with version information and optimizations
RUN go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildDate=${BUILD_DATE} -X main.gitCommit=${VCS_REF}" \
    -o api ./cmd/api && \
    # Verify the binary was built
    test -f api

# Use a minimal base image for the final stage
FROM alpine:latest

# Build arguments for labels
ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF

# Labels
LABEL org.opencontainers.image.title="OpenTelemetry Go CRUD API" \
      org.opencontainers.image.description="Go REST API with OpenTelemetry instrumentation" \
      org.opencontainers.image.authors="support@arquivolivre.com.br" \
      org.opencontainers.image.vendor="Arquivo Livre" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.source="https://github.com/thiagorb/otel-example-go" \
      org.opencontainers.image.documentation="https://github.com/thiagorb/otel-example-go/blob/main/README.md" \
      org.opencontainers.image.licenses="MIT"

# Install runtime dependencies and create non-root user
RUN apk --no-cache add ca-certificates tzdata wget && \
    addgroup -g 1000 -S appuser && \
    adduser -u 1000 -S appuser -G appuser && \
    mkdir -p /app && \
    chown -R appuser:appuser /app

# Set the working directory
WORKDIR /app

# Copy files with correct ownership
COPY --from=builder --chown=appuser:appuser /app/api .
# Copy timezone data from deps stage
COPY --from=deps /usr/share/zoneinfo /usr/share/zoneinfo

# Use non-root user
USER appuser

# Expose the port the app runs on
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Command to run the application
CMD ["./api"]
