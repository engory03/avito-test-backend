.PHONY: dev build run swag test up seed down help

# Prefer `docker compose` (V2); fall back to `docker-compose` (V1).
DOCKER_COMPOSE ?= $(shell docker compose version >/dev/null 2>&1 && echo docker compose || echo docker-compose)

up:
	@echo "Starting Docker Compose stack..."
	$(DOCKER_COMPOSE) up -d --build

down:
	@echo "Stopping Docker Compose stack..."
	$(DOCKER_COMPOSE) down

seed:
	@echo "Seeding database (expects Postgres reachable per .env, e.g. after make up)..."
	go run ./cmd/seed

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
	gotestsum --format testname

help:
	@echo "Available commands:"
	@echo "  up     - Start project with dependencies (docker compose up -d --build)"
	@echo "  down   - Stop Docker Compose stack"
	@echo "  seed   - Fill DB with test data (go run ./cmd/seed; use .env, DB_HOST=localhost after make up)"
	@echo "  dev    - Start Air hot reload (with Swagger hot reload)"
	@echo "  build  - Build Go app into ./bin/app (Swagger updated first)"
	@echo "  run    - Run Go app (no reload)"
	@echo "  swag   - Generate Swagger docs"
	@echo "  test   - Run tests with gotestsum"
	@echo "  help   - Show this help"
