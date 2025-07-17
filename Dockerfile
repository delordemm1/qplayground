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
FROM golang:1.25-rc-alpine AS backend-builder
# FROM golang:1.24-alpine AS backend-builder

# Install libvips and its development dependencies for Cgo compilation
RUN apk update && apk add --no-cache \
    vips \
    vips-dev \
    pkgconfig \
    jpeg-dev \
    tiff-dev \
    libxml2 \
    zlib \
    build-base \
    gcc
# Set the working directory for the backend build
WORKDIR /app

# Copy Go module files to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/main ./cmd/app/main.go


# =========================================================================
# Stage 3: Final Production Image
#
# This stage starts from a minimal Alpine image and copies only the
# necessary artifacts from the previous stages for a small and secure image.
# =========================================================================
FROM alpine:latest

# Install libvips and its dependencies
# The `vips` package provides the core libvips library.
# You might need to add other packages depending on the image formats
# your Go application needs to handle (e.g., webp, tiff, jpeg, png).
# `pkgconfig` is also often needed if your Go binding uses cgo to find libvips.
#
# Check Alpine's package repository for the exact package names for libvips and its desired dependencies:
# https://pkgs.alpinelinux.org/packages
RUN apk update && apk add --no-cache \
    vips \
    libjpeg \
    libpng \
    libwebp \
    libxml2 \
    zlib \
    && rm -rf /var/cache/apk/* # Clean apk cache for this stage too!

# Create a non-root user and group for enhanced security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory inside the final container
WORKDIR /app

# Copy the compiled Go binary from the backend-builder stage
COPY --from=backend-builder /app/main .

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
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Expose the port the Go application will listen on
EXPOSE 8084

# The command to run when the container starts
CMD ["./main"]
