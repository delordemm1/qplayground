# Use Playwright Go base image which includes browsers and dependencies
FROM mcr.microsoft.com/playwright/go:v1.44.0-jammy AS base

# Builder stage
FROM golang:1.25-rc-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git build-base

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the CLI binary
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/cli ./cmd/main.go

# Final stage
FROM mcr.microsoft.com/playwright/go:v1.44.0-jammy

# Create app user
RUN groupadd -r appgroup && useradd -r -g appgroup appuser

# Set working directory
WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/cli /app/cli

# Create output directory
RUN mkdir -p /app/output && chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Set entrypoint
ENTRYPOINT ["/app/cli"]