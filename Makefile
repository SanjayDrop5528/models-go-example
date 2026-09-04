# ==============================================================================
# Makefile for models-go-example
# Dynamic Model & Schema Migration Engine Example Application
# ==============================================================================

.PHONY: help run run-all run-all-server run-server run-memory run-memory-server run-postgres run-postgres-server run-mysql run-mysql-server run-mongodb run-mongodb-server test build clean swagger deps

# Default target
.DEFAULT_GOAL := help

# Colors for terminal output
BLUE  := \033[36m
GREEN := \033[32m
RESET := \033[0m

help: ## Show help instructions for all Makefile targets
	@echo "$(BLUE)Available Makefile Targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-24s$(RESET) %s\n", $$1, $$2}'

deps: ## Download and tidy Go dependencies
	@echo "$(BLUE)Tidying dependencies...$(RESET)"
	go mod tidy

run: run-memory ## Default run target: runs the in-memory adapter example

run-all: run-memory run-postgres run-mysql run-mongodb ## Run all adapter examples sequentially for testing

run-all-server: ## Run all adapter servers concurrently in the background with Swagger UI
	@echo "$(BLUE)Starting all Adapter Servers simultaneously...$(RESET)"
	@echo "$(GREEN)  1. Fiber Engine Server:    http://localhost:8080/swagger/$(RESET)"
	@echo "$(GREEN)  2. PostgreSQL Server:      http://localhost:8081/swagger/$(RESET)"
	@echo "$(GREEN)  3. MySQL Server:           http://localhost:8082/swagger/$(RESET)"
	@echo "$(GREEN)  4. MongoDB Server:         http://localhost:8083/swagger/$(RESET)"
	@echo "$(GREEN)  5. In-Memory Server:       http://localhost:8084/swagger/$(RESET)"
	@trap 'kill 0' EXIT INT TERM; \
	SERVER=true go run ./examples/memory_example & \
	SERVER=true go run ./examples/postgres_example & \
	SERVER=true go run ./examples/mysql_example & \
	SERVER=true go run ./examples/mongodb_example & \
	go run ./cmd/server & \
	wait


run-server: ## Run the main Fiber API server with Swagger UI (port 8080)
	@echo "$(BLUE)Starting Dynamic Model Engine Server...$(RESET)"
	go run ./cmd/server

run-memory: ## Run the complete In-Memory adapter capabilities demo
	@echo "$(BLUE)Running In-Memory Adapter Example...$(RESET)"
	go run ./examples/memory_example

run-memory-server: ## Run the In-Memory adapter demo with Swagger UI server running
	@echo "$(BLUE)Running In-Memory Adapter Example with Swagger Server...$(RESET)"
	SERVER=true go run ./examples/memory_example

run-postgres: ## Run the PostgreSQL adapter capabilities demo
	@echo "$(BLUE)Running PostgreSQL Adapter Example...$(RESET)"
	go run ./examples/postgres_example

run-postgres-server: ## Run the PostgreSQL adapter demo with Swagger UI server running
	@echo "$(BLUE)Running PostgreSQL Adapter Example with Swagger Server...$(RESET)"
	SERVER=true go run ./examples/postgres_example

run-mysql: ## Run the MySQL adapter capabilities demo
	@echo "$(BLUE)Running MySQL Adapter Example...$(RESET)"
	go run ./examples/mysql_example

run-mysql-server: ## Run the MySQL adapter demo with Swagger UI server running
	@echo "$(BLUE)Running MySQL Adapter Example with Swagger Server...$(RESET)"
	SERVER=true go run ./examples/mysql_example

run-mongodb: ## Run the MongoDB adapter capabilities demo
	@echo "$(BLUE)Running MongoDB Adapter Example...$(RESET)"
	go run ./examples/mongodb_example

run-mongodb-server: ## Run the MongoDB adapter demo with Swagger UI server running
	@echo "$(BLUE)Running MongoDB Adapter Example with Swagger Server...$(RESET)"
	SERVER=true go run ./examples/mongodb_example

test: ## Run unit tests across all example packages
	@echo "$(BLUE)Running tests...$(RESET)"
	go test -v ./...

build: ## Build server and example binaries into bin/
	@echo "$(BLUE)Building binaries...$(RESET)"
	@mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/memory_example ./examples/memory_example
	go build -o bin/postgres_example ./examples/postgres_example
	go build -o bin/mysql_example ./examples/mysql_example
	go build -o bin/mongodb_example ./examples/mongodb_example
	@echo "$(GREEN)Binaries built successfully in bin/$(RESET)"

swagger: ## Regenerate Swagger API documentation using swag CLI
	@echo "$(BLUE)Generating Swagger documentation...$(RESET)"
	swag init -g cmd/server/main.go -o docs

clean: ## Remove built binaries and temporary files
	@echo "$(BLUE)Cleaning build output...$(RESET)"
	rm -rf bin
