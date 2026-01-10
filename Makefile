.PHONY: help build run dev docker-up docker-down docker-build migrate-up migrate-down migrate-create test clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the server binary
	@echo "Building tracehub-server..."
	@go build -o bin/tracehub-server ./cmd/server

run: ## Run the server locally
	@echo "Running tracehub-server..."
	@go run ./cmd/server/main.go

dev: ## Run with hot reload (requires air)
	@echo "Running in development mode..."
	@air

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@docker-compose build

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "Server: http://localhost:8080"
	@echo "Database: localhost:5432"

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-logs: ## Show Docker logs
	@docker-compose logs -f server

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/tracehub?sslmode=disable" up

migrate-down: ## Rollback last migration
	@echo "Rolling back migration..."
	@migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/tracehub?sslmode=disable" down 1

migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name)
	@echo "Creating migration: $(name)"
	@migrate create -ext sql -dir migrations -seq $(name)

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

coverage: test ## Generate coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@docker-compose down -v

install-tools: ## Install development tools
	@echo "Installing tools..."
	@go install github.com/air-verse/air@latest
	@echo "Tools installed!"

.DEFAULT_GOAL := help
