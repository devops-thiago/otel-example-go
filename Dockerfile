# Use the official Go image as the base image
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates for dependency downloads
RUN apk add --no-cache git ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Configure Go proxy and checksum database
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV CGO_ENABLED=0
ENV GOOS=linux

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies with retry mechanism
RUN go mod download || (sleep 5 && go mod download) || go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -a -installsuffix cgo -ldflags="-w -s" -o api ./cmd/api

# Use a minimal base image for the final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/api .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./api"]
