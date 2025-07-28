# =========================================================================
# Stage 1: Frontend Asset Builder
#
# This stage uses a Node.js image to install dependencies and build the
# static frontend assets (JS, CSS) using Vite/SvelteKit.
# =========================================================================
FROM oven/bun:1-slim AS frontend-builder

# Set the working directory for the frontend build
WORKDIR /app

# Copy package.json and lock files
COPY package.json bun.lock ./

# Install frontend dependencies
RUN bun install

# Copy the rest of the frontend source code
COPY . .

# Build the frontend assets for production
# This command runs the "build" script in your package.json
RUN bun run build


# =========================================================================
# Stage 2: Backend Binary Builder
#
# This stage uses a Go image to compile the backend application into a
# single, statically-linked executable.
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
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/app ./cmd/app/main.go

# =========================================================================
# Stage 2: Final Production Image with Playwright
#
# This stage uses the official Playwright image based on Ubuntu 24.04 (Noble).
# =========================================================================
FROM mcr.microsoft.com/playwright:v1.52.0-noble

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    pkg-config \
    zlib1g-dev \
    gcc \
    # Clean up the apt cache to reduce layer size
    && rm -rf /var/lib/apt/lists/*

# Set the working directory inside the final container
WORKDIR /app

# Copy the compiled Go binary from the backend-builder stage
COPY --from=backend-builder /app/app .

# Copy the built frontend assets from the frontend-builder stage.
# Inertia.js/Vite typically puts everything into a 'public/build' directory.
# We copy the entire 'public' directory which will contain this build folder.
COPY --from=frontend-builder /app/public ./public

# Copy the separate 'static' directory as requested
COPY --from=backend-builder /app/static ./static

# Copy the resources directory which contains HTML templates ---
COPY --from=backend-builder /app/resources ./resources

# Create a .env file
RUN touch .env

# Ensure the non-root user owns all the application files
# RUN chown -R appuser:appgroup /app

# Switch to the non-root user
# USER appuser

# Expose the port the Go application will listen on
EXPOSE 8084

# The command to run when the container starts
CMD ["./app"]
