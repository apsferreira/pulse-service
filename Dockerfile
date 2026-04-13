# ---- Build stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Dependências do sistema — layer separada, muda raramente
RUN apk add --no-cache git ca-certificates

# Copiar go.mod e go.sum ANTES do código fonte
# Isso garante que o cache de dependências só é invalidado quando mudam as deps
COPY go.mod go.sum ./

ENV GOTOOLCHAIN=local
ENV CGO_ENABLED=0
ENV GOOS=linux

# Download de dependências — layer cacheada enquanto go.mod/go.sum não mudarem
RUN go mod download

# Copiar o código fonte
COPY . .

# Build do binário
RUN go build -ldflags="-s -w" -o /app/pulse-service ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/pulse-service .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 3015

ENTRYPOINT ["/app/pulse-service"]
