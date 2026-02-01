# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build daemon and CLI
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/moltbunker-daemon ./cmd/daemon
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/moltbunker ./cmd/cli

# Production stage
FROM alpine:3.19

RUN apk add --no-cache \
    ca-certificates \
    tor \
    containerd \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -S moltbunker && adduser -S moltbunker -G moltbunker

# Copy binaries from builder
COPY --from=builder /bin/moltbunker-daemon /usr/local/bin/
COPY --from=builder /bin/moltbunker /usr/local/bin/

# Copy default config
COPY configs/daemon.yaml /etc/moltbunker/daemon.yaml

# Create data directories
RUN mkdir -p /var/lib/moltbunker/data \
             /var/lib/moltbunker/keys \
             /var/lib/moltbunker/tor \
             /var/lib/moltbunker/containers \
    && chown -R moltbunker:moltbunker /var/lib/moltbunker

# Switch to non-root user
USER moltbunker

# Expose P2P port and Tor SOCKS
EXPOSE 9000 9050

# Data volume
VOLUME ["/var/lib/moltbunker"]

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD moltbunker status || exit 1

# Default command
ENTRYPOINT ["moltbunker-daemon"]
CMD ["--data=/var/lib/moltbunker", "--port=9000"]
