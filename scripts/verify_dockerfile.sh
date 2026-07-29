#!/usr/bin/env bash
set -euo pipefail

echo "==> Verifying Dockerfile static requirements..."

DOCKERFILE="Dockerfile"

if [ ! -f "$DOCKERFILE" ]; then
    echo "ERROR: Dockerfile not found!"
    exit 1
fi

# Check multi-stage base image is a golang image new enough for go.mod
if ! grep -E -q "FROM golang:[0-9]+\.[0-9]+" "$DOCKERFILE"; then
    echo "ERROR: Dockerfile does not build from a versioned golang base image"
    exit 1
fi

IMAGE_GO_VERSION=$(grep -E -o "FROM golang:[0-9]+\.[0-9]+" "$DOCKERFILE" | head -1 | cut -d: -f2)
MOD_GO_VERSION=$(grep -E -o "^go [0-9]+\.[0-9]+" go.mod | head -1 | cut -d' ' -f2)
if [ "$(printf '%s\n%s\n' "$MOD_GO_VERSION" "$IMAGE_GO_VERSION" | sort -V | head -1)" != "$MOD_GO_VERSION" ]; then
    echo "ERROR: builder image golang:$IMAGE_GO_VERSION is older than go.mod requirement go $MOD_GO_VERSION"
    exit 1
fi

# Check CGO_ENABLED=0
if ! grep -q "CGO_ENABLED=0" "$DOCKERFILE"; then
    echo "ERROR: Dockerfile does not set CGO_ENABLED=0"
    exit 1
fi

# Check HEALTHCHECK directive
if ! grep -q "HEALTHCHECK" "$DOCKERFILE"; then
    echo "ERROR: Dockerfile does not contain HEALTHCHECK directive"
    exit 1
fi

# Check wget usage in HEALTHCHECK
if ! grep -q "wget" "$DOCKERFILE"; then
    echo "ERROR: HEALTHCHECK does not use wget"
    exit 1
fi

# Check target metrics endpoint in HEALTHCHECK
if ! grep -E -q "http://localhost:8080/metrics|:8080/metrics" "$DOCKERFILE"; then
    echo "ERROR: HEALTHCHECK does not target http://localhost:8080/metrics"
    exit 1
fi

echo "✓ Static Dockerfile verification passed."

echo "==> Verifying static Go build with CGO_ENABLED=0..."
CGO_ENABLED=0 GOOS=linux "${GO_BIN:-go}" build -o /dev/null .
echo "✓ Go static compilation (CGO_ENABLED=0 GOOS=linux) succeeded."

if command -v docker >/dev/null 2>&1; then
    echo "==> Running docker build and docker inspect..."
    docker build -t tbd-bot:test .
    HEALTHCHECK_CONFIG=$(docker inspect --format='{{json .Config.Healthcheck}}' tbd-bot:test)
    echo "Docker inspect healthcheck output: $HEALTHCHECK_CONFIG"
    if [ "$HEALTHCHECK_CONFIG" = "null" ] || [ -z "$HEALTHCHECK_CONFIG" ]; then
        echo "ERROR: docker inspect reported null or empty Healthcheck config!"
        exit 1
    fi
    echo "✓ Docker build and docker inspect validation succeeded."
else
    echo "ℹ Docker CLI not detected in current environment. Static validation and CGO_ENABLED=0 compilation passed cleanly."
fi
