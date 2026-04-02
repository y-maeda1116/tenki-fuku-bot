# Go CLI + Desktop Template

A cross-platform Go application template supporting both CLI and Desktop (Wails) interfaces.

## Features

- **Dual Interface:** CLI (Cobra) and Desktop (Wails + React)
- **Shared Core:** Business logic shared between CLI and Desktop
- **Structured Logging:** zap-based logging with multiple levels
- **Configuration:** Viper-based config with YAML/TOML/JSON support
- **Hot Reload:** Air for CLI, Wails dev for Desktop
- **Testing:** mockgen with 80%+ coverage goal
- **CI/CD:** GitHub Actions for testing and cross-platform building

## Project Structure

```
.
├── cmd/
│   ├── app/
│   │   └── main.go           # Background app (signal handling)
│   ├── cli/
│   │   └── main.go           # CLI entry point (Cobra)
│   └── desktop/
│       ├── main.go           # Desktop entry point (Wails)
│       ├── wails.json        # Wails configuration
│       └── frontend/         # React + TypeScript frontend
├── internal/
│   ├── cli/                  # CLI command definitions
│   ├── config/               # Configuration (Viper)
│   ├── core/                 # Shared business logic
│   ├── logger/               # Logging (zap)
│   ├── ui/                   # Wails UI bindings
│   └── version/              # Version injected at build time
├── test/
│   └── mocks/                # Generated mocks
├── .github/workflows/        # CI/CD
├── Makefile                  # Build targets
├── config.yaml.example       # Configuration template
├── env.example               # Environment variables template
├── air.toml                  # Hot reload (CLI)
└── go.mod                    # Go module definition
```

## Prerequisites

- Go 1.26 or later
- Node.js 20 or later (for Desktop)
- Make

## Getting Started

```bash
# Clone repository
git clone https://github.com/y-maeda1116/template-go-cross.git
cd template-go-cross

# Install Go dependencies
go mod download

# Install frontend dependencies (for Desktop only)
cd cmd/desktop/frontend && npm install && cd ../..
```

### Configuration

```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your settings
```

## Usage

### CLI

```bash
# Run CLI
make run-cli

# Show help
make run-cli ARGS="--help"

# Say hello
make run-cli ARGS="hello --name World"

# Build CLI
make build-cli
```

### Desktop

```bash
# Run Desktop in dev mode
make run-desktop

# Build Desktop
make build-desktop
```

### Development

```bash
# Run tests (excludes CGO-dependent desktop packages)
make test

# Run tests with coverage
make test-coverage

# Run tests with race detector
make test-race

# Generate mocks
make mocks

# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean
```

## Architecture

```
Application Layer
├── CLI (Cobra)
├── Desktop (Wails + React)
└── Background App (signal handling)
         ↓
Core Business Logic Layer
└── Shared services (internal/core)
         ↓
Infrastructure Layer
├── Config (Viper)
└── Logger (zap)
```

## License

MIT
