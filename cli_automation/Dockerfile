# =========================================================================
# Stage 1: Backend Binary Builder (Ubuntu-based)
#
# This stage uses an Ubuntu-based Go image to compile the CLI app.
# =========================================================================
FROM golang:1.25-rc-bookworm AS builder

# Install dependencies using apt-get for Ubuntu
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    pkg-config \
    zlib1g-dev \
    gcc \
    # Clean up the apt cache to reduce layer size
    && rm -rf /var/lib/apt/lists/*

# Set the working directory for the backend build
WORKDIR /app

# Copy Go module files to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/cli ./cmd/main.go

# =========================================================================
# Stage 2: Final Production Image with Playwright
#
# This stage uses the official Playwright image based on Ubuntu 24.04 (Noble).
# =========================================================================
FROM mcr.microsoft.com/playwright:v1.52.0-noble

# Create a non-root system user and group for enhanced security
# RUN addgroup --system appgroup && adduser --system --ingroup appgroup appuser

# Set the working directory inside the final container
WORKDIR /app

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/cli .

# Ensure the non-root user owns all the application files
# RUN chown -R appuser:appgroup /app

# Switch to the non-root user
# USER appuser

# The command to run when the container starts
ENTRYPOINT ["/app/cli"]
