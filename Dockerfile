# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies for CGO (required by SQLite)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled for SQLite support
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o minim .

# Runtime stage
FROM alpine:3.20

# Install runtime dependencies for SQLite
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/minim .

# Copy static files for the web UI
COPY --from=builder /app/static ./static

# Create data directory
RUN mkdir -p /root/.minim

# Expose default port
EXPOSE 3333

# Run the server directly (execserver runs in foreground)
CMD ["./minim", "execserver"]
