# Quick Start Guide

## Building & Running

### Simplest Way (using Makefile)
```bash
# Build for your current platform
make build

# Run the application
make run

# Or in one command:
make build && ./bin/omp-tui-*  # (matches your platform)
```

### Using the Build Script
```bash
./build.sh
```

### Manual Go Commands
```bash
# Build
CGO_ENABLED=0 go build -o omp-tui ./cmd/omp-tui

# Run
CGO_ENABLED=0 go run ./cmd/omp-tui
```

## Building for All Platforms

```bash
# Using Makefile (recommended)
make build-all

# Using build.sh script
./build.sh

# Manual cross-compilation
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/omp-tui-linux-amd64 ./cmd/omp-tui
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/omp-tui-darwin-amd64 ./cmd/omp-tui
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/omp-tui-darwin-arm64 ./cmd/omp-tui
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/omp-tui-windows-amd64.exe ./cmd/omp-tui
```

## Installing to Your System

```bash
# Install to $GOBIN or $HOME/go/bin
make install

# Then run from anywhere:
omp-tui

# Or manually:
sudo cp bin/omp-tui-* /usr/local/bin/omp-tui
chmod +x /usr/local/bin/omp-tui
omp-tui
```

## Makefile Targets Reference

| Command | Description |
|---------|-------------|
| `make help` | Show all available targets |
| `make build` | Build for current platform |
| `make build-all` | Build for Linux, macOS, Windows |
| `make build-linux` | Build Linux binary |
| `make build-darwin` | Build macOS binaries (Intel + ARM) |
| `make build-windows` | Build Windows binary |
| `make run` | Build and run |
| `make install` | Install to PATH |
| `make clean` | Remove all build artifacts |
| `make fmt` | Format code |
| `make vet` | Check code with go vet |
| `make test` | Run tests |
| `make lint` | Run linter |
| `make deps` | Download dependencies |
| `make tidy` | Tidy go.mod |
| `make dev-setup` | Set up dev environment |

## Key Features

- **Version Filtering** - Press `V` to filter by 0.3.7, 0.3.DL, or open.mp
- **Smart Caching** - Server data cached with 24-hour validity
- **Debounced Updates** - 500ms delay before querying selected server
- **Search & Filter** - Press `/` to search, filters shown in bottom panel
- **Manual Refresh** - Press `R` to force fetch fresh server data

## Project Layout

```
omp-launcher-tui/
├── Makefile                 # Build automation (use this!)
├── build.sh                 # Alternative build script
├── README.md               # Full documentation
├── MAKEFILE.md             # Detailed Makefile guide
├── MAKEFILE_SUMMARY.md     # Quick reference
├── servers.json            # Fallback server list
├── go.mod & go.sum         # Dependencies
├── cmd/
│   └── omp-tui/
│       └── main.go         # Entry point
├── internal/
│   ├── config/             # Config & favorites
│   ├── server/             # Server querying & caching
│   ├── launcher/           # Launch execution
│   └── tui/               # UI implementation
└── bin/                    # Built binaries (created by make build)
    ├── omp-tui-darwin-amd64
    ├── omp-tui-darwin-arm64
    ├── omp-tui-linux-amd64
    └── omp-tui-windows-amd64.exe
```

## Keybindings Reference

### Main View
| Key | Action |
|-----|--------|
| ↑ ↓ | Navigate servers |
| Enter | Connect to server |
| C | Open configuration |
| / | Search servers |
| R | Refresh server list |
| S | Cycle sort mode |
| F | Switch to Favorites |
| M | Switch to Master list |
| A | Add favorite manually |
| ★ | Toggle favorite |
| D | Delete from favorites |
| P | Enter password |
| Q | Quit |

### Config Modal
| Key | Action |
|-----|--------|
| Ctrl+B | Browse for file/directory |
| Ctrl+T | Test GTA SA path |
| Esc | Close and return |

## Typical Workflows

### Daily Development
```bash
# Start development session
make dev-setup      # One-time setup
make fmt            # Format code
make vet            # Check issues
make run            # Build and test
```

### Code Quality Check
```bash
make fmt vet lint test
```

### Release Preparation
```bash
make clean
make build-all
make install
ls -lh bin/
```

### Install for Users
```bash
make build
sudo cp bin/omp-tui-* /usr/local/bin/omp-tui
chmod +x /usr/local/bin/omp-tui
```

## Customization

```bash
# Custom version string
VERSION=1.0.0 make build-all

# Custom output directory
BUILD_DIR=./release make build

# Enable CGO (disable static binary)
CGO_ENABLED=1 make build

# Verbose build output
GOFLAGS="-v -x" make build
```

## Requirements

- **Go 1.22+** - From [golang.org](https://golang.org)
- **GNU Make** - Usually pre-installed on macOS/Linux
  - macOS: `brew install make`
  - Linux: `sudo apt-get install make`
  - Windows: Use WSL or `choco install make`

## Troubleshooting

**"make: command not found"**
```bash
# Install make
brew install make          # macOS
sudo apt-get install make  # Linux
```

**"no Go installation found"**
```bash
# Install Go 1.22+
# Visit https://golang.org/dl and install
```

**Build succeeds but binary won't run on macOS**
```bash
# This is a signing issue; use:
CGO_ENABLED=0 go run ./cmd/omp-tui

# Or build with:
make build
```

**"permission denied" when running binary**
```bash
# Make it executable
chmod +x bin/omp-tui-*

# Then run
./bin/omp-tui-darwin-arm64
```

## Next Steps

1. Read [README.md](README.md) for full documentation
2. Read [MAKEFILE.md](MAKEFILE.md) for detailed Makefile usage
3. Press `C` in the app to configure settings
4. Set `gta_path` to your GTA: San Andreas directory (use Ctrl+T to test)
5. Add your favorite servers with `A` key or `★` to toggle
6. Run `make run` and enjoy!

## Support

For issues or questions:
- Check [README.md](README.md) troubleshooting section
- See [MAKEFILE.md](MAKEFILE.md) for build issues
- Review the project structure in [internal/](internal/) directory
