# Makefile Summary

A comprehensive Makefile has been created for the `omp-launcher-tui` project with the following features:

## Quick Reference

```bash
# Common commands
make help                 # Show all available targets
make build               # Build for current platform
make run                 # Build and run the app
make clean              # Remove build artifacts
make build-all          # Build for Linux, macOS, Windows
```

## All Available Targets

### Build Targets
- `make build` - Build for current platform
- `make build-all` - Build for all platforms (Linux, macOS, Windows)
- `make build-linux` - Build for Linux (amd64)
- `make build-darwin` - Build for macOS (amd64 + arm64 universal)
- `make build-windows` - Build for Windows (amd64)

### Run & Install
- `make run` - Build and run the application
- `make install` - Install binary to `$GOBIN` or `$HOME/go/bin`

### Code Quality
- `make test` - Run tests
- `make lint` - Run golangci-lint
- `make fmt` - Format code with gofmt
- `make vet` - Run go vet static analysis

### Dependency Management
- `make deps` - Download and verify dependencies
- `make tidy` - Tidy go.mod
- `make dev-setup` - Install development tools

### Cleanup & Release
- `make clean` - Remove all build artifacts
- `make release` - Build all platforms and show release artifacts
- `make watch` - Watch for changes and rebuild (inotify-tools required)

### Help
- `make help` - Display help message

## Environment Variables

Customize builds with:
```bash
VERSION=1.0.0 make build-all        # Custom version
BUILD_DIR=./dist make build         # Custom output directory
CGO_ENABLED=1 make build            # Enable CGO
GOFLAGS="-x" make build             # Custom Go flags
```

## Features

✓ **Auto-detects platform** - Correctly builds for Linux, macOS, or Windows
✓ **Cross-compilation** - Can build for all platforms from any platform
✓ **Static binaries** - CGO disabled by default for portability
✓ **Organized output** - All binaries in `./bin/` directory
✓ **Development tools** - Includes fmt, vet, test, lint targets
✓ **Installation support** - `make install` adds to PATH
✓ **Clean interface** - Simple, memorable target names
✓ **Comprehensive help** - `make help` shows all options

## Build Output

After `make build-all`, binaries are in `./bin/`:
```
bin/
├── omp-tui-darwin-amd64       (8.9 MB - macOS Intel)
├── omp-tui-darwin-arm64       (8.9 MB - macOS Apple Silicon)
├── omp-tui-linux-amd64        (8.8 MB - Linux x86_64)
└── omp-tui-windows-amd64.exe  (8.7 MB - Windows x86_64)
```

## Usage Examples

### Daily Development
```bash
make fmt              # Format code
make vet             # Check for issues
make test            # Run tests
make run             # Build and run
```

### Preparing a Release
```bash
make clean           # Clean old builds
make build-all       # Build for all platforms
ls -lh bin/         # List release artifacts
```

### Cross-Platform Testing
```bash
make build-linux && ./bin/omp-tui-linux-amd64
make build-darwin && ./bin/omp-tui-darwin-arm64
```

### Installing Globally
```bash
make install
omp-tui              # Now available from anywhere
```

## Files Created

- **Makefile** - Main build/run automation script (148 lines)
- **MAKEFILE.md** - Detailed usage guide with examples

## Notes

- Makefile uses standard Unix make syntax (compatible with GNU Make)
- All targets are properly phony (won't conflict with real files)
- Builds are deterministic and reproducible
- Static binaries require no runtime dependencies (except Wine/Proton)
- Version can be changed: `VERSION=0.2.0 make build-all`
