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
		app:           application,
		layout:        layout,
		cfg:           cfg,
		passwords:     make(map[string]string),
		sortMode:      server.SortNone,
		viewMode:      ViewMasterList,
		version:       version,
		updateChecker: updateChecker,
		lastQueryTime: make(map[string]time.Time),
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
		a.RefreshServers()
	}()

	return a.app.SetRoot(root, true).EnableMouse(false).Run()
}

func (a *App) RefreshServers() {
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

	for i := range servers {
		servers[i].Loading = true
	}

	a.servers = servers
	a.applyFilterAndSort()

	a.setBusy(false, fmt.Sprintf("Loaded %d servers", len(servers)))
	go a.queryServers(servers)
}

func (a *App) queryServers(servers []server.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobs := make(chan int)
	workers := 64
	var wg sync.WaitGroup
	var completed int32
	total := len(servers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				entry := servers[idx]
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
				a.updateServer(entry)

				// Update progress
				current := atomic.AddInt32(&completed, 1)
				a.app.QueueUpdateDraw(func() {
					a.layout.SetStatus(fmt.Sprintf("Updated %d of %d servers", current, total))
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
		a.layout.SetStatus(fmt.Sprintf("All %d servers updated", total))
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

func (a *App) applyFilterAndSort() {
	filtered := make([]server.Server, 0, len(a.servers))
	query := strings.TrimSpace(strings.ToLower(a.searchQuery))
	for _, srv := range a.servers {
		if query == "" || strings.Contains(strings.ToLower(srv.Name), query) || strings.Contains(strings.ToLower(srv.Addr()), query) {
			filtered = append(filtered, srv)
		}
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

	// Add search query if present
	if a.searchQuery != "" {
		title += fmt.Sprintf(" [Search: \"%s\"]", a.searchQuery)
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
					go a.RefreshServers()
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
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Query immediately first
	a.queryAndUpdateServer(srv)

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

	res.Loading = false
	a.updateServer(res)

	// Also update in favorites if it exists there
	a.updateFavoriteServer(res)

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
	rules, err := server.QueryServerRules(ctx, srv.Host, srv.Port)
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
		if query == "" || strings.Contains(strings.ToLower(srv.Name), query) || strings.Contains(strings.ToLower(srv.Addr()), query) {
			filtered = append(filtered, srv)
		}
	}
	server.SortServers(filtered, a.sortMode)
	a.filteredFavorites = filtered
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

				a.app.QueueUpdateDraw(func() {
					a.favorites[idx].Name = res.Name
					a.favorites[idx].Players = res.Players
					a.favorites[idx].MaxPlayers = res.MaxPlayers
					a.favorites[idx].Ping = res.Ping
					a.favorites[idx].Passworded = res.Passworded
					a.favorites[idx].Loading = false
					a.favorites[idx].LastUpdated = res.LastUpdated

					if a.viewMode == ViewFavorites {
						a.applyFavoritesFilterAndSort()
						a.layout.UpdateTable(a.filteredFavorites)
					}
				})
			}(i)
		}

		wg.Wait()
		a.app.QueueUpdateDraw(func() {
			a.layout.SetStatus(fmt.Sprintf("Refreshed %d favorites", len(a.favorites)))
		})
	}()
}
