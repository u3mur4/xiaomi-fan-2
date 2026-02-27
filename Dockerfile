# --- Stage 1: The Builder ---
# Use an official Go image that is multi-arch.
# Using a specific version is better for reproducibility.
# Alpine is used for a smaller build stage.
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to leverage Docker's build cache.
# This step will only be re-run if these files change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application for the target platform.
# TARGETPLATFORM is an automatic variable provided by buildx (e.g., linux/amd64, linux/arm64)
# CGO_ENABLED=0 creates a static binary without any C dependencies.
# -ldflags="-s -w" strips debug symbols to make the binary smaller.
ARG TARGETPLATFORM
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /xiaomi-fan-2 .


# --- Stage 2: The Final Image ---
# Use a minimal, multi-arch base image like Alpine Linux.
FROM alpine:latest

# It's a security best practice to run as a non-root user.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Set the working directory
WORKDIR /app

# Copy the static binary from the builder stage
COPY --from=builder /xiaomi-fan-2 /app/xiaomi-fan-2

# EXPOSE the port your server listens on.
EXPOSE 35352

# Command to run the application
CMD ["/app/xiaomi-fan-2", "server"]