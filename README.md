# Open.MP TUI Launcher

> **⚠️ AI-Generated Project Notice**  
> This project was largely generated using AI assistance. While functional, users may encounter unexpected behaviors, bugs, or incomplete features. Use at your own discretion and please report any issues you find.

A cross-platform, terminal-based Open.MP launcher for GTA: San Andreas multiplayer.
Built in Go with tview, designed for SSH, Steam Deck, and low-resource systems.

## Server List Fetching

The launcher fetches servers from the **Open.MP API** (`https://api.open.mp/servers`), which provides a real-time list of all Open.MP servers. Server information includes:

- Server name (hostname)
- Host and port
- Current player count and max players
- Password protection status
- Gamemode and language

## Features

- **Server Browser**: Browse Open.MP servers in a responsive TUI
- **Favorites System**: Save your favorite servers to a separate list with quick toggle
- **Master List & Favorites Views**: Switch between master server list and your favorites
- **Live Server Info**: Real-time updates for selected server (ping, players, rules) with 500ms debounce
- **Ping History Chart**: Visual ASCII chart showing ping history over time
- **Player List**: View online players for the selected server (with SA-MP limitation notice when unavailable)
- **Server Rules**: View server rules in a sorted table format
- **Search & Filter**: 
  - Search servers by name/IP
  - Filter by version (0.3.7, 0.3.DL, open.mp)
  - Combined filter display panel
- **Sort Options**: Sort by ping or player count
- **Smart Caching**: 
  - Cached server data including ping and player counts
  - 24-hour cache validity for automatic refreshes
  - Manual refresh (R key) always fetches fresh data
  - Preserved ping data when merging server lists
- **Password Support**: Securely enter passwords for locked servers (never persisted)
- **Browse-Only Mode**: Optional mode to view servers without connecting (great for streaming/demos)
- **Cross-Platform Launcher**: Automatic Wine/Proton detection on Linux/macOS; native Windows support
- **Persistent Config**: Saves nickname, GTA path, open.mp launcher path to config file
- **Master List Manager**: Add, edit, and manage multiple master server lists
- **File Browser**: Built-in file browser for selecting GTA path and launcher location
- **SSH-Ready**: Works over SSH and on Steam Deck (no GUI dependencies)
- **Static Binary**: Single compiled binary with zero external dependencies (except Wine/Proton at runtime)

## Quick Start

### Prerequisites

- **macOS**: Xcode Command Line Tools (`xcode-select --install`)
- **Linux**: `build-essential` and `pkg-config`
- **Windows**: Download pre-built binary or install Go 1.22+
- **Open.MP Client**: GTA: San Andreas with Open.MP installed
- **Wine/Proton** (Linux/macOS): Auto-detected, optional if using native binary

### Build

```sh
# Clone and navigate
git clone https://github.com/openmultiplayer/omp-launcher-tui
cd omp-launcher-tui

# Fetch dependencies
go mod download

# Build for current platform
CGO_ENABLED=0 go build -o omp-tui ./cmd/omp-tui

# Or use the Makefile
make build
```

### Run

```sh
# Initialize configuration (first time setup)
./omp-tui init

# Run TUI mode (interactive browser)
./omp-tui

# Run CLI mode (direct connection)
./omp-tui connect <alias|host:port>
```

**Init Command:**
```sh
# Initialize configuration and fetch server list
./omp-tui init

# Initialize with GTA path and OMP launcher path
./omp-tui init --gta-path "/path/to/GTA San Andreas" --omp-launcher "/path/to/omp-launcher"

# Initialize with only GTA path
./omp-tui init --gta-path "/path/to/GTA San Andreas"
```

The `init` command will:
- Create configuration directory
- Generate default `config.json` file
- Apply provided GTA path and OMP launcher path (if flags are used)
- Create empty `favorites.json` file
- Generate `master_lists.json` with Open.MP official server list
- Fetch servers from the master list
- Query each server for detailed information
- Save all server data to `servers_cache.json`

This is useful for:
- First-time setup
- Resetting configuration to defaults
- Pre-populating server cache for faster startup
- Automated setup with predefined paths

**Export/Import Commands:**
```sh
# Export configuration, favorites, and master lists to a file
./omp-tui export my-backup.json

# Import configuration from a file
./omp-tui import my-backup.json

# Export to a subdirectory (creates directory if needed)
./omp-tui export backup/config-$(date +%Y%m%d).json
```

Export/Import features:
- Exports all configuration in a single JSON file
- Includes config settings, favorites, and master lists
- Shows summary of exported/imported data
- Import requires confirmation before overwriting
- Useful for backups, migration, and sharing configurations

**CLI Mode Examples:**
```sh
# Connect using alias from favorites
./omp-tui connect my-server

# Connect to a server directly (port defaults to 7777)
./omp-tui connect 127.0.0.1

# Connect with custom port
./omp-tui connect 127.0.0.1:8888

# Connect to a remote server
./omp-tui connect play.example.com:7777
```

**CLI Mode Features:**
- **init**: Initialize configuration and fetch server list
  - Creates all necessary config files
  - Optional `--gta-path` and `--omp-launcher` flags for automated setup
  - Fetches and caches servers from master list
  - Pre-queries servers for detailed information
- **connect**: Direct connection without TUI
  - Supports alias lookup from favorites (faster and easier)
  - Supports host:port format or host only (defaults to port 7777)
  - Automatic server query to check password requirement
  - Password prompt if server is password-protected
  - Uses game path and launcher path from config
  - Helpful error messages guide you to run `init` if config is missing
- **export**: Export configuration, favorites, and master lists to a file
  - Single JSON file containing all settings
  - Useful for backups and migration
- **import**: Import configuration from an exported file
  - Restores config, favorites, and master lists
  - Requires confirmation before overwriting

On macOS, if you get a signing error, use:
```sh
CGO_ENABLED=0 go run ./cmd/omp-tui
```

## Configuration

Configuration files are stored in the application support directory:
- **macOS**: `~/Library/Application Support/omp-tui/`
- **Linux**: `~/.config/omp-tui/`
- **Windows**: `%APPDATA%\omp-tui\`

### Config Files

- `config.json` - Main configuration
- `favorites.json` - Saved favorite servers
- `masterlist.json` - Master server list sources
- `servers_cache.json` - Cached server list (includes ping, players, rules)
  - Updates when servers are queried
  - Used on startup to display servers immediately
  - 24-hour validity for automatic refreshes
  - Manual refresh (R key) always updates cache with fresh data

### Example Config

```json
{
  "nickname": "Player",
  "gta_path": "/path/to/GTA",
  "omp_launcher": "/path/to/omp-launcher",
  "runtime": "auto",
  "master_server": "https://api.open.mp/servers",
  "browse_only": false
}
```

- **nickname**: Your in-game name
- **gta_path**: Path to your GTA: San Andreas installation
- **omp_launcher**: Path to open.mp launcher executable
- **runtime**: `auto` (detect), `wine`, `proton`, or `native` (Windows)
- **master_server**: Open.MP API endpoint (default: `https://api.open.mp/servers`)
- **browse_only**: When `true`, disables server connections (browse/view only mode)

## Keybindings

### Main View

| Key | Action |
|-----|-Search by server name or IP |
| `V` | Filter by version (0.3.7, 0.3.DL, open.mp) |
| `R` | Refresh server list (fetches fresh data)
| `Enter` | Connect to selected server |
| `C` | Open configuration modal |
| `/` | Open search (by server name or IP) |
| `R` | Refresh server list from master |
| `S` | Cycle sort mode (none → ping → players) |
| `F` | Switch to Favorites view |
| `M` | Switch to Master List view |
| `A` | Add server to favorites manually |
| `★` | Toggle favorite for selected server |
| `D` | Remove server from favorites (in Favorites view) |
| `P` | Enter password for locked server |
| `Q` | Quit |

### Configuration Modal

| Key | Action |
|-----|--------|
| `Ctrl+B` | Browse for file/directory |
| `Ctrl+T` | Test GTA SA Path (check if gta_sa.exe exists) |
| `Esc` | Close modal and return to main view |
| `↑` `↓` | Navigate between fields |

## Wine / Proton Setup

### Linux with Wine

```bash
# Wine will be auto-detected
# Set GTA path in configuration
```

### Linux with Proton

```bash
# Proton will be auto-detected
# Set GTA path to your Steam installation
```

### macOS with Wine

Install via Homebrew:
```bash
brew install wine-stable
```

## Fallback Server List

If the Open.MP API is unreachable, the launcher uses cached servers or `servers.json` in the current directory. Update [servers.json](servers.json) with your custom list:

```json
[
  {
    "name": "Community Server",
    "host": "192.168.1.100",
    "port": 7777
  }
]
```

## Troubleshooting

### "No supported runtime found"
Install Wine or Proton, or use the native binary on Windows.

### "Unable to find Open.MP client executable"
Set `gta_path` in config to your GTA: San Andreas folder. Use `Ctrl+T` in the config modal to test if `gta_sa.exe` exists.

### "Player list unavailable (SA-MP limitation)"
Some servers don't expose the player list via the query protocol. This is a SA-MP protocol limitation, not a bug.

### Network errors when fetching servers
Check your internet connectivity. The launcher will use cached servers if the master server is unavailable.

### Passwords not working
Ensure the server is password-protected. Passwords are held in memory only and never written to disk.

### Can't connect to servers / Enter key not working
Check if "Browse Only Mode" is enabled in the configuration (press `C`). When enabled, server connections are disabled for viewing purposes only.

## Project Structure

```
.
├── cmd/
│   └── omp-tui/
│       └── main.go                 # Entry point
├── internal/
│   ├── config/
│   │   ├── config.go               # Config struct and defaults
│   │   ├── favorites.go            # Favorites management
│   │   ├── load.go                 # Load/save from disk
│   │   ├── masterlist.go           # Master list management
│   │   └── paths.go                # Config directory resolution
│   ├── server/
│   │   ├── model.go                # Server data structure
│   │   ├── master.go               # Master server fetch
│   │   ├── query.go                # UDP server query (SA-MP protocol)
│   │   ├── cache.go                # Server list caching
│   │   └── sort.go                 # Server sorting utilities
│   ├── launcher/
│   │   ├── launcher.go             # Launch executable with Wine/Proton
│   │   └── runtime.go              # Runtime detection
│   └── tui/
│       ├── app.go                  # Main app logic and state
│       ├── layout.go               # UI layout with tview
│       ├── keys.go                 # Keybinding help text
│       ├── modals.go               # Search, password, and favorites dialogs
│       ├── filebrowser.go          # Built-in file browser
│       ├── masterlist.go           # Master list manager UI
│       └── update.go               # GitHub update checker
├── go.mod                          # Go module definition
├── go.sum                          # Dependency checksums
├── Makefile                        # Build commands
├── build.sh                        # Multi-platform build script
├── servers.json                    # Fallback server list
└── README.md                       # This file
```
Smart Updates**: 
  - Selected server info updates every second after 500ms debounce
  - Prevents unnecessary queries when quickly browsing servers
- **Intelligent Caching**:
  - Servers cached with ping, players, and rules data
  - 24-hour cache validity for startup refreshes
  - Manual refresh always fetches fresh data
  - Cache merging preserves ping data during server list updates
- **Version Filtering**: Static filter for SA-MP 0.3.7, 0.3.DL, and open.mp versions
## Design Notes

- **No CGO**: Zero external C dependencies; static binary
- **Non-blocking**: All network operations run in goroutines; UI never freezes
- **Real-time Updates**: Selected server info updates every second automatically
- **SA-MP Protocol**: UDP query compatible with SA-MP and Open.MP servers
- **Graceful Degradation**: Falls back to cached servers if master unreachable
- **Security**: Passwords held in memory; never written to config
- **Cross-Platform**: Same code builds on Linux, macOS, Windows

## Building for Production

```bash
# Linux (amd64)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o omp-tui-linux-amd64 ./cmd/omp-tui

# macOS (universal)
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o omp-tui-darwin-amd64 ./cmd/omp-tui
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o omp-tui-darwin-arm64 ./cmd/omp-tui

# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o omp-tui-windows-amd64.exe ./cmd/omp-tui
```

## Contributing

Contributions welcome! Please submit issues and PRs to the GitHub repository.
