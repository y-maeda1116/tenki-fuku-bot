.PHONY: build build-cli build-desktop run-cli run-desktop clean test test-coverage test-race mocks lint fmt security-check help

APP_NAME := myapp
BIN_DIR := bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/y-maeda1116/template-go-cross/internal/version.Version=$(VERSION)"

CLI_MAIN := ./cmd/cli
DESKTOP_DIR := ./cmd/desktop
FRONTEND_DIR := $(DESKTOP_DIR)/frontend
TEST_PKGS := ./internal/... ./cmd/app/... ./cmd/cli/...

# Detect OS
ifeq ($(OS),Windows_NT)
    EXE_EXT := .exe
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        EXE_EXT :=
    endif
    ifeq ($(UNAME_S),Darwin)
        EXE_EXT :=
    endif
endif

# --- Build ---

build-cli:
	@echo "Building CLI for current OS..."
	@mkdir -p $(BIN_DIR)
	@go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)$(EXE_EXT) $(CLI_MAIN)

build-desktop:
	@echo "Building Desktop for current OS..."
	@cd $(FRONTEND_DIR) && npm install && npm run build
	@cd $(DESKTOP_DIR) && wails build -clean -tags webview2

build-all: build-cli build-desktop

# --- Run ---

run-cli:
	@echo "Running CLI..."
	@go run $(CLI_MAIN) $(ARGS)

run-desktop:
	@echo "Running Desktop..."
	@cd $(DESKTOP_DIR) && wails dev

# --- Test ---

test:
	@echo "Running tests..."
	@go test -v $(TEST_PKGS)

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out $(TEST_PKGS)
	@go tool cover -html=coverage.out -o coverage.html

test-race:
	@echo "Running tests with race detector..."
	@go test -race -v $(TEST_PKGS)

# --- Mocks ---

mocks:
	@echo "Generating mocks..."
	@mkdir -p test/mocks
	@mockgen -source=internal/core/service.go -destination=test/mocks/service_mock.go

# --- Lint / Format ---

fmt:
	@go fmt ./...

lint:
	@echo "Installing golangci-lint if needed..."
	@which golangci-lint || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest)
	@echo "Running linter..."
	@golangci-lint run $(TEST_PKGS)

# --- Clean ---

clean:
	@rm -rf $(BIN_DIR) coverage.out coverage.html
	@cd $(DESKTOP_DIR) && rm -rf build/bin

# --- Security ---

security-check: ensure-golangci-lint ensure-govulncheck
	@echo "Running security checks..."
	@echo "  → golangci-lint"
	@$(GOLANGCI_BIN) run $(TEST_PKGS)
	@echo "  → govulncheck"
	@$(GOVULNCHECK_BIN) $(TEST_PKGS)
	@echo "All security checks passed."

GOLANGCI_BIN := $(shell go env GOPATH)/bin/golangci-lint
GOVULNCHECK_BIN := $(shell go env GOPATH)/bin/govulncheck

ensure-golangci-lint:
	@test -x $(GOLANGCI_BIN) || (echo "Installing golangci-lint v2..." && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest)

ensure-govulncheck:
	@test -x $(GOVULNCHECK_BIN) || GOBIN=$(shell go env GOPATH)/bin go install golang.org/x/vuln/cmd/govulncheck@latest

# --- Help ---

help:
	@echo "Available targets:"
	@echo "  build-cli       - Build CLI for current OS"
	@echo "  build-desktop    - Build Desktop for current OS"
	@echo "  build-all        - Build CLI and Desktop"
	@echo "  run-cli          - Run CLI (use ARGS=\"--help\" for options)"
	@echo "  run-desktop      - Run Desktop in dev mode"
	@echo "  test             - Run all tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  test-race        - Run tests with race detector"
	@echo "  mocks            - Generate mocks"
	@echo "  fmt              - Format Go code"
	@echo "  lint             - Run linter"
	@echo "  security-check   - Run golangci-lint + govulncheck"
	@echo "  clean            - Remove build artifacts"
