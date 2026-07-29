# Build Stage
FROM golang:1.25-alpine AS builder

# Install git for downloading dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy module files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary statically
RUN CGO_ENABLED=0 GOOS=linux go build -o tbd-bot .

# Final Stage (Tiny and Secure)
FROM alpine:latest

WORKDIR /app/

# Install ca-certificates so the bot can make HTTPS/WebSocket calls to Discord
# tzdata is useful for timezone-aware cron jobs
RUN apk --no-cache add ca-certificates tzdata

# Enforce UTC timezone as per spec
ENV TZ=UTC

# Copy the compiled binary from the builder stage
COPY --from=builder /app/tbd-bot .

# The bot health check port
EXPOSE 8080

# Native Docker HEALTHCHECK using wget against the /metrics endpoint.
# The HTTP server only starts after dbot.Init registers slash commands with
# Discord, so /metrics is unreachable for as long as that round trip takes. A
# short start period makes the container flap to unhealthy during a normal
# boot; 60s covers the registration without hiding a real hang, since
# deploy.yml still fails the deployment at 120s.
HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/metrics || exit 1

# Run as non-root user
USER nobody:nobody

CMD ["./tbd-bot"]

