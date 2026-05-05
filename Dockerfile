# ---- Build stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copiar go.mod e go.sum ANTES do código fonte
# Isso garante que o cache de dependências só é invalidado quando mudam as deps
COPY go.mod go.sum ./

ENV GOTOOLCHAIN=auto
ENV CGO_ENABLED=0
ENV GOOS=linux

# Baixar dependências (cacheado enquanto go.mod/go.sum não mudam)
RUN go mod download

# Copiar o código fonte
COPY . .

# Build do binário
RUN go build -ldflags="-s -w" -o /app/pulse-service ./cmd/server

# ---- Runtime stage ----
# Sem apk add — copia certs e tzdata do builder para evitar DNS no build
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/pulse-service .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 3015

ENTRYPOINT ["/app/pulse-service"]
