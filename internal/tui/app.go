package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rsetiawan7/omp-launcher-tui/internal/config"
	"github.com/rsetiawan7/omp-launcher-tui/internal/launcher"
	"github.com/rsetiawan7/omp-launcher-tui/internal/server"
)

const (
	SERVER_VERSION_037    = "0.3.7"
	SERVER_VERSION_03DL   = "0.3.DL"
	SERVER_VERSION_OPENMP = "open.mp"
)

type ViewMode int

const (
	ViewMasterList ViewMode = iota
	ViewFavorites
)

type App struct {
	app                 *tview.Application
	layout              *Layout
	cfg                 config.Config
	servers             []server.Server
	filtered            []server.Server
	favorites           []server.Server
	filteredFavorites   []server.Server
	passwords           map[string]string
	searchQuery         string
	sortMode            server.SortMode
	viewMode            ViewMode
	versionFilters      map[string]bool
	refreshLock         sync.Mutex
	refreshing          bool
	busy                bool
	version             string
	updateChecker       UpdateChecker
	lastQueryTime       map[string]time.Time
	queryLock           sync.Mutex
	lastUserInteraction time.Time
	userInteractionLock sync.Mutex
	lastModeSwitch      time.Time
	modeSwitchLock      sync.Mutex
	currentlySelected   *server.Server
	selectedServerLock  sync.Mutex
	cancelServerUpdate  context.CancelFunc
	pingHistory         []int64
	pingHistoryLock     sync.Mutex
}

func NewApp(cfg config.Config, version string, updateChecker UpdateChecker) *App {
	application := tview.NewApplication()
	layout := NewLayout()

	// Load active master server from master lists
	activeMaster, err := config.GetActiveMasterList()
	if err == nil && activeMaster != "" {
		cfg.MasterServer = activeMaster
	}

	app := &App{
		app:            application,
		layout:         layout,
		cfg:            cfg,
		passwords:      make(map[string]string),
		sortMode:       server.SortNone,
		viewMode:       ViewMasterList,
		versionFilters: make(map[string]bool),
		version:        version,
		updateChecker:  updateChecker,
		lastQueryTime:  make(map[string]time.Time),
	}
	app.setKeybindings()
	app.layout.SetSelectionChangedFunc(app.onServerSelected)
	app.loadFavorites()
	app.updateStatusKeys()

	// Show browse-only warning if enabled
	if cfg.BrowseOnly {
		app.layout.SetStatus("⚠ BROWSE-ONLY MODE - Server connections disabled")
	}

	return app
}

func (a *App) Run() error {
	root := a.layout.Root()

	// Try to load cached servers first
	go func() {
		cached, err := server.LoadCache()
		if err == nil && len(cached) > 0 {
			a.app.QueueUpdateDraw(func() {
				a.servers = cached
				a.applyFilterAndSort()
				a.layout.SetStatus(fmt.Sprintf("Loaded %d servers from cache", len(cached)))
			})
		}
		// Then refresh in background
		time.Sleep(500 * time.Millisecond)
		a.RefreshServers(false) // forceRefresh=false on startup to use cache
	}()

	return a.app.SetRoot(root, true).EnableMouse(false).Run()
}

func (a *App) RefreshServers(forceRefresh bool) {
	a.refreshLock.Lock()
	if a.refreshing || a.busy {
		a.refreshLock.Unlock()
		return
	}
	a.refreshing = true
	a.busy = true
	a.refreshLock.Unlock()
	defer func() {
		a.refreshLock.Lock()
		a.refreshing = false
		a.refreshLock.Unlock()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	setStatus := func(message string) {
		a.app.QueueUpdateDraw(func() {
			a.layout.SetStatus(message)
		})
	}

	setStatus("Refreshing servers...")
	servers, err := server.FetchServers(ctx, a.cfg.MasterServer)
	if err != nil {
		a.setBusy(false, fmt.Sprintf("Server list error: %v", err))
		return
	}

	// Merge with existing cached data to preserve ping and other info
	existingServers := make(map[string]server.Server)
	for _, srv := range a.servers {
		key := fmt.Sprintf("%s:%d", srv.Host, srv.Port)
		existingServers[key] = srv
	}

	for i := range servers {
		key := fmt.Sprintf("%s:%d", servers[i].Host, servers[i].Port)
		if cached, exists := existingServers[key]; exists {
			// Preserve cached data
			servers[i].Ping = cached.Ping
			servers[i].Rules = cached.Rules
			servers[i].LastUpdated = cached.LastUpdated
		}
		servers[i].Loading = true
	}

	a.servers = servers
	a.applyFilterAndSort()

	a.setBusy(false, fmt.Sprintf("Loaded %d servers", len(servers)))
	go a.queryServers(servers, forceRefresh)
}

func (a *App) queryServers(servers []server.Server, forceRefresh bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobs := make(chan int)
	workers := 64
	var wg sync.WaitGroup
	var completed int32
	var skipped int32
	total := len(servers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				entry := servers[idx]

				// Skip querying if server was updated less than 24 hours ago (only if not forcing refresh)
				if !forceRefresh && !entry.LastUpdated.IsZero() && time.Since(entry.LastUpdated) < 24*time.Hour {
					entry.Loading = false
					a.updateServer(entry)
					current := atomic.AddInt32(&skipped, 1)
					a.app.QueueUpdateDraw(func() {
						a.layout.SetStatus(fmt.Sprintf("Loaded from cache: %d, Updated: %d of %d servers", current, atomic.LoadInt32(&completed), total))
					})
					continue
				}

				res, err := server.QueryServer(ctx, entry.Host, entry.Port)
				if err != nil {
					continue
				}
				entry.Name = res.Name
				entry.Players = res.Players
				entry.MaxPlayers = res.MaxPlayers
				entry.Ping = res.Ping
				entry.Passworded = res.Passworded
				entry.Loading = false
				entry.LastUpdated = res.LastUpdated

				// Query server rules
				rules, err := server.QueryServerRules(ctx, entry.Host, entry.Port)
				if err == nil {
					entry.Rules = rules
				}

				a.updateServer(entry)

				// Update progress
				current := atomic.AddInt32(&completed, 1)
				a.app.QueueUpdateDraw(func() {
					a.layout.SetStatus(fmt.Sprintf("Loaded from cache: %d, Updated: %d of %d servers", atomic.LoadInt32(&skipped), current, total))
				})
			}
		}()
	}

	for idx := range servers {
		jobs <- idx
	}
	close(jobs)
	wg.Wait()

	a.app.QueueUpdateDraw(func() {
		totalSkipped := atomic.LoadInt32(&skipped)
		totalUpdated := atomic.LoadInt32(&completed)
		a.layout.SetStatus(fmt.Sprintf("Loaded from cache: %d, Updated: %d servers", totalSkipped, totalUpdated))
	})

	// Save cache after all servers are updated
	go func() {
		if err := server.SaveCache(a.servers); err != nil {
			a.app.QueueUpdateDraw(func() {
				a.layout.SetStatus(fmt.Sprintf("Failed to save cache: %v", err))
			})
		}
	}()
}

func (a *App) updateServer(updated server.Server) {
	a.app.QueueUpdateDraw(func() {
		// Update in main servers list
		for i := range a.servers {
			if a.servers[i].Host == updated.Host && a.servers[i].Port == updated.Port {
				a.servers[i] = updated
				break
			}
		}

		// Only update the UI if we're in master list view
		if a.viewMode != ViewMasterList {
			return
		}

		// Update in filtered list and get the position
		for i := range a.filtered {
			if a.filtered[i].Host == updated.Host && a.filtered[i].Port == updated.Port {
				a.filtered[i] = updated
				// Update just this row in the table
				a.layout.UpdateTableRow(i, updated)
				break
			}
		}
	})
}

func (a *App) updateFavoriteServer(updated server.Server) {
	a.app.QueueUpdateDraw(func() {
		// Update in favorites list
		for i := range a.favorites {
			if a.favorites[i].Host == updated.Host && a.favorites[i].Port == updated.Port {
				a.favorites[i] = updated
				break
			}
		}

		// Only update the UI if we're in favorites view
		if a.viewMode != ViewFavorites {
			return
		}

		// Update in filtered favorites list and get the position
		for i := range a.filteredFavorites {
			if a.filteredFavorites[i].Host == updated.Host && a.filteredFavorites[i].Port == updated.Port {
				a.filteredFavorites[i] = updated
				// Update just this row in the table
				a.layout.UpdateTableRow(i, updated)
				break
			}
		}
	})
}

// updateFavoriteServerInFile updates the favorites file with rules and last updated timestamp
func (a *App) updateFavoriteServerInFile(srv server.Server) {
	favorites, err := config.LoadFavorites()
	if err != nil {
		return
	}

	// Find and update the favorite with rules and last updated
	updated := false
	for i := range favorites.Servers {
		if favorites.Servers[i].Host == srv.Host && favorites.Servers[i].Port == srv.Port {
			favorites.Servers[i].Name = srv.Name
			favorites.Servers[i].LastUpdated = srv.LastUpdated.Format(time.RFC3339)
			favorites.Servers[i].Rules = srv.Rules
			updated = true
			break
		}
	}

	if updated {
		config.SaveFavorites(favorites)
	}
}

func (a *App) applyFilterAndSort() {
	filtered := make([]server.Server, 0, len(a.servers))
	query := strings.TrimSpace(strings.ToLower(a.searchQuery))
	for _, srv := range a.servers {
		// Apply text search filter
		if query != "" && !strings.Contains(strings.ToLower(srv.Name), query) && !strings.Contains(strings.ToLower(srv.Addr()), query) {
			continue
		}
		// Apply version filter
		if !a.matchesVersionFilter(srv) {
			continue
		}
		filtered = append(filtered, srv)
	}
	server.SortServers(filtered, a.sortMode)
	a.filtered = filtered
	a.updateTableTitle()
	a.layout.UpdateTable(filtered)
}

func (a *App) updateTableTitle() {
	var title string
	if a.viewMode == ViewFavorites {
		title = "★ Favorites"
	} else {
		title = "Servers"
	}

	// Add sort mode
	switch a.sortMode {
	case server.SortPing:
		title += " [Sort: Ping ↓]"
	case server.SortPlayers:
		title += " [Sort: Players ↓]"
	}

	a.layout.SetTableTitle(title)
}

func (a *App) showConfigModal() {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Configuration (Ctrl+B: Browse | Ctrl+T: Test | Esc: Close)")

	// Get active master list name
	masterListName, err := config.GetActiveMasterListName()
	if err != nil {
		masterListName = "Unknown"
	}

	form.AddInputField("Master List", masterListName, 30, nil, nil)
	form.GetFormItemByLabel("Master List").(*tview.InputField).SetDisabled(true)

	form.AddInputField("Nickname", a.cfg.Nickname, 24, nil, func(text string) {
		a.cfg.Nickname = text
		_ = config.Save(a.cfg)
	})

	// Create GTA Path field with custom keybinding and test button
	gtaPathLabel := "GTA SA Path"
	form.AddInputField(gtaPathLabel, a.cfg.GTAPath, 30, nil, func(text string) {
		a.cfg.GTAPath = text
		_ = config.Save(a.cfg)
	})

	// Add custom input capture to GTA Path field
	gtaPathItem := form.GetFormItemByLabel(gtaPathLabel).(*tview.InputField)
	gtaPathItem.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlB {
			a.showFileBrowser("Select GTA Directory", func(path string) {
				a.cfg.GTAPath = path
				_ = config.Save(a.cfg)
				gtaPathItem.SetText(path)
			}, a.cfg.GTAPath)
			return nil
		} else if event.Key() == tcell.KeyCtrlT {
			// Test if gta_sa.exe exists
			testPath := gtaPathItem.GetText()
			if testPath == "" {
				a.layout.SetStatus("GTA SA Path is empty")
				return nil
			}
			exePath := testPath + "/gta_sa.exe"
			if _, err := os.Stat(exePath); err == nil {
				a.layout.SetStatus("✓ gta_sa.exe found")
			} else {
				a.layout.SetStatus("✗ gta_sa.exe not found")
			}
			return nil
		}
		return event
	})

	// open.mp Launcher Location field
	form.AddInputField("open.mp Launcher", a.cfg.OMPLauncher, 40, nil, func(text string) {
		a.cfg.OMPLauncher = text
		_ = config.Save(a.cfg)
	})

	// Add custom input capture to open.mp Launcher field
	ompLauncherItem := form.GetFormItemByLabel("open.mp Launcher").(*tview.InputField)
	ompLauncherItem.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlB {
			a.showFileBrowser("Select open.mp Launcher", func(path string) {
				a.cfg.OMPLauncher = path
				_ = config.Save(a.cfg)
				ompLauncherItem.SetText(path)
			}, a.cfg.OMPLauncher)
			return nil
		}
		return event
	})
	form.AddDropDown("Runtime", []string{"auto", "wine", "proton", "native"}, runtimeIndex(a.cfg.Runtime), func(option string, _ int) {
		a.cfg.Runtime = config.Runtime(option)
		_ = config.Save(a.cfg)
	})

	// Browse Only checkbox
	form.AddCheckbox("Browse Only Mode", a.cfg.BrowseOnly, func(checked bool) {
		a.cfg.BrowseOnly = checked
		_ = config.Save(a.cfg)
		if checked {
			a.layout.SetStatus("⚠ BROWSE-ONLY MODE - Server connections disabled")
		} else {
			a.layout.SetStatus("Browse-only mode disabled")
		}
	})

	// Add custom input capture to form for arrow key navigation and escape
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			a.setKeybindings()
			a.app.SetRoot(a.layout.Root(), true)
			a.app.SetFocus(a.layout.Table())
			return nil
		case tcell.KeyUp:
			// Move to previous field (Tab)
			return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
		case tcell.KeyDown:
			// Move to next field (Shift+Tab)
			return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModShift)
		}
		return event
	})

	// Clear global keybindings for modal
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.setKeybindings()
			a.app.SetRoot(a.layout.Root(), true)
			a.app.SetFocus(a.layout.Table())
			return nil
		}
		return event
	})

	a.app.SetRoot(form, true).SetFocus(form)
}

func runtimeIndex(rt config.Runtime) int {
	switch rt {
	case config.RuntimeWine:
		return 1
	case config.RuntimeProton:
		return 2
	case config.RuntimeNative:
		return 3
	default:
		return 0
	}
}

func (a *App) setBusy(value bool, status string) {
	a.refreshLock.Lock()
	a.busy = value
	a.refreshLock.Unlock()
	if status != "" {
		a.app.QueueUpdateDraw(func() {
			a.layout.SetStatus(status)
		})
	}
}

func (a *App) isBusy() bool {
	a.refreshLock.Lock()
	defer a.refreshLock.Unlock()
	return a.busy
}

func (a *App) updateStatusKeys() {
	keys := ""
	if a.viewMode == ViewFavorites {
		// Favorites view
		keys = "[::b]↑↓[::] Navigate  [::b]C[::] Config  [::b]Enter[::] Connect  [::b]/[::] Search  [::b]R[::] Refresh  [::b]S[::] Sort  [::b]F[::] Master List  [::b]A[::] Add  [::b]D[::] Remove  [::b]Q[::] Quit"
	} else {
		// Server table is focused (default)
		keys = "[::b]↑↓[::] Navigate  [::b]C[::] Config  [::b]Enter[::] Connect  [::b]/[::] Search  [::b]R[::] Refresh  [::b]S[::] Sort  [::b]F[::] Favorites  [::b]A[::] Add Fav  [::b]★[::] Fav Server  [::b]M[::] Master  [::b]Q[::] Quit"
	}
	a.layout.SetKeysText(keys)
}

func (a *App) setKeybindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if a.isBusy() {
			switch event.Key() {
			case tcell.KeyCtrlC:
				a.app.Stop()
				return nil
			case tcell.KeyRune:
				if strings.ToLower(string(event.Rune())) == "q" {
					a.app.Stop()
					return nil
				}
			}
			return nil
		}
		switch event.Key() {
		case tcell.KeyEnter:
			a.handleConnect()
			return nil
		case tcell.KeyRune:
			// Only apply keybindings when table is focused
			if a.app.GetFocus() != a.layout.Table() {
				return event
			}
			switch strings.ToLower(string(event.Rune())) {
			case "q":
				a.app.Stop()
				return nil
			case "c":
				a.showConfigModal()
				return nil
			case "r":
				if a.viewMode == ViewFavorites {
					go a.refreshFavorites()
				} else {
					go a.RefreshServers(true) // forceRefresh=true for manual refresh
				}
				return nil
			case "s":
				a.cycleSortMode()
				return nil
			case "p":
				a.promptPassword()
				return nil
			case "u":
				a.checkForUpdates()
				return nil
			case "/":
				a.promptSearch()
				return nil
			case "m":
				if a.viewMode != ViewFavorites {
					a.showMasterListManager()
				}
				return nil
			case "f":
				a.toggleViewMode()
				return nil
			case "*":
				if a.viewMode != ViewFavorites {
					a.toggleFavorite()
				}
				return nil
			case "a":
				a.addCustomFavorite()
				return nil
			case "v":
				a.showVersionFilterDialog()
				return nil
			case "d":
				if a.viewMode == ViewFavorites {
					a.toggleFavorite() // Remove from favorites
				}
				return nil
			}
		case tcell.KeyCtrlC:
			a.app.Stop()
			return nil
		}
		return event
	})
}

func (a *App) onServerSelected(row int) {
	var list []server.Server
	if a.viewMode == ViewFavorites {
		list = a.filteredFavorites
	} else {
		list = a.filtered
	}

	if row <= 0 || row-1 >= len(list) {
		// Cancel previous update if selection is invalid
		a.selectedServerLock.Lock()
		if a.cancelServerUpdate != nil {
			a.cancelServerUpdate()
			a.cancelServerUpdate = nil
		}
		a.currentlySelected = nil
		a.selectedServerLock.Unlock()
		return
	}

	// Mark user interaction
	a.userInteractionLock.Lock()
	a.lastUserInteraction = time.Now()
	a.userInteractionLock.Unlock()

	srv := list[row-1]

	// Cancel previous update goroutine if different server selected
	a.selectedServerLock.Lock()
	if a.currentlySelected != nil && (a.currentlySelected.Host != srv.Host || a.currentlySelected.Port != srv.Port) {
		if a.cancelServerUpdate != nil {
			a.cancelServerUpdate()
		}
		// Clear ping history for new server
		a.pingHistoryLock.Lock()
		a.pingHistory = []int64{}
		a.pingHistoryLock.Unlock()
	}
	a.currentlySelected = &srv
	a.selectedServerLock.Unlock()

	// Start continuous update for this server
	ctx, cancel := context.WithCancel(context.Background())
	a.selectedServerLock.Lock()
	a.cancelServerUpdate = cancel
	a.selectedServerLock.Unlock()

	go a.updateSelectedServerContinuously(ctx, srv)
}

func (a *App) updateSelectedServerContinuously(ctx context.Context, srv server.Server) {
	// Wait 500ms before the first query (debounce)
	select {
	case <-ctx.Done():
		return
	case <-time.After(500 * time.Millisecond):
		// Continue to query
	}

	// Check if this is still the selected server after waiting
	a.selectedServerLock.Lock()
	if a.currentlySelected == nil || a.currentlySelected.Host != srv.Host || a.currentlySelected.Port != srv.Port {
		a.selectedServerLock.Unlock()
		return
	}
	a.selectedServerLock.Unlock()

	// Query the server for the first time
	a.queryAndUpdateServer(srv)

	// Continue querying every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check if this is still the selected server
			a.selectedServerLock.Lock()
			if a.currentlySelected == nil || a.currentlySelected.Host != srv.Host || a.currentlySelected.Port != srv.Port {
				a.selectedServerLock.Unlock()
				return
			}
			a.selectedServerLock.Unlock()

			a.queryAndUpdateServer(srv)
		}
	}
}

func (a *App) queryAndUpdateServer(srv server.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := server.QueryServer(ctx, srv.Host, srv.Port)
	if err != nil {
		return
	}

	// Query rules and add to server
	rules, err := server.QueryServerRules(ctx, srv.Host, srv.Port)
	if err == nil {
		res.Rules = rules
	} else {
		res.Rules = map[string]string{}
	}

	res.Loading = false
	a.updateServer(res)

	// Also update in favorites if it exists there
	a.updateFavoriteServer(res)

	// Update favorites file with rules and last updated
	if config.IsFavorite(res.Host, res.Port) {
		go a.updateFavoriteServerInFile(res)
	}

	// Add ping to history
	pingMs := res.Ping.Milliseconds()
	a.pingHistoryLock.Lock()
	a.pingHistory = append(a.pingHistory, pingMs)
	// Keep last 50 pings
	if len(a.pingHistory) > 50 {
		a.pingHistory = a.pingHistory[1:]
	}
	history := make([]int64, len(a.pingHistory))
	copy(history, a.pingHistory)
	a.pingHistoryLock.Unlock()

	// Update ping chart
	a.app.QueueUpdateDraw(func() {
		a.layout.SetPingChart(history)
	})

	// Query players
	players, err := server.QueryServerPlayers(ctx, srv.Host, srv.Port)
	if err != nil || len(players) == 0 {
		a.app.QueueUpdateDraw(func() {
			a.layout.SetPlayers([]string{}, res.Players)
		})
	} else {
		a.app.QueueUpdateDraw(func() {
			a.layout.SetPlayers(players, res.Players)
		})
	}

	// Query rules
	rules, err = server.QueryServerRules(ctx, srv.Host, srv.Port)
	if err != nil {
		a.app.QueueUpdateDraw(func() {
			a.layout.SetRules(map[string]string{})
		})
		return
	}

	a.app.QueueUpdateDraw(func() {
		a.layout.SetRules(rules)
	})
}

func (a *App) selectedServer() (server.Server, bool) {
	row, _ := a.layout.Table().GetSelection()
	var list []server.Server
	if a.viewMode == ViewFavorites {
		list = a.filteredFavorites
	} else {
		list = a.filtered
	}
	if row <= 0 || row-1 >= len(list) {
		return server.Server{}, false
	}
	return list[row-1], true
}

func (a *App) cycleSortMode() {
	switch a.sortMode {
	case server.SortNone:
		a.sortMode = server.SortPing
	case server.SortPing:
		a.sortMode = server.SortPlayers
	default:
		a.sortMode = server.SortNone
	}
	if a.viewMode == ViewFavorites {
		a.applyFavoritesFilterAndSort()
		a.layout.UpdateTable(a.filteredFavorites)
		a.updateTableTitle()
	} else {
		a.applyFilterAndSort()
	}
}

func (a *App) handleConnect() {
	if a.cfg.BrowseOnly {
		a.layout.SetStatus("⚠ Browse-only mode enabled. Cannot connect to servers.")
		return
	}

	srv, ok := a.selectedServer()
	if !ok {
		return
	}
	if srv.Passworded {
		key := srv.Addr()
		if a.passwords[key] == "" {
			a.promptPassword()
			return
		}
	}
	a.launchServer(srv)
}

func (a *App) launchServer(srv server.Server) {
	password := a.passwords[srv.Addr()]
	opts := launcher.LaunchOptions{
		Host:     srv.Host,
		Port:     srv.Port,
		Nickname: a.cfg.Nickname,
		GTAPath:  a.cfg.GTAPath,
		Password: password,
	}
	a.app.Stop()
	if err := launcher.Launch(a.cfg, opts); err != nil {
		fmt.Printf("launch error: %v\n", err)
	}
}

func (a *App) loadFavorites() {
	favorites, err := config.LoadFavorites()
	if err != nil {
		return
	}

	a.favorites = make([]server.Server, len(favorites.Servers))
	for i, fav := range favorites.Servers {
		a.favorites[i] = server.Server{
			Name:    fav.Name,
			Host:    fav.Host,
			Port:    fav.Port,
			Loading: true,
		}
	}
	a.applyFavoritesFilterAndSort()
}

func (a *App) applyFavoritesFilterAndSort() {
	filtered := make([]server.Server, 0, len(a.favorites))
	query := strings.TrimSpace(strings.ToLower(a.searchQuery))
	for _, srv := range a.favorites {
		// Apply text search filter
		if query != "" && !strings.Contains(strings.ToLower(srv.Name), query) && !strings.Contains(strings.ToLower(srv.Addr()), query) {
			continue
		}
		// Apply version filter
		if !a.matchesVersionFilter(srv) {
			continue
		}
		filtered = append(filtered, srv)
	}
	server.SortServers(filtered, a.sortMode)
	a.filteredFavorites = filtered
}

func (a *App) matchesVersionFilter(srv server.Server) bool {
	// If no version filters are active, show all servers
	if len(a.versionFilters) == 0 {
		return true
	}

	// Check if server has rules
	if len(srv.Rules) == 0 {
		return false
	}

	// Static allowed versions
	allowedVersions := []string{SERVER_VERSION_037, SERVER_VERSION_03DL, SERVER_VERSION_OPENMP}

	// Check version rule
	if version, ok := srv.Rules["version"]; ok {
		for _, allowedVer := range allowedVersions {
			ver := allowedVer
			if allowedVer == SERVER_VERSION_OPENMP {
				ver = "omp"
			}

			if strings.Contains(version, ver) && a.versionFilters[allowedVer] {
				return true
			}
		}
	}

	return false
}

func (a *App) collectAvailableVersions() []string {
	// Return static allowed versions
	return []string{SERVER_VERSION_037, SERVER_VERSION_03DL, SERVER_VERSION_OPENMP}
}

func (a *App) updateFilterPanel() {
	var filters []string

	// Add search query if present
	if a.searchQuery != "" {
		filters = append(filters, fmt.Sprintf("Search: \"%s\"", a.searchQuery))
	}

	// Add version filters
	activeVersionFilters := make([]string, 0, len(a.versionFilters))
	for version, active := range a.versionFilters {
		if active {
			activeVersionFilters = append(activeVersionFilters, version)
		}
	}
	if len(activeVersionFilters) > 0 {
		filters = append(filters, fmt.Sprintf("Version: %s", strings.Join(activeVersionFilters, ", ")))
	}

	if len(filters) == 0 {
		a.layout.UpdateFilterPanel("No filters active")
	} else {
		text := fmt.Sprintf("Filters: %s", strings.Join(filters, " | "))
		a.layout.UpdateFilterPanel(text)
	}
}

func (a *App) toggleViewMode() {
	// Check if enough time has passed since last mode switch (2 second cooldown)
	a.modeSwitchLock.Lock()
	if time.Since(a.lastModeSwitch) < 1*time.Second {
		a.modeSwitchLock.Unlock()
		a.layout.SetStatus("Please wait before switching modes again")
		return
	}
	a.lastModeSwitch = time.Now()
	a.modeSwitchLock.Unlock()

	if a.viewMode == ViewMasterList {
		a.viewMode = ViewFavorites
		a.applyFavoritesFilterAndSort()
		a.layout.UpdateTable(a.filteredFavorites)
		a.updateTableTitle()
		a.layout.SetStatus(fmt.Sprintf("Switched to Favorites view (%d servers)", len(a.filteredFavorites)))
	} else {
		a.viewMode = ViewMasterList
		// Don't call applyFilterAndSort because it calls updateTableTitle
		// We want to update the title after changing view mode
		filtered := make([]server.Server, 0, len(a.servers))
		query := strings.TrimSpace(strings.ToLower(a.searchQuery))
		for _, srv := range a.servers {
			if query == "" || strings.Contains(strings.ToLower(srv.Name), query) || strings.Contains(strings.ToLower(srv.Addr()), query) {
				filtered = append(filtered, srv)
			}
		}
		server.SortServers(filtered, a.sortMode)
		a.filtered = filtered
		a.layout.UpdateTable(a.filtered)
		a.updateTableTitle()
		a.layout.SetStatus(fmt.Sprintf("Switched to Master List view (%d servers)", len(a.filtered)))
	}
	a.updateStatusKeys()
}

func (a *App) toggleFavorite() {
	srv, ok := a.selectedServer()
	if !ok {
		return
	}

	if config.IsFavorite(srv.Host, srv.Port) {
		// Remove from favorites
		if err := config.RemoveFavorite(srv.Host, srv.Port); err != nil {
			a.layout.SetStatus(fmt.Sprintf("Failed to remove favorite: %v", err))
			return
		}
		a.layout.SetStatus(fmt.Sprintf("Removed %s from favorites", srv.Name))

		// Update local favorites list
		newFavorites := make([]server.Server, 0, len(a.favorites))
		for _, f := range a.favorites {
			if f.Host != srv.Host || f.Port != srv.Port {
				newFavorites = append(newFavorites, f)
			}
		}
		a.favorites = newFavorites

		// Refresh view if in favorites mode
		if a.viewMode == ViewFavorites {
			a.applyFavoritesFilterAndSort()
			a.layout.UpdateTable(a.filteredFavorites)
		}
	} else {
		// Add to favorites
		if err := config.AddFavorite(srv.Name, srv.Host, srv.Port); err != nil {
			a.layout.SetStatus(fmt.Sprintf("Failed to add favorite: %v", err))
			return
		}
		a.layout.SetStatus(fmt.Sprintf("Added %s to favorites", srv.Name))

		// Update local favorites list
		a.favorites = append(a.favorites, srv)
		a.applyFavoritesFilterAndSort()
	}
}

func (a *App) addCustomFavorite() {
	a.showAddFavoriteDialog()
}

func (a *App) refreshFavorites() {
	if len(a.favorites) == 0 {
		a.layout.SetStatus("No favorites to refresh")
		return
	}

	a.layout.SetStatus("Refreshing favorites...")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var wg sync.WaitGroup
		for i := range a.favorites {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				srv := a.favorites[idx]
				res, err := server.QueryServer(ctx, srv.Host, srv.Port)
				if err != nil {
					return
				}

				// Query rules
				rules, err := server.QueryServerRules(ctx, srv.Host, srv.Port)
				if err == nil {
					res.Rules = rules
				} else {
					res.Rules = map[string]string{}
				}

				a.app.QueueUpdateDraw(func() {
					a.favorites[idx].Name = res.Name
					a.favorites[idx].Players = res.Players
					a.favorites[idx].MaxPlayers = res.MaxPlayers
					a.favorites[idx].Ping = res.Ping
					a.favorites[idx].Passworded = res.Passworded
					a.favorites[idx].Loading = false
					a.favorites[idx].LastUpdated = res.LastUpdated
					a.favorites[idx].Rules = res.Rules

					if a.viewMode == ViewFavorites {
						a.applyFavoritesFilterAndSort()
						a.layout.UpdateTable(a.filteredFavorites)
					}
				})

				// Update favorites file with rules and last updated
				go a.updateFavoriteServerInFile(res)
			}(i)
		}

		wg.Wait()
		a.app.QueueUpdateDraw(func() {
			a.layout.SetStatus(fmt.Sprintf("Refreshed %d favorites", len(a.favorites)))
		})
	}()
}
