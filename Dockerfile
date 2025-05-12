FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Install required packages
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with version info
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o telegpt ./cmd/bot

# Create a minimal production image
FROM alpine:latest

# Add CA certificates for HTTPS requests and necessary utilities
RUN apk --no-cache add ca-certificates curl

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/telegpt .

# Copy config.yaml.example (will be overridden by mounted config)
COPY --from=builder /app/config.yaml.example /app/config.yaml.example

# Create a healthcheck script, create a non-root user and set permissions
RUN echo '#!/bin/sh\npgrep telegpt >/dev/null || exit 1\nexit 0' > /app/healthcheck.sh && \
    chmod +x /app/healthcheck.sh && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup && \
    chown -R appuser:appgroup /app

# Set environment variables
ENV CONFIG_FILE=/app/config.yaml

# Configure health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 CMD ["/app/healthcheck.sh"]

# Switch to non-root user
USER appuser

# Set the entrypoint
ENTRYPOINT ["./telegpt"]
