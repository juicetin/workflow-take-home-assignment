# Database configuration
DATABASE_URL ?= postgres://workflow:workflow123@localhost:5876/workflow_engine?sslmode=disable
MIGRATIONS_PATH ?= ./migrations

# Build commands
.PHONY: build
build:
	go build -o bin/api ./main.go

.PHONY: build-migrate
build-migrate:
	go build -o bin/migrate ./scripts/migrate.go

# Migration commands
.PHONY: migrate-up
migrate-up:
	./scripts/migrate.sh up "$(DATABASE_URL)"

.PHONY: migrate-version
migrate-version:
	./scripts/migrate.sh version "$(DATABASE_URL)"

# Development commands
.PHONY: run
run:
	go run main.go

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -rf bin/

# Docker commands
.PHONY: docker-build
docker-build:
	docker build -t workflow-api .

.PHONY: docker-run
docker-run:
	docker-compose up --build api

# Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build            - Build the API binary"
	@echo "  build-migrate    - Build the migrate CLI tool"
	@echo "  migrate-up       - Run database migrations"
	@echo "  migrate-version  - Show current migration version"
	@echo "  run              - Run the API server"
	@echo "  test             - Run tests"
	@echo "  clean            - Clean build artifacts"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run with Docker Compose"