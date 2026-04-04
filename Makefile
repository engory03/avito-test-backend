.PHONY: dev build run swag test test-coverage up seed down help

# Prefer `docker compose` (V2); fall back to `docker-compose` (V1).
DOCKER_COMPOSE ?= $(shell docker compose version >/dev/null 2>&1 && echo docker compose || echo docker-compose)

up:
	@echo "Starting Docker Compose stack..."
	$(DOCKER_COMPOSE) up -d --build

down:
	@echo "Stopping Docker Compose stack..."
	$(DOCKER_COMPOSE) down

seed:
	@echo "Seeding database..."
	$(DOCKER_COMPOSE) exec app ./seeder

dev:
	@echo "Generating Swagger docs..."
	@swag init --parseDependency --parseInternal
	@echo "Swagger docs generated"
	@echo "Starting dev mode (Air hot reload)..."
	air

build:
	@echo "Generating Swagger docs..."
	@swag init --parseDependency --parseInternal
	@echo "Swagger docs generated"
	@echo "Building Go application..."
	@mkdir -p bin
	go build -o bin/app ./cmd/app
	@echo "Build complete: bin/app"

run:
	@echo "Generating Swagger docs..."
	@swag init --parseDependency --parseInternal
	@echo "Swagger docs generated"
	@echo "Running Go application..."
	go run ./cmd/app

swag:
	@echo "Generating Swagger docs..."
	@swag init --parseDependency --parseInternal -g ./cmd/app/main.go
	@echo "Swagger docs generated"

test:
	@echo "Running tests with gotestsum..."
	gotestsum --format testname -- -count=1 -cover ./...

test-coverage:
	@echo "Running tests with coverage profile (coverage.out)..."
	go test ./... -count=1 -cover -coverprofile=coverage.out -covermode=atomic
	@echo ""
	@echo "Per-file summary (last line = total):"
	@go tool cover -func=coverage.out

help:
	@echo "Available commands:"
	@echo "  up             - Start project with dependencies (docker compose up -d --build)"
	@echo "  down           - Stop Docker Compose stack"
	@echo "  seed           - Fill DB via docker compose"
	@echo "  dev            - Start Air hot reload (with Swagger hot reload)"
	@echo "  build          - Build Go app into ./bin/app (Swagger updated first)"
	@echo "  run            - Run Go app (no reload)"
	@echo "  swag           - Generate Swagger docs"
	@echo "  test           - Run tests with gotestsum (with per-package -cover)"
	@echo "  test-coverage  - Full merged coverage profile -> coverage.out + go tool cover -func"
	@echo "  help           - Show this help"
