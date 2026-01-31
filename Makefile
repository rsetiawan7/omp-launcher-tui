.PHONY: help build build-all build-linux build-darwin build-windows run clean install test lint fmt vet

# Build variables
BINARY_NAME := omp-tui
VERSION := 0.1.0
BUILD_DIR := ./bin
MAIN_PKG := ./cmd/omp-tui
UNAME_S := $(shell uname -s)
GOFLAGS := -v
CGO_ENABLED := 0

# Platform-specific settings
ifeq ($(UNAME_S),Linux)
    PLATFORM := linux
    ARCH := amd64
    OUTPUT := $(BUILD_DIR)/$(BINARY_NAME)-$(PLATFORM)-$(ARCH)
endif
ifeq ($(UNAME_S),Darwin)
    PLATFORM := darwin
    ARCH := $(shell uname -m)
    ifeq ($(ARCH),arm64)
        ARCH_OUTPUT := arm64
    else
        ARCH_OUTPUT := amd64
    endif
    OUTPUT := $(BUILD_DIR)/$(BINARY_NAME)-$(PLATFORM)-$(ARCH_OUTPUT)
endif
ifeq ($(OS),Windows_NT)
    PLATFORM := windows
    ARCH := amd64
    OUTPUT := $(BUILD_DIR)/$(BINARY_NAME)-$(PLATFORM)-$(ARCH).exe
endif

help:
	@echo "$(BINARY_NAME) Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build              Build for current platform"
	@echo "  make build-all          Build for all platforms (Linux, macOS, Windows)"
	@echo "  make build-linux        Build for Linux (amd64)"
	@echo "  make build-darwin       Build for macOS (amd64 + arm64)"
	@echo "  make build-windows      Build for Windows (amd64)"
	@echo "  make run                Build and run the application"
	@echo "  make install            Install binary to GOBIN or \$$HOME/go/bin"
	@echo "  make clean              Remove build artifacts"
	@echo "  make test               Run tests"
	@echo "  make lint               Run linter (golangci-lint)"
	@echo "  make fmt                Format code with gofmt"
	@echo "  make vet                Run go vet"
	@echo "  make help               Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION                 Version string (default: $(VERSION))"
	@echo "  BUILD_DIR               Output directory (default: $(BUILD_DIR))"
	@echo "  CGO_ENABLED             Enable CGO (default: 0 for static binary)"

## Build targets
build: $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) v$(VERSION) for $(PLATFORM)/$(ARCH)..."
	@CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) -o $(OUTPUT) $(MAIN_PKG)
	@echo "✓ Built: $(OUTPUT)"

build-all: build-linux build-darwin build-windows
	@echo "✓ Build complete for all platforms"

build-linux: $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) v$(VERSION) for Linux (amd64)..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PKG)
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

build-darwin: $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) v$(VERSION) for macOS (amd64 + arm64)..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PKG)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PKG)
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64"
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64"

build-windows: $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) v$(VERSION) for Windows (amd64)..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PKG)
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

## Run and install targets
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(OUTPUT)

install: build
	@echo "Installing $(BINARY_NAME)..."
	@install -m 755 $(OUTPUT) $${GOBIN:-$$HOME/go/bin}/$(BINARY_NAME)
	@echo "✓ Installed to $${GOBIN:-$$HOME/go/bin}/$(BINARY_NAME)"

## Code quality targets
test:
	@echo "Running tests..."
	@go test -v ./...

lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1)
	@golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

## Cleanup
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "✓ Clean complete"

deps:
	@echo "Downloading and verifying dependencies..."
	@go mod download
	@go mod verify
	@echo "✓ Dependencies verified"

tidy:
	@echo "Tidying go.mod..."
	@go mod tidy
	@echo "✓ go.mod tidied"

## Development helpers
dev-setup: deps
	@echo "Setting up development environment..."
	@command -v golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✓ Development environment ready"

watch:
	@echo "Watching for changes and rebuilding..."
	@while true; do \
		inotifywait -e modify -r ./internal ./cmd 2>/dev/null && make build; \
	done

## Release helpers
release: build-all
	@echo "Release artifacts ready in $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/
