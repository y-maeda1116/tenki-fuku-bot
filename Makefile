.PHONY: build run test test-coverage clean fmt lint help

APP_NAME := tenki-fuku-bot
BIN_DIR := bin

# --- Build ---

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/cli

# --- Run ---

run:
	@echo "Running $(APP_NAME)..."
	@go run ./cmd/cli

# --- Test ---

test:
	@echo "Running tests..."
	@go test -v ./internal/...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./internal/...
	@go tool cover -html=coverage.out -o coverage.html

# --- Format / Lint ---

fmt:
	@go fmt ./...

lint:
	@echo "Running linter..."
	@golangci-lint run ./internal/...

# --- Clean ---

clean:
	@rm -rf $(BIN_DIR) coverage.out coverage.html

# --- Help ---

help:
	@echo "Available targets:"
	@echo "  build          - Build the bot"
	@echo "  run            - Run the bot"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  fmt            - Format Go code"
	@echo "  lint           - Run linter"
	@echo "  clean          - Remove build artifacts"
