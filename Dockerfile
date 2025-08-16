# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shbucket ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl nginx apache2-utils
WORKDIR /app

# Create directories
RUN mkdir -p /app/storage /app/config /app/certs /app/logs

# Copy binary
COPY --from=builder /app/shbucket .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Set permissions
RUN chmod +x shbucket

# Create non-root user
RUN addgroup -g 1001 -S shbucket && \
    adduser -S shbucket -u 1001 -G shbucket && \
    chown -R shbucket:shbucket /app

USER shbucket

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["./shbucket"]