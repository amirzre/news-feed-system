# Variables
APP_NAME := news-feed
BINARY_NAME := main
MAIN_PATH := ./cmd/server

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
