# Multi-stage build cho ProxVN Server
FROM golang:1.24-alpine AS builder

# Install dependencies
RUN apk add --no-cache git make gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY src/backend/go.mod src/backend/go.sum ./
RUN go mod download

# Copy source code
COPY src/backend/ ./

# Build server
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=Docker" \
    -o proxvn-server ./cmd/server

# Build client (optional, for convenience)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=Docker" \
    -o proxvn-client ./cmd/client

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/proxvn-server .
COPY --from=builder /build/proxvn-client .

# Copy frontend files
COPY src/frontend/ ./frontend/

# Create directories
RUN mkdir -p /data /backups /logs

# Environment variables
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8882
ENV DB_PATH=/data/proxvn.db
ENV BACKUP_DIR=/backups
ENV LOG_LEVEL=info

# Expose ports
# 8882 - Main tunnel server
# 8881 - HTTP admin
# 10000-20000 - Public ports for tunnels
EXPOSE 8882 8881 10000-20000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8881/health || exit 1

# Run as non-root user
RUN addgroup -g 1000 proxvn && \
    adduser -D -u 1000 -G proxvn proxvn && \
    chown -R proxvn:proxvn /app /data /backups /logs

USER proxvn

# Default command
CMD ["./proxvn-server"]
