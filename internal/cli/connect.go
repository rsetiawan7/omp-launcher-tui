package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rsetiawan7/omp-launcher-tui/internal/config"
	"github.com/rsetiawan7/omp-launcher-tui/internal/launcher"
	"github.com/rsetiawan7/omp-launcher-tui/internal/server"
)

// Connect connects to a server directly via CLI
func Connect(host string, port int, alias string) error {
	// Load config to get game path and launcher path
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\n\nRun '%s init' to generate initial configuration", err, os.Args[0])
	}

	// Check if game path and launcher path are configured
	if cfg.GTAPath == "" || cfg.OMPLauncher == "" {
		return fmt.Errorf("game path and OMP launcher are not configured\n\nRun '%s init --gta-path <path> --omp-launcher <path>' to set up configuration\nOr run '%s' (TUI mode) and configure them using the 'C' key", os.Args[0], os.Args[0])
	}

	// Query server information
	if alias != "" {
		fmt.Printf("Connecting to '%s' (%s:%d)...\n", alias, host, port)
	} else {
		fmt.Printf("Connecting to %s:%d...\n", host, port)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv, err := server.QueryServer(ctx, host, port)
	if err != nil {
		return fmt.Errorf("failed to query server: %w", err)
	}

	fmt.Printf("\nServer: %s\n", srv.Name)
	fmt.Printf("Players: %d/%d\n", srv.Players, srv.MaxPlayers)
	fmt.Printf("Ping: %v\n", srv.Ping)

	if srv.Passworded {
		fmt.Printf("Password: Required\n")
	} else {
		fmt.Printf("Password: Not required\n")
	}

	// If server requires password, prompt for it
	password := ""
	if srv.Passworded {
		fmt.Print("\nEnter server password: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = strings.TrimSpace(input)
		if password == "" {
			return fmt.Errorf("password is required for this server")
		}
	}

	// Launch the game
	fmt.Printf("\nLaunching game...\n")
	opts := launcher.LaunchOptions{
		Host:     host,
		Port:     port,
		Nickname: cfg.Nickname,
		GTAPath:  cfg.GTAPath,
		Password: password,
	}

	err = launcher.Launch(cfg, opts)
	if err != nil {
		return fmt.Errorf("failed to launch game: %w", err)
	}

	fmt.Printf("Game launched successfully!\n")
	return nil
}

// ParseAddress parses an address in the format "host:port" or "host" (defaults to port 7777)
func ParseAddress(addr string) (string, int, error) {
	parts := strings.Split(addr, ":")

	host := strings.TrimSpace(parts[0])
	if host == "" {
		return "", 0, fmt.Errorf("host cannot be empty")
	}

	// Default port to 7777 if not specified
	port := 7777
	if len(parts) == 2 {
		portStr := strings.TrimSpace(parts[1])
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return "", 0, fmt.Errorf("invalid port: %w", err)
		}
	} else if len(parts) > 2 {
		return "", 0, fmt.Errorf("invalid address format. Expected 'host:port' or 'host' (e.g., '127.0.0.1:7777' or '127.0.0.1')")
	}

	if port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("port must be between 1 and 65535")
	}

	return host, port, nil
}

// ResolveAddress resolves an address which can be either an alias or host:port or host format
// If it's an alias, it looks up the favorite server and returns its host:port
// If it's not an alias, it tries to parse it as host:port or host (defaults to port 7777)
func ResolveAddress(addr string) (host string, port int, alias string, err error) {
	// First, try to find it as an alias in favorites
	favorites, loadErr := config.LoadFavorites()
	if loadErr == nil {
		for _, fav := range favorites.Servers {
			if fav.Alias == addr {
				// Found as alias
				return fav.Host, fav.Port, fav.Alias, nil
			}
		}
	}

	// Not found as alias, try parsing as host:port or host
	host, port, err = ParseAddress(addr)
	if err != nil {
		return "", 0, "", err
	}

	return host, port, "", nil
}
