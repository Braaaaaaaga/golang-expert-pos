DOCKER_COMPOSE = docker-compose
GO = go

.PHONY: help build run test clean up down logs stress-test

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	$(GO) build -o bin/rate-limiter ./main.go

run: ## Run the application locally
	$(GO) run main.go

test: ## Run tests
	$(GO) test -v ./tests/...

test-coverage: ## Run tests with coverage
	$(GO) test -v -coverprofile=coverage.out ./tests/...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

up: ## Start services with docker-compose
	$(DOCKER_COMPOSE) up -d

down: ## Stop services
	$(DOCKER_COMPOSE) down

logs: ## Show logs
	$(DOCKER_COMPOSE) logs -f rate-limiter

stress-test: ## Run stress test (server must be running)
	@echo "Running stress test against http://localhost:8080"
	@echo "Make sure the server is running (make up)"
	./stress-test.sh

docker-build: ## Build Docker image
	docker build -t rate-limiter .

dev: ## Start development environment
	$(DOCKER_COMPOSE) up -d redis
	@echo "Redis started. Run 'make run' to start the application"

install: ## Install dependencies
	$(GO) mod tidy
	$(GO) mod download

format: ## Format code
	$(GO) fmt ./...
	$(GO) mod tidy

lint: ## Run linter
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Example usage commands
example-requests: ## Show example API requests
	@echo "Example requests to test the rate limiter:"
	@echo ""
	@echo "1. Normal request (IP-based limiting):"
	@echo "   curl http://localhost:8080/test"
	@echo ""
	@echo "2. Request with API token:"
	@echo "   curl -H \"API_KEY: abc123\" http://localhost:8080/test"
	@echo ""
	@echo "3. Health check:"
	@echo "   curl http://localhost:8080/health"
	@echo ""
	@echo "4. Test different limits:"
	@echo "   curl -H \"API_KEY: def456\" http://localhost:8080/test  # 50 req/s"
	@echo "   curl -H \"API_KEY: xyz789\" http://localhost:8080/test  # 200 req/s"