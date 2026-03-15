# ---- Build stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for layer caching
COPY go.mod ./
COPY go.sum* ./
RUN go mod tidy && go mod download

# Copy source
COPY . .

# Build the binary
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOTOOLCHAIN=auto
RUN go build -ldflags="-s -w" -o /app/pulse-service ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.20

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
