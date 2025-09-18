# Aktuell Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop examples deps

# Go environment configuration - dynamically detect Go installation
DETECTED_GO_BINARY := $(shell which go)

# Variables
BINARY_NAME=aktuell
DOCKER_IMAGE=aktuell:latest

# Default target
help: ## Display this help message
	@echo "Aktuell - Real-time MongoDB synchronization"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)


deps: ## Download Go dependencies
	@echo "Using Go: $$(go version)"
	go mod download
	go mod tidy

build: deps ## Build the server binary
	@echo "Building Aktuell server..."
	go build -o $(BINARY_NAME) ./cmd/server
	@echo "Build complete: $(BINARY_NAME)"

run: build ## Run the server locally
	./$(BINARY_NAME)

# Testing targets
test: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -short ./pkg/...

test-unit: test ## Alias for test

test-integration: ## Run integration tests (requires MongoDB)
	@echo "Running integration tests..."
	@echo "Note: This requires MongoDB to be running on localhost:27017"
	INTEGRATION_TESTS=1 go test -v -tags=integration ./tests/... -timeout=60s

test-all: ## Run all tests (unit + integration)
	@echo "Running all tests..."
	go test -v -short ./pkg/...
	@echo ""
	@echo "Running integration tests..."
	INTEGRATION_TESTS=1 go test -v -tags=integration ./tests/...

test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -cover ./pkg/...
	go test -coverprofile=coverage.out ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-cover-all: ## Run all tests (unit + integration) with coverage
	@echo "Running all tests with coverage..."
	go test -coverprofile=coverage-unit.out ./pkg/...
	INTEGRATION_TESTS=1 go test -tags integration -coverprofile=coverage-integration.out -coverpkg=./pkg/... ./tests/...
	@echo "Merging coverage profiles..."
	echo "mode: set" > coverage-merged.out
	tail -n +2 coverage-unit.out >> coverage-merged.out
	tail -n +2 coverage-integration.out >> coverage-merged.out
	go tool cover -html=coverage-merged.out -o coverage-all.html
	@echo "Combined coverage report generated: coverage-all.html"
	@echo "Unit test coverage: coverage-unit.out"
	@echo "Integration test coverage: coverage-integration.out"
	@echo "Combined coverage: coverage-merged.out"

test-cover-integration: ## Run integration tests with coverage only
	@echo "Running integration tests with coverage..."
	@echo "Note: This requires MongoDB to be running on localhost:27017"
	INTEGRATION_TESTS=1 go test -tags integration -coverprofile=coverage-integration.out -coverpkg=./pkg/... -v ./tests/...
	go tool cover -html=coverage-integration.out -o coverage-integration.html
	@echo "Integration test coverage report generated: coverage-integration.html"

test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./pkg/...

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	go test -race -short ./pkg/...

test-clean: ## Clean test cache and artifacts
	go clean -testcache
	rm -f coverage.out coverage.html coverage-unit.out coverage-integration.out coverage-merged.out coverage-all.html coverage-integration.html

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	go clean

# Docker targets
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run with Docker Compose
	docker-compose up -d

docker-stop: ## Stop Docker Compose services
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

# MongoDB quick operations
mongo-insert: ## Insert a new random user into MongoDB
	@echo "Inserting a new user into MongoDB..."
	docker exec aktuell-db mongosh aktuell --eval "db.users.insertOne({name: 'User $$(date +%s)', email: 'user$$(date +%s)@example.com', age: $$((RANDOM % 50 + 20)), salary: Math.floor(Math.random() * 80000) + 40000, loginCount: Math.floor(Math.random() * 100) + 1, status: ['active', 'inactive', 'pending'][Math.floor(Math.random() * 3)], skills: [['JavaScript', 'React'], ['Python', 'Django'], ['Go', 'MongoDB'], ['Java', 'Spring'], ['Node.js', 'Express']][Math.floor(Math.random() * 5)], department: ['Engineering', 'Marketing', 'Sales', 'HR', 'Finance'][Math.floor(Math.random() * 5)], createdAt: new Date()})" --quiet

mongo-update: ## Update a random user in MongoDB
	@echo "Updating a random user in MongoDB..."
	docker exec aktuell-db mongosh aktuell --eval "db.users.updateOne({}, {\$$set: {lastUpdated: new Date(), status: 'updated', salary: Math.floor(Math.random() * 50000) + 50000}})" --quiet

mongo-delete: ## Delete one user from MongoDB
	@echo "Deleting one user from MongoDB..."
	docker exec aktuell-db mongosh aktuell --eval "db.users.deleteOne({})" --quiet

mongo-count: ## Count users in MongoDB
	@echo "Counting users in MongoDB..."
	@docker exec aktuell-db mongosh aktuell --eval "db.users.countDocuments()" --quiet

mongo-list: ## List all users in MongoDB
	@echo "Listing all users in MongoDB..."
	@docker exec aktuell-db mongosh aktuell --eval "db.users.find().pretty()" --quiet

mongo-clear: ## Clear all users from MongoDB
	@echo "Clearing all users from MongoDB..."
	@echo "⚠️  This will delete ALL users!"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	docker exec aktuell-db mongosh aktuell --eval "db.users.deleteMany({})" --quiet
	@echo "✅ All users deleted!"

mongo-demo: ## Run comprehensive MongoDB demo operations
	@echo "Running comprehensive MongoDB demo..."
	./scripts/docker-mongo-operations.sh

mongo-status: ## Show MongoDB connection status
	@echo "Checking MongoDB status..."
	@docker exec aktuell-db mongosh --eval "db.adminCommand('hello')" --quiet >/dev/null 2>&1 && echo "✅ MongoDB is running and accessible" || echo "❌ MongoDB is not accessible"

# Example targets
examples: build ## Build example applications
	@echo "Building example applications..."
	sh -c "cd examples/client && go build -o client main.go"
	sh -c "cd examples/generator && go build -o generator main.go"
	sh -c "cd examples/validation && go build -o validation main.go"
	sh -c "cd examples/snapshot && go build -o snapshot main.go"
	@echo "Example binaries created:"
	@echo "  examples/client/client - Aktuell client example"
	@echo "  examples/generator/generator - Data generator for testing"
	@echo "  examples/validation/validation - Subscription validation test"
	@echo "  examples/snapshot/snapshot - Snapshot functionality test"

run-client: examples ## Run the example client
	cd examples/client && ./client

run-generator: examples ## Run the data generator
	cd examples/generator && ./generator

test-validation: examples ## Build and run subscription validation test
	@echo "Running subscription validation test..."
	cd examples/validation && ./validation

test-snapshot: examples ## Build and run snapshot functionality test
	@echo "Running snapshot test..."
	cd examples/snapshot && ./snapshot

# Development targets
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	@if ! command -v docker >/dev/null 2>&1; then \
		echo "Error: Docker is required for development"; \
		exit 1; \
	fi
	@if ! command -v docker-compose >/dev/null 2>&1; then \
		echo "Error: Docker Compose is required for development"; \
		exit 1; \
	fi
	docker-compose up -d mongodb
	@echo "Waiting for MongoDB to be ready..."
	@sleep 10
	@echo "Development environment ready!"

dev-run: dev-setup build ## Run in development mode with MongoDB
	./$(BINARY_NAME)

# Utility targets
format: ## Format Go code
	go fmt ./...

lint: ## Run golint
	@if command -v golint >/dev/null 2>&1; then \
		golint ./...; \
	else \
		echo "golint not installed, skipping..."; \
	fi

vet: ## Run go vet
	go vet ./...

check: format vet lint test ## Run all checks

# Installation targets
install: build ## Install binary to GOPATH/bin
	go install ./cmd/server

uninstall: ## Remove binary from GOPATH/bin
	rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)

.DEFAULT_GOAL := help