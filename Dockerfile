FROM golang:1.24-alpine AS deps

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org

COPY go.mod go.sum ./

RUN go mod download || (sleep 5 && go mod download) || go mod download

FROM golang:1.24-alpine AS builder

ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF

RUN apk add --no-cache git

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY --from=deps /go/pkg /go/pkg
COPY go.mod go.sum ./

COPY cmd/api/main.go ./cmd/api/
COPY internal/config/*.go ./internal/config/
COPY internal/database/*.go ./internal/database/
COPY internal/handlers/*.go ./internal/handlers/
COPY internal/logging/*.go ./internal/logging/
COPY internal/middleware/*.go ./internal/middleware/
COPY internal/models/*.go ./internal/models/
COPY internal/repository/*.go ./internal/repository/
COPY pkg/utils/*.go ./pkg/utils/

RUN find . -name '*_test.go' -type f -delete

RUN go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildDate=${BUILD_DATE} -X main.gitCommit=${VCS_REF}" \
    -o api ./cmd/api && \
    test -f api

FROM alpine:latest

ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF

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

RUN apk --no-cache add ca-certificates tzdata wget && \
    addgroup -g 1000 -S appuser && \
    adduser -u 1000 -S appuser -G appuser && \
    mkdir -p /app && \
    chown -R appuser:appuser /app

WORKDIR /app

COPY --from=builder --chown=root:root --chmod=755 /app/api .
COPY --from=deps /usr/share/zoneinfo /usr/share/zoneinfo

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./api"]
