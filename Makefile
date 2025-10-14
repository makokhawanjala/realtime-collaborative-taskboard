.PHONY: help build run test clean docker-up docker-down migrate seed lint

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building..."
	@go build -o bin/server cmd/server/main.go

run: ## Run the application
	@echo "Running..."
	@go run cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./tests/integration/...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	@docker-compose -f docker/docker-compose.yml up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	@docker-compose -f docker/docker-compose.yml down

docker-logs: ## View logs
	@docker-compose -f docker/docker-compose.yml logs -f

migrate: ## Run database migrations
	@echo "Running migrations..."
	@psql $(DATABASE_URL) < scripts/init_db.sql

seed: ## Seed database with sample data
	@echo "Seeding database..."
	@psql $(DATABASE_URL) < scripts/seed_data.sql

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

dev: docker-up run ## Start development environment