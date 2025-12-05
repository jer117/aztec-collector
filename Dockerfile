# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o aztec-collector ./cmd/collector/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Add ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/aztec-collector /app/aztec-collector

# Copy example config
COPY example-config.yaml /app/config.yaml

# Create non-root user
RUN adduser -D -g '' collector
USER collector

# Expose metrics port
EXPOSE 13285

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:13285/health || exit 1

ENTRYPOINT ["/app/aztec-collector"]
CMD ["--config", "/app/config.yaml"]

