.PHONY: build run test clean generate docker-build docker-run lint dev deps tidy tools help

# Application name
APP_NAME := k8s-graph-explorer
VERSION := 1.0.0

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GORUN := $(GOCMD) run
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
BINARY_DIR := bin
BINARY_NAME := $(BINARY_DIR)/$(APP_NAME)

# Build flags
LDFLAGS := -ldflags="-w -s -X main.Version=$(VERSION)"

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application binary
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v ./cmd/server

## run: Run the application
run:
	@echo "Running $(APP_NAME)..."
	$(GORUN) ./cmd/server

## dev: Run with hot reload (requires air)
dev:
	@echo "Running in development mode with hot reload..."
	air

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

## tidy: Tidy go modules
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run --timeout 5m

## fmt: Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## generate: Generate GraphQL code
generate:
	@echo "Generating GraphQL code..."
	$(GORUN) github.com/99designs/gqlgen generate

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) -f ../docker/backend/Dockerfile .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

## docker-run: Run in Docker
docker-run:
	@echo "Running in Docker..."
	docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

## tools: Install development tools
tools:
	@echo "Installing development tools..."
	go install github.com/99designs/gqlgen@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest

## swagger: Generate Swagger docs
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/server/main.go -o docs/swagger

## all: Run all checks and build
all: fmt vet lint test build
	@echo "All checks passed and build complete!"
