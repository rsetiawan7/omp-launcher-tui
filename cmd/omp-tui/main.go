package main

import (
	"fmt"
	"os"

	"github.com/rsetiawan7/omp-launcher-tui/internal/config"
	"github.com/rsetiawan7/omp-launcher-tui/internal/tui"
)

var Version = "1.0.0"

func main() {
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
