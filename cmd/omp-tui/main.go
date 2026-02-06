package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rsetiawan7/omp-launcher-tui/internal/cli"
	"github.com/rsetiawan7/omp-launcher-tui/internal/config"
	"github.com/rsetiawan7/omp-launcher-tui/internal/tui"
)

var Version = "1.2.0"

func main() {
	// Check if CLI mode is requested
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "init":
			// Define init-specific flags
			initCmd := flag.NewFlagSet("init", flag.ExitOnError)
			gtaPath := initCmd.String("gta-path", "", "Path to GTA San Andreas installation")
			ompLauncher := initCmd.String("omp-launcher", "", "Path to open.mp launcher executable")

			// Parse flags
			if err := initCmd.Parse(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
				os.Exit(1)
			}

			// Initialize with options
			opts := cli.InitOptions{
				GTAPath:     *gtaPath,
				OMPLauncher: *ompLauncher,
			}

			if err := cli.Init(opts); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)

		case "connect":
			if len(os.Args) < 3 {
				fmt.Fprintf(os.Stderr, "Usage: %s connect <alias|host[:port]>\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "Examples:\n")
				fmt.Fprintf(os.Stderr, "  %s connect my-server        # Connect using alias\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "  %s connect 127.0.0.1        # Connect using IP (port defaults to 7777)\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "  %s connect 127.0.0.1:7777   # Connect using IP with custom port\n", os.Args[0])
				os.Exit(1)
			}

			host, port, alias, err := cli.ResolveAddress(os.Args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if err := cli.Connect(host, port, alias); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return

		case "export":
			if len(os.Args) < 3 {
				fmt.Fprintf(os.Stderr, "Usage: %s export <output-file>\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "Examples:\n")
				fmt.Fprintf(os.Stderr, "  %s export my-config.json\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "  %s export backup/config-$(date +%%Y%%m%%d).json\n", os.Args[0])
				os.Exit(1)
			}

			if err := cli.Export(os.Args[2]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)

		case "import":
			if len(os.Args) < 3 {
				fmt.Fprintf(os.Stderr, "Usage: %s import <input-file>\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "Examples:\n")
				fmt.Fprintf(os.Stderr, "  %s import my-config.json\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "  %s import backup/config-20260206.json\n", os.Args[0])
				os.Exit(1)
			}

			if err := cli.Import(os.Args[2]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	// Run TUI mode
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load error: %v\n", err)
		os.Exit(1)
	}
	checker := tui.GitHubChecker{Owner: "rsetiawan7", Repo: "omp-launcher-tui"}
	app := tui.NewApp(cfg, Version, checker)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "app error: %v\n", err)
		os.Exit(1)
	}
}
