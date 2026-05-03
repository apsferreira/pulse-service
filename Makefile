.PHONY: build run test lint swagger vendor docker-build docker-push help

IMAGE ?= ghcr.io/institutoitinerante/pulse-service
TAG   ?= latest

## build: compila o binário usando vendor
build:
	CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags="-s -w" -o bin/pulse-service ./cmd/server

## run: executa localmente (requer DATABASE_URL e SERVICE_TOKEN)
run:
	go run ./cmd/server

## test: executa testes unitários
test:
	go test -mod=vendor -race -count=1 ./...

## lint: executa golangci-lint
lint:
	golangci-lint run ./...

## swagger: regenera docs OpenAPI a partir das anotações swag
swagger:
	docker run --rm \
		-v "$(shell pwd)":/app \
		-w /app \
		golang:1.25-alpine \
		sh -c "GOTOOLCHAIN=auto go install github.com/swaggo/swag/cmd/swag@latest && \
		       swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal"
	@echo "Docs gerados em docs/swagger/"
	@echo "Swagger UI: https://pulse-api.institutoitinerante.com.br/swagger/index.html"

## vendor: atualiza vendor com go mod vendor
vendor:
	docker run --rm \
		-v "$(shell pwd)":/app \
		-w /app \
		golang:1.25-alpine \
		sh -c "GOTOOLCHAIN=auto go mod tidy && go mod vendor"

## docker-build: constrói a imagem Docker
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-push: pusha imagem para GHCR
docker-push: docker-build
	docker push $(IMAGE):$(TAG)

## help: lista targets disponíveis
help:
	@grep -E '^##' Makefile | sed 's/## //'
