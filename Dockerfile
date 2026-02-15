# =============================================================================
# Moltbunker Multi-Stage Dockerfile
# Builds: moltbunkerd (daemon), moltbunker (CLI), moltbunker-api (HTTP API)
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Builder
# -----------------------------------------------------------------------------
FROM golang:1.24-bookworm AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

WORKDIR /build

# Copy dependency manifests first for layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build all three binaries with static linking and stripped symbols
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /out/moltbunkerd ./cmd/daemon

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /out/moltbunker ./cmd/cli

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /out/moltbunker-api ./cmd/api

# -----------------------------------------------------------------------------
# Stage 2: Runtime
# -----------------------------------------------------------------------------
FROM debian:bookworm-slim

# Install minimal runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy binaries from builder
COPY --from=builder /out/moltbunkerd /usr/local/bin/moltbunkerd
COPY --from=builder /out/moltbunker /usr/local/bin/moltbunker
COPY --from=builder /out/moltbunker-api /usr/local/bin/moltbunker-api

# Copy default configuration
COPY configs/daemon.yaml /etc/moltbunker/daemon.yaml

# Create non-root user and group
RUN groupadd --system moltbunker && \
    useradd --system --gid moltbunker --create-home --home-dir /home/moltbunker moltbunker

# Create data and runtime directories
RUN mkdir -p /var/lib/moltbunker/data \
             /var/lib/moltbunker/keys \
             /var/lib/moltbunker/keystore \
             /var/lib/moltbunker/tor \
             /var/lib/moltbunker/cache \
             /var/lib/moltbunker/containers \
             /run/moltbunker \
    && chown -R moltbunker:moltbunker /var/lib/moltbunker /run/moltbunker

# Data volume
VOLUME ["/var/lib/moltbunker"]

# Switch to non-root user
USER moltbunker

# Expose ports:
#   8080 - HTTP API
#   9090 - P2P base (DHT/libp2p)
#   9091 - P2P TLS transport
EXPOSE 8080 9090 9091

# Health check via CLI doctor command
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD ["moltbunker", "doctor"] || exit 1

# Entrypoint: run the daemon
ENTRYPOINT ["moltbunkerd"]
CMD ["--data=/var/lib/moltbunker"]

# OCI image labels
LABEL org.opencontainers.image.title="moltbunker" \
      org.opencontainers.image.description="Permissionless, fully encrypted P2P network for containerized compute resources" \
      org.opencontainers.image.url="https://github.com/moltbunker/moltbunker" \
      org.opencontainers.image.source="https://github.com/moltbunker/moltbunker" \
      org.opencontainers.image.vendor="Moltbunker" \
      org.opencontainers.image.licenses="MIT"
