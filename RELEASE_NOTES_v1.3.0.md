# Release Notes - v1.3.0

## 🎉 What's New in v1.3.0

### ✨ New Features

#### macOS CrossOver Support
- **CrossOver runtime detection** - Automatically detects CrossOver installation on macOS
  - CrossOver is now a supported runtime alongside Wine and Proton
  - Auto-detection of CrossOver wine binary at `/Applications/CrossOver.app/Contents/SharedSupport/CrossOver/bin/wine`
  
- **New configuration fields** for CrossOver setup:
  - `crossover_launcher` - Path to Windows build of omp-launcher-tui.exe in CrossOver bottle (e.g., `C:\Program Files\omp-launcher-tui\omp-launcher-tui.exe`)
  - `crossover_bottle` - Optional bottle name to use (defaults to default bottle)
  
- **CrossOver workflow**:
  1. Install Windows build of omp-launcher-tui.exe in your CrossOver bottle
  2. Run macOS native build to browse servers
  3. Configure CrossOver runtime and launcher path via `C` key
  4. Connect to servers - Windows executable runs in CrossOver and launches GTA SA

- **UI enhancements**:
  - Press `C` to open configuration modal
  - New "CrossOver Launcher" field with file browser support (`Ctrl+B`)
  - New "CrossOver Bottle" field for optional bottle name
  - Runtime dropdown now includes "crossover" option

### 🔧 Improvements
- Updated runtime detection logic to prioritize CrossOver on macOS
- Enhanced launcher to support CrossOver-specific execution path
- Improved documentation with comprehensive CrossOver setup guide
- Added example configuration for CrossOver users

### 📝 Configuration Example

```json
{
  "runtime": "crossover",
  "crossover_launcher": "C:\\Program Files\\omp-launcher-tui\\omp-launcher-tui.exe",
  "crossover_bottle": "YourBottleName"
}
```

### 🐛 Bug Fixes
- Fixed `.gitignore` to exclude `*.log` files

---

## 📦 Installation

Download the appropriate binary for your platform from the releases page:

- **Linux (amd64)**: `omp-tui-linux-amd64`
- **macOS (Intel)**: `omp-tui-darwin-amd64`
- **macOS (Apple Silicon)**: `omp-tui-darwin-arm64`
- **Windows**: `omp-tui-windows-amd64.exe`

Make the binary executable (Linux/macOS):
```bash
chmod +x omp-tui-*
```

---

## 🚀 Quick Start with CrossOver (macOS)

```bash
# 1. Install Windows build in CrossOver bottle
# Download omp-tui-windows-amd64.exe and place it in your bottle

# 2. Run macOS native build
./omp-tui-darwin-arm64  # or darwin-amd64 for Intel

# 3. Press C to configure
# Set Runtime: crossover
# Set CrossOver Launcher: C:\Program Files\omp-launcher-tui\omp-tui-windows-amd64.exe
# Set CrossOver Bottle: YourBottleName (optional)

# 4. Browse and connect to servers!
```

---

## 📝 Full Changelog

- Add CrossOver runtime support for macOS
- Add `crossover_launcher` and `crossover_bottle` configuration fields
- Add CrossOver detection in runtime auto-detection
- Add `launchViaCrossOver` function for CrossOver-specific launching
- Add `isCrossOverInstalled` helper function
- Update configuration modal with CrossOver fields and file browser support
- Update README.md with comprehensive CrossOver setup guide
- Update QUICKSTART.md with CrossOver requirements
- Update STATUS.md to list CrossOver as a feature
- Add `*.log` to `.gitignore`
- Bump version to 1.3.0

---

## 🙏 Acknowledgments

CrossOver support enables macOS users to run GTA: San Andreas and Open.MP without needing to configure Wine manually. This makes the launcher more accessible to macOS users who already have CrossOver installed!

---

## 📚 Documentation

For detailed setup instructions, see:
- [README.md](README.md) - Section "macOS with CrossOver"
- [QUICKSTART.md](QUICKSTART.md) - Requirements section
- [STATUS.md](STATUS.md) - Features list

---

## ⬆️ Upgrading from v1.2.0

No breaking changes. Existing configurations will continue to work. CrossOver support is an additional option for macOS users.

To use CrossOver:
1. Update to v1.3.0
2. Press `C` in the TUI
3. Set Runtime to "crossover"
4. Configure CrossOver Launcher and optionally the Bottle name
5. Enjoy!
