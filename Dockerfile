# ---- Build stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy all source first (go mod tidy needs source to resolve imports)
COPY . .

# GOTOOLCHAIN=auto permite usar go 1.25+ quando exigido por dependências (pgx v5.9+)
ENV GOTOOLCHAIN=auto
ENV CGO_ENABLED=0
ENV GOOS=linux

# Generate go.sum and download dependencies
RUN go mod tidy && go mod download

# Build the binary
RUN go build -ldflags="-s -w" -o /app/pulse-service ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.21

WORKDIR /app

# Install CA certs and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/pulse-service .

# Run as non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 3015

ENTRYPOINT ["/app/pulse-service"]
