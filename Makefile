.PHONY: dev build run swag test test-e2e test-coverage loadtest loadtest-smoke loadtest-docker up seed seed-local down help

# Prefer `docker compose` (V2); fall back to `docker-compose` (V1).
DOCKER_COMPOSE ?= $(shell docker compose version >/dev/null 2>&1 && echo docker compose || echo docker-compose)

# Packages for default test/coverage (exclude ./e2e — needs Docker; use test-e2e).
TEST_PKGS ?= $(shell go list ./... | grep -v '/e2e$$')

up:
	@echo "Starting Docker Compose stack..."
	$(DOCKER_COMPOSE) up -d --build

down:
	@echo "Stopping Docker Compose stack..."
	$(DOCKER_COMPOSE) down

seed:
	@echo "Seeding database..."
	$(DOCKER_COMPOSE) exec app ./seeder

# Seed using local Go binary; uses .env (e.g. DB_HOST=localhost). Postgres must be reachable.
seed-local:
	@echo "Seeding database (local go run ./cmd/seed)..."
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
	gotestsum --format testname -- -count=1 -cover $(TEST_PKGS)

test-e2e:
	@echo "E2E tests (Docker: testcontainers Postgres + in-process API; or set E2E_BASE_URL for an external server)..."
	gotestsum --format testname -- -count=1 -v ./e2e/

# Load test: k6 against GET .../slots/list (see loadtest/RESULTS.md). Requires `make seed` and running API.
BASE_URL ?= http://127.0.0.1:8080
LOADTEST_RATE ?= 50
LOADTEST_DURATION ?= 90s

loadtest:
	@command -v k6 >/dev/null 2>&1 || { echo "Install k6: https://grafana.com/docs/k6/latest/set-up/install-k6/"; exit 1; }
	k6 run loadtest/k6/slots_list.js -e BASE_URL=$(BASE_URL) -e LOADTEST_RATE=$(LOADTEST_RATE) -e LOADTEST_DURATION=$(LOADTEST_DURATION)

loadtest-smoke:
	@command -v k6 >/dev/null 2>&1 || { echo "Install k6: https://grafana.com/docs/k6/latest/set-up/install-k6/"; exit 1; }
	k6 run loadtest/k6/slots_list.js -e BASE_URL=$(BASE_URL) -e LOADTEST_PROFILE=smoke

# k6 via Docker (no local k6). Linux: set BASE_URL to host IP if host.docker.internal is unavailable.
loadtest-docker:
	docker run --rm -i --add-host=host.docker.internal:host-gateway \
		-e BASE_URL=http://host.docker.internal:8080 \
		-e LOADTEST_RATE=$(LOADTEST_RATE) \
		-e LOADTEST_DURATION=$(LOADTEST_DURATION) \
		grafana/k6:0.57.0 run - < loadtest/k6/slots_list.js

test-coverage:
	@echo "Running tests with coverage profile (coverage.out)..."
	go test $(TEST_PKGS) -count=1 -cover -coverprofile=coverage.out -covermode=atomic
	@echo ""
	@echo "Per-file summary (last line = total):"
	@go tool cover -func=coverage.out

lint:
	@echo "Running linter..."
	golangci-lint run ./...

lint-fix:
	@echo "Running linter with autofix..."
	golangci-lint run --fix ./...

help:
	@echo "Available commands:"
	@echo "  up             - Start project with dependencies (docker compose up -d --build)"
	@echo "  down           - Stop Docker Compose stack"
	@echo "  seed           - Fill DB via docker compose"
	@echo "  seed-local     - Fill DB via go run ./cmd/seed (.env → localhost Postgres, no Docker)"
	@echo "  dev            - Start Air hot reload (with Swagger hot reload)"
	@echo "  build          - Build Go app into ./bin/app (Swagger updated first)"
	@echo "  run            - Run Go app (no reload)"
	@echo "  swag           - Generate Swagger docs"
	@echo "  test           - Run tests with gotestsum (excludes ./e2e; use test-e2e for E2E)"
	@echo "  test-e2e       - HTTP E2E (testcontainers Postgres by default; optional E2E_BASE_URL=... for external API)"
	@echo "  loadtest       - k6 load test slots/list (BASE_URL, LOADTEST_RATE, LOADTEST_DURATION)"
	@echo "  loadtest-smoke - Short k6 sanity run (LOADTEST_PROFILE=smoke)"
	@echo "  loadtest-docker - Same as loadtest using grafana/k6 image"
	@echo "  test-coverage  - Full merged coverage profile -> coverage.out + go tool cover -func"
	@echo "  lint           - Run linter"
	@echo "  lint-fix       - Run linter with autofix"
	@echo "  help           - Show this help"
