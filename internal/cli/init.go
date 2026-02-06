package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rsetiawan7/omp-launcher-tui/internal/config"
	"github.com/rsetiawan7/omp-launcher-tui/internal/server"
)

// InitOptions holds configuration options for initialization
type InitOptions struct {
	GTAPath     string
	OMPLauncher string
}

// Init initializes the configuration directory and files
func Init(opts InitOptions) error {
	// Get config directory
	configDir, err := config.ConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	// Check if config directory exists
	if _, err := os.Stat(configDir); err == nil {
		fmt.Printf("Configuration directory already exists: %s\n", configDir)
		fmt.Print("Do you want to reset configuration? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Init cancelled.")
			return nil
		}
	}

	// Create config directory
	if err := os.MkdirAll(configDir, config.DefaultPerms); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	fmt.Printf("✓ Created config directory: %s\n", configDir)

	// Generate default config file
	cfg := config.DefaultConfig()

	// Apply provided options
	if opts.GTAPath != "" {
		cfg.GTAPath = opts.GTAPath
		fmt.Printf("✓ GTA path set to: %s\n", opts.GTAPath)
	}
	if opts.OMPLauncher != "" {
		cfg.OMPLauncher = opts.OMPLauncher
		fmt.Printf("✓ OMP launcher path set to: %s\n", opts.OMPLauncher)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	configPath, _ := config.ConfigPath()
	fmt.Printf("✓ Generated config file: %s\n", configPath)

	// Create empty favorites file
	favorites := config.Favorites{Servers: []config.FavoriteServer{}}
	if err := config.SaveFavorites(favorites); err != nil {
		return fmt.Errorf("failed to save favorites: %w", err)
	}
	favoritesPath, _ := config.FavoritesPath()
	fmt.Printf("✓ Generated favorites file: %s\n", favoritesPath)

	// Create master list file with Open.MP official server
	masterLists := config.MasterLists{
		Lists: []config.MasterList{
			{
				Name:        "Open.MP Official",
				Host:        "https://api.open.mp/servers",
				Description: "Official Open.MP master server list",
				Active:      true,
			},
		},
	}
	if err := config.SaveMasterLists(masterLists); err != nil {
		return fmt.Errorf("failed to save master lists: %w", err)
	}
	masterListPath, _ := config.MasterListPath()
	fmt.Printf("✓ Generated master list file: %s\n", masterListPath)

	// Fetch servers from master list
	fmt.Printf("\nFetching servers from master list...\n")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	servers, err := server.FetchFromMaster(ctx, "https://api.open.mp/servers")
	if err != nil {
		return fmt.Errorf("failed to fetch from master list: %w", err)
	}
	fmt.Printf("✓ Fetched %d servers from master list\n", len(servers))

	// Query each server for detailed information
	fmt.Printf("\nQuerying server details (this may take a while)...\n")
	queriedServers := queryServers(servers)
	fmt.Printf("✓ Successfully queried %d servers\n", len(queriedServers))

	// Save cache
	if err := server.SaveCache(queriedServers); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}
	cachePath, _ := config.CachePath()
	fmt.Printf("✓ Saved servers cache: %s\n", cachePath)

	// Print summary
	fmt.Printf("\n========================================\n")
	fmt.Printf("Initialization completed successfully!\n")
	fmt.Printf("========================================\n\n")
	fmt.Printf("Generated files:\n")
	fmt.Printf("  Config:        %s\n", configPath)
	fmt.Printf("  Favorites:     %s\n", favoritesPath)
	fmt.Printf("  Master List:   %s\n", masterListPath)
	fmt.Printf("  Servers Cache: %s\n", cachePath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("1. Edit your config at: %s\n", configPath)
	fmt.Printf("2. Set your nickname, GTA path, and OMP launcher path\n")
	fmt.Printf("3. Run '%s' to start the TUI\n", filepath.Base(os.Args[0]))

	return nil
}

// queryServers queries all servers concurrently with a limit
func queryServers(servers []server.Server) []server.Server {
	const maxConcurrent = 50
	const timeout = 3 * time.Second

	queriedServers := make([]server.Server, 0, len(servers))
	var mu sync.Mutex
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, maxConcurrent)

	for i := range servers {
		wg.Add(1)
		go func(srv server.Server) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Query server with rules in one call
			queriedSrv, err := server.QueryServerWithRules(ctx, srv.Host, srv.Port)
			if err != nil {
				// Skip servers that fail to respond
				return
			}

			mu.Lock()
			queriedServers = append(queriedServers, queriedSrv)
			mu.Unlock()

			// Print progress
			if len(queriedServers)%100 == 0 {
				mu.Lock()
				fmt.Printf("  Queried %d servers...\n", len(queriedServers))
				mu.Unlock()
			}
		}(servers[i])
	}

	wg.Wait()
	return queriedServers
}
