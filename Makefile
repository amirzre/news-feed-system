# Variables
APP_NAME := news-feed
BINARY_NAME := main
MAIN_PATH := ./cmd/server
MIGRATION_PATH := ./migrations
DATABASE_URL := postgres://postgres:postgres@localhost:5432/news_feed?sslmode=disable

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Default target
all: help

## Help
.PHONY: help
help: ## Show this help message
	@echo "${BLUE}$(APP_NAME) - Development Commands${NC}"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make ${YELLOW}<target>${NC}\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  ${GREEN}%-15s${NC} %s\n", $$1, $$2 }' $(MAKEFILE_LIST)


## Development
.PHONY: dev
dev: ## Run the application in development mode with hot reload (requires air)
	@echo "${GREEN}Starting development server with hot reload...${NC}"
	air

.PHONY: run
run: ## Run the application directly
	@echo "${GREEN}Running application...${NC}"
	go run $(MAIN_PATH)


.PHONY: build
build: ## Build the application
	@echo "${GREEN}Building application...${NC}"
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: build-linux
build-linux: ## Build for Linux
	@echo "${GREEN}Building for Linux...${NC}"
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux $(MAIN_PATH)

.PHONY: build-windows
build-windows: ## Build for Windows
	@echo "${GREEN}Building for Windows...${NC}"
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows.exe $(MAIN_PATH)

.PHONY: build-mac
build-mac: ## Build for macOS
	@echo "${GREEN}Building for macOS...${NC}"
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-mac $(MAIN_PATH)

build-all: build-linux build-windows build-mac ## Build for all platforms


## Code Quality
.PHONY: format
format: ## Format code with gofmt
	@echo "${GREEN}Formatting code...${NC}"
	gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	@echo "${GREEN}Running go vet...${NC}"
	go vet ./...

.PHONY: security-check
security-check: ## Run gosec security scanner
	@echo "${GREEN}Running security check...${NC}"
	gosec ./...

code-check: format vet ## Run all code quality checks


## Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "${GREEN}Building Docker image...${NC}"
	docker build -t $(APP_NAME) .

.PHONY: docker-dev
docker-dev: ## Run development environment with docker compose
	@echo "${GREEN}Starting development environment...${NC}"
	docker compose -f docker-compose.dev.yml up -d

.PHONY: docker-dev-stop
docker-dev-stop: ## Stop Docker dev containers
	@echo "${YELLOW}Stopping Docker containers...${NC}"
	docker compose -f docker-compose.dev.yml down

.PHONY: docker-prod
docker-prod: ## Run production environment with docker compose
	@echo "${GREEN}Starting production environment...${NC}"
	docker compose up --build -d

.PHONY: docker-stop
docker-stop: ## Stop Docker containers
	@echo "${YELLOW}Stopping Docker containers...${NC}"
	docker compose down


## Database Migrations
.PHONY: migration-create
migration-create: ## Create a new migration file (usage: make migration-create name=create_users_table)
	@if [ -z "$(name)" ]; then \
		echo "${RED}Error: name parameter is required${NC}"; \
		echo "${YELLOW}Usage: make migration-create name=create_users_table${NC}"; \
		exit 1; \
	fi
	@echo "${GREEN}Creating migration: $(name)${NC}"
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq $(name)

.PHONY: migration-up
migration-up: ## Run all pending migrations
	@echo "${GREEN}Running migrations up...${NC}"
	migrate -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" up

.PHONY: migration-down
migration-down: ## Rollback the last migration
	@echo "${YELLOW}Rolling back last migration...${NC}"
	migrate -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" down 1

.PHONY: migration-down-all
migration-down-all: ## Rollback all migrations (DANGER!)
	@echo "${RED}WARNING: This will rollback ALL migrations!${NC}"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		migrate -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" down -all; \
	else \
		echo "Cancelled."; \
	fi

.PHONY: migration-version
migration-version: ## Show current migration version
	@echo "${GREEN}Current migration version:${NC}"
	migrate -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" version

.PHONY: migration-drop
migration-drop: ## Drop everything in database (DANGER!)
	@echo "${RED}WARNING: This will DROP ALL TABLES!${NC}"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		migrate -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" drop; \
	else \
		echo "Cancelled."; \
	fi

# Tests
test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -short ./internal/handler ./internal/service ./internal/repository
