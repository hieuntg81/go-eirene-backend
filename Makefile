.PHONY: build run test clean migrate-up migrate-down swagger lint fmt help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Binary name
BINARY_NAME=rescue-api
BINARY_UNIX=$(BINARY_NAME)_unix

# Main package path
MAIN_PACKAGE=./cmd/server

# Database
DB_URL ?= postgres://postgres:postgres@localhost:5432/rescue_app?sslmode=disable

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PACKAGE)

# Run the application
run:
	$(GORUN) $(MAIN_PACKAGE)

# Run with hot reload (requires air)
dev:
	air

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run database migrations up
migrate-up:
	migrate -path ./migrations -database "$(DB_URL)" up

# Run database migrations down
migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir ./migrations -seq $$name

# Generate Swagger documentation
swagger:
	swag init -g cmd/server/main.go -o docs

# Run linter
lint:
	golangci-lint run ./...

# Format code
fmt:
	$(GOFMT) -s -w .

# Vet code
vet:
	$(GOCMD) vet ./...

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v $(MAIN_PACKAGE)

# Install development tools
install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Setup development environment
setup: deps install-tools
	cp .env.example .env
	@echo "Development environment setup complete!"
	@echo "Please edit .env with your configuration"

# Help
help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make dev            - Run with hot reload (requires air)"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make deps           - Download dependencies"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make migrate-create - Create a new migration"
	@echo "  make swagger        - Generate Swagger documentation"
	@echo "  make lint           - Run linter"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Vet code"
	@echo "  make build-linux    - Build for Linux"
	@echo "  make install-tools  - Install development tools"
	@echo "  make setup          - Setup development environment"

# Default target
.DEFAULT_GOAL := help
