# Release Notes - v1.2.0

## 🎉 What's New in v1.2.0

### ✨ New Features

#### Export/Import Commands
- **Export configuration** - Backup your entire setup (config, favorites, and master lists) to a single JSON file
  ```bash
  ./omp-tui export my-backup.json
  ```
- **Import configuration** - Restore from backup with confirmation prompt
  ```bash
  ./omp-tui import my-backup.json
  ```
- Perfect for backups, migration between systems, and sharing configurations

#### Enhanced Init Command
- **New flags** for automated setup:
  - `--gta-path <path>` - Set GTA San Andreas installation path during init
  - `--omp-launcher <path>` - Set open.mp launcher path during init
  ```bash
  ./omp-tui init --gta-path "/path/to/GTA" --omp-launcher "/path/to/launcher"
  ```
- Eliminates manual configuration editing for automated deployments

#### Improved Connect Command
- **Default port** - Port 7777 is now used when not specified
  ```bash
  ./omp-tui connect 127.0.0.1        # Automatically uses port 7777
  ./omp-tui connect 127.0.0.1:8888   # Custom port still works
  ```
- **Better error messages** - Helpful guidance to use `init` command when config is missing
- More user-friendly experience for CLI users

### 🔧 Improvements
- Added comprehensive unit tests for address parsing
- Enhanced documentation in README.md and QUICKSTART.md
- Clearer error messages guide users through setup process

### ⚠️ Breaking Changes
- `Init()` function now requires `InitOptions` parameter (internal API change)

---

## 📦 Installation

Download the appropriate binary for your platform from the [releases page](https://github.com/rsetiawan7/omp-launcher-tui/releases/tag/v1.2.0):

- **Linux (amd64)**: `omp-tui-linux-amd64`
- **macOS (Intel)**: `omp-tui-darwin-amd64`
- **macOS (Apple Silicon)**: `omp-tui-darwin-arm64`
- **Windows**: `omp-tui-windows-amd64.exe`

Make the binary executable (Linux/macOS):
```bash
chmod +x omp-tui-*
```

---

## 🚀 Quick Start

```bash
# First time setup with automated configuration
./omp-tui init --gta-path "/path/to/GTA" --omp-launcher "/path/to/launcher"

# Connect to a server (port defaults to 7777)
./omp-tui connect 127.0.0.1

# Export your configuration for backup
./omp-tui export backup.json

# Run interactive TUI mode
./omp-tui
```

---

## 📝 Full Changelog

- Add export command to backup config, favorites, and master lists to single file
- Add import command to restore configuration from exported file
- Add `--gta-path` and `--omp-launcher` flags to init command for automated setup
- Make port optional in connect command (defaults to 7777)
- Improve error messages in connect to guide users to use init command
- Add comprehensive unit tests for ParseAddress function
- Update README.md and QUICKSTART.md with new features documentation
- Bump version to 1.2.0

**Full Commit**: `9ce3b63`

---

## 🙏 Acknowledgments

This release includes several quality-of-life improvements based on user feedback. Thank you to everyone who contributed ideas and reported issues!

---

## 📖 Documentation

For detailed documentation, see:
- [README.md](https://github.com/rsetiawan7/omp-launcher-tui/blob/main/README.md)
- [QUICKSTART.md](https://github.com/rsetiawan7/omp-launcher-tui/blob/main/QUICKSTART.md)

---

## 🐛 Known Issues

None reported yet. Please open an issue if you encounter any problems!
