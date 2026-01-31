# Makefile Usage Guide

## Quick Start

```bash
# Build for current platform
make build

# Run the application
make run

# Clean build artifacts
make clean
```

## Available Targets

### Building

| Target | Purpose |
|--------|---------|
| `make build` | Build for current platform (Linux/macOS/Windows) |
| `make build-all` | Build for all supported platforms |
| `make build-linux` | Build for Linux (amd64) |
| `make build-darwin` | Build for macOS (amd64 + arm64 universal) |
| `make build-windows` | Build for Windows (amd64) |

**Output Location**: `./bin/` directory

Example:
```bash
$ make build-all
Building omp-tui v0.1.0 for Linux (amd64)...
✓ Built: ./bin/omp-tui-linux-amd64
Building omp-tui v0.1.0 for macOS (amd64 + arm64)...
✓ Built: ./bin/omp-tui-darwin-amd64
✓ Built: ./bin/omp-tui-darwin-arm64
Building omp-tui v0.1.0 for Windows (amd64)...
✓ Built: ./bin/omp-tui-windows-amd64.exe
✓ Build complete for all platforms
```

### Running & Installing

| Target | Purpose |
|--------|---------|
| `make run` | Build and run the application |
| `make install` | Install binary to `$GOBIN` or `$HOME/go/bin` |

Example:
```bash
# Run the app
$ make run
Building omp-tui v0.1.0 for darwin/arm64...
✓ Built: ./bin/omp-tui-darwin-arm64
Running omp-tui...
[TUI opens]

# Install to system
$ make install
Installing omp-tui...
✓ Installed to /Users/username/go/bin/omp-tui
```

### Code Quality

| Target | Purpose |
|--------|---------|
| `make test` | Run all tests |
| `make lint` | Run golangci-lint (requires installation) |
| `make fmt` | Format code with gofmt |
| `make vet` | Run go vet for static analysis |

Example:
```bash
$ make fmt vet
Formatting code...
Running go vet...
```

### Dependency Management

| Target | Purpose |
|--------|---------|
| `make deps` | Download and verify dependencies |
| `make tidy` | Tidy go.mod (remove unused, add missing) |
| `make dev-setup` | Set up development environment |

Example:
```bash
$ make dev-setup
Setting up development environment...
✓ Development environment ready

$ make deps
Downloading and verifying dependencies...
✓ Dependencies verified
```

### Cleanup & Release

| Target | Purpose |
|--------|---------|
| `make clean` | Remove all build artifacts |
| `make release` | Build all platforms and list release artifacts |

Example:
```bash
$ make release
Building omp-tui v0.1.0 for all platforms...
[Build output...]
Release artifacts ready in ./bin/
-rwxr-xr-x  omp-tui-linux-amd64
-rwxr-xr-x  omp-tui-darwin-amd64
-rwxr-xr-x  omp-tui-darwin-arm64
-rwxr-xr-x  omp-tui-windows-amd64.exe
```

### Other Utilities

| Target | Purpose |
|--------|---------|
| `make watch` | Watch for changes and rebuild (requires inotify-tools) |
| `make help` | Display this help message |

## Environment Variables

Customize builds with these variables:

```bash
# Set custom version
make build VERSION=0.2.0

# Set custom output directory
make build BUILD_DIR=./dist

# Enable CGO (disabled by default for static binaries)
make build CGO_ENABLED=1

# Verbose build output
make build GOFLAGS="-v -x"
```

Examples:
```bash
# Build version 1.0.0 with custom output
make build-all VERSION=1.0.0 BUILD_DIR=./release

# Build with verbose logging
make build GOFLAGS="-v -x"
```

## Common Workflows

### Development

```bash
# Set up environment once
make dev-setup

# Work and test
make fmt
make vet
make test
make run
```

### Preparing Release

```bash
# Verify everything
make deps
make fmt
make vet
make test

# Build all platforms
make build-all

# Clean previous builds first
make clean && make build-all

# List release artifacts
ls -lh bin/
```

### Continuous Development

```bash
# Watch for changes (Unix/Linux only)
make watch

# Or manually rebuild on changes
make clean build
```

## Troubleshooting

### "make: command not found"
Install GNU Make:
```bash
# macOS
brew install make

# Linux (usually pre-installed)
sudo apt-get install make

# Windows (use WSL or install via chocolatey)
choco install make
```

### "golangci-lint not installed" (when running lint)
Install the linter:
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Build output in wrong location
Specify `BUILD_DIR`:
```bash
make build BUILD_DIR=./dist
```

### Different build output per platform
The Makefile auto-detects your platform. For cross-compilation, use specific targets:
```bash
make build-linux      # Always builds Linux version
make build-darwin     # Always builds macOS versions
make build-windows    # Always builds Windows version
```

## Binary Locations

After building, binaries are in `./bin/`:

```
bin/
├── omp-tui-darwin-amd64       # macOS Intel
├── omp-tui-darwin-arm64       # macOS Apple Silicon
├── omp-tui-linux-amd64        # Linux x86_64
└── omp-tui-windows-amd64.exe  # Windows x86_64
```

Run directly:
```bash
./bin/omp-tui-darwin-arm64
./bin/omp-tui-linux-amd64
./bin/omp-tui-windows-amd64.exe
```

Or install to your PATH:
```bash
make install
omp-tui  # Now available globally
```
