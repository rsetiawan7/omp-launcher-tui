# 📖 Documentation Index

Welcome to the Open.MP TUI Launcher project! This is your guide to all available documentation.

## 🚀 Start Here

**New to the project?** Start with one of these:

1. **[QUICKSTART.md](QUICKSTART.md)** - Get up and running in 5 minutes
2. **[STATUS.md](STATUS.md)** - Visual overview of what's included
3. **[API_INTEGRATION.md](API_INTEGRATION.md)** - How it uses the Open.MP API

## 📚 Complete Guides

### User-Facing
- **[README.md](README.md)** - Full feature documentation, setup instructions, troubleshooting
- **[QUICKSTART.md](QUICKSTART.md)** - Fast onboarding, common commands
- **[API_INTEGRATION.md](API_INTEGRATION.md)** - Open.MP API integration details

### Build & Development
- **[Makefile](Makefile)** - Build automation script (use `make help`)
- **[MAKEFILE.md](MAKEFILE.md)** - Detailed Makefile documentation with examples
- **[MAKEFILE_SUMMARY.md](MAKEFILE_SUMMARY.md)** - Quick Makefile reference
- **[build.sh](build.sh)** - Alternative cross-platform build script

### Project Status
- **[PROJECT_COMPLETION.md](PROJECT_COMPLETION.md)** - Detailed completion report
- **[STATUS.md](STATUS.md)** - Visual project summary

## 🗂️ Project Structure

```
📁 omp-launcher-tui/
├── 📄 README.md                  ← Full documentation
├── 📄 QUICKSTART.md              ← Get started fast
├── 📄 STATUS.md                  ← Project overview
├── 📄 PROJECT_COMPLETION.md      ← Completion report
├── 📄 MAKEFILE.md                ← Build guide
├── 📄 MAKEFILE_SUMMARY.md        ← Quick reference
├── 📄 Makefile                   ← Build automation
├── 📄 build.sh                   ← Shell build script
├── 📄 config.json                ← Example config
├── 📄 servers.json               ← Fallback servers
├── 📄 go.mod                     ← Dependencies
│
├── 📁 cmd/
│   └── omp-tui/
│       └── main.go               ← Application entry point
│
├── 📁 internal/
│   ├── config/                   ← Configuration system
│   │   ├── config.go
│   │   ├── favorites.go
│   │   ├── load.go
│   │   ├── masterlist.go
│   │   └── paths.go
│   │
│   ├── server/                   ← Server list & querying
│   │   ├── model.go
│   │   ├── master.go
│   │   ├── query.go
│   │   ├── cache.go
│   │   └── sort.go
│   │
│   ├── launcher/                 ← Client launching
│   │   ├── launcher.go
│   │   └── runtime.go
│   │
│   └── tui/                      ← UI implementation
│       ├── app.go
│       ├── layout.go
│       ├── keys.go
│       ├── modals.go
│       ├── filebrowser.go
│       ├── masterlist.go
│       └── update.go
│
└── 📁 bin/                       ← Compiled binaries
    ├── omp-tui-darwin-amd64
    ├── omp-tui-darwin-arm64
    ├── omp-tui-linux-amd64
    └── omp-tui-windows-amd64.exe
```

## 🎯 Quick Navigation

## 🌟 Recent Updates

- **Version Filtering** - Filter servers by 0.3.7, 0.3.DL, or open.mp (Press V)
- **Smart Caching** - 24-hour cache with preserved ping/player data
- **Debounced Updates** - 500ms delay prevents excessive server queries
- **Enhanced Filtering** - Combined search and version filter display
- **Manual Refresh** - R key always fetches fresh data

### I want to...

**...get started immediately**
```bash
make build && make run
```
→ See [QUICKSTART.md](QUICKSTART.md)

**...build for all platforms**
```bash
make build-all
```
→ See [MAKEFILE.md](MAKEFILE.md)

**...install globally**
```bash
make install
```
→ See [MAKEFILE_SUMMARY.md](MAKEFILE_SUMMARY.md)

**...understand the project**
→ Read [STATUS.md](STATUS.md)

**...learn about the build system**
→ Read [MAKEFILE.md](MAKEFILE.md)

**...configure the application**
→ Edit `~/.config/omp-tui/config.json` (see [README.md](README.md))

**...run code quality checks**
```bash
make fmt && make vet && make test
```
→ See [MAKEFILE_SUMMARY.md](MAKEFILE_SUMMARY.md)

## 📋 Documentation by Topic

### Building & Running
| Topic | Document |
|-------|----------|
| Quick start | [QUICKSTART.md](QUICKSTART.md) |
| All build targets | [MAKEFILE.md](MAKEFILE.md) |
| Make reference | [MAKEFILE_SUMMARY.md](MAKEFILE_SUMMARY.md) |
| Environment vars | [MAKEFILE.md](MAKEFILE.md#environment-variables) |

### Features & Usage
| Topic | Document |
|-------|----------|
| Full feature list | [README.md](README.md#features) |
| Keybindings | [README.md](README.md#keybindings) or [STATUS.md](STATUS.md#-keybindings) |
| Configuration | [README.md](README.md#configuration) |
| Wine/Proton setup | [README.md](README.md#wine--proton-setup) |

### Development
| Topic | Document |
|-------|----------|
| Code structure | [STATUS.md](STATUS.md#-what's-included) |
| Project layout | [README.md](README.md#project-structure) |
| Troubleshooting | [README.md](README.md#troubleshooting) |
| Code quality targets | [MAKEFILE.md](MAKEFILE.md#code-quality) |

### Project Status
| Topic | Document |
|-------|----------|
| What's implemented | [STATUS.md](STATUS.md#-status-production-ready) |
| Deliverables | [PROJECT_COMPLETION.md](PROJECT_COMPLETION.md#-deliverables) |
| Complete checklist | [PROJECT_COMPLETION.md](PROJECT_COMPLETION.md#-completion-checklist) |

## 🔑 Key Commands

```bash
# Development
make help                    # Show all Makefile targets
make build                  # Build for current platform
make run                    # Build and run
make fmt && make vet        # Code quality checks

# Release
make clean && make build-all # Clean build all platforms
make install                # Install to system
make release                # List release artifacts

# Maintenance
make deps                   # Update dependencies
make tidy                   # Tidy go.mod
make clean                  # Remove build artifacts
```

## 📞 Getting Help

1. **Quick questions?** → Check [QUICKSTART.md](QUICKSTART.md)
2. **Build issues?** → See [MAKEFILE.md](MAKEFILE.md#troubleshooting)
3. **Usage problems?** → Read [README.md](README.md#troubleshooting)
4. **Project overview?** → Look at [STATUS.md](STATUS.md)
5. **Implementation details?** → Check code comments in `internal/`

## ✅ What's Ready

- ✅ Full Go source code (production-quality)
- ✅ 4 compiled binaries (Linux, macOS Intel/ARM, Windows)
- ✅ Comprehensive Makefile (20+ targets)
- ✅ 6 documentation files
- ✅ Favorites system with persistent storage
- ✅ Real-time server updates (ping, players, rules)
- ✅ Ping history chart visualization
- ✅ Example configuration and fallback servers
- ✅ Zero external dependencies
- ✅ Ready for distribution

## 🎮 Next Steps

1. Read [QUICKSTART.md](QUICKSTART.md) to get started
2. Run `make build && make run`
3. Configure at `~/.config/omp-tui/config.json`
4. Enjoy!

---

**For more details, see [README.md](README.md) or [STATUS.md](STATUS.md).**
