.PHONY: up down restart build logs ps clean help

# Default target
.DEFAULT_GOAL := help

## setup: Initialize the project (copy .env, install dependencies)
setup:
	@echo "Setting up project..."
	@if [ ! -f .env ]; then cp .env.example .env; echo "Created .env"; fi
	@if [ ! -f backend/.env ]; then cp backend/.env.example backend/.env; echo "Created backend/.env"; fi
	@echo "Installing dashboard dependencies..."
	cd dashboard && npm install
	@echo "Downloading backend dependencies..."
	cd backend && go mod download
	@echo "Setup complete! You can now run 'make up' to start the project."

## up: Start the entire project using docker-compose
up:
	docker-compose up -d

## down: Stop the entire project
down:
	docker-compose down

## restart: Restart the entire project
restart: down up

## build: Build all docker images
build:
	docker-compose build

## logs: View logs from all services
logs:
	docker-compose logs -f

## ps: List running containers
ps:
	docker-compose ps

## dev-backend: Run backend in development mode (requires local Neo4j or docker-compose up neo4j)
dev-backend:
	cd backend && make dev

## dev-dashboard: Run dashboard in development mode
dev-dashboard:
	cd dashboard && npm run dev

## dev: Run both backend and dashboard concurrently
dev:
	./scripts/dev.sh

## up-deps: Start only the dependencies (Neo4j)
up-deps:
	docker-compose up -d neo4j

## clean: Remove all containers and volumes
clean:
	docker-compose down -v
	cd backend && make clean

## help: Show this help message
help:
	@echo "K8S Graph Explorer - Management Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //g' | column -t -s ':' | sed -e 's/^/  /'
