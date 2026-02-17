package launcher

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/rsetiawan7/omp-launcher-tui/internal/config"
)

func DetectRuntime(cfg config.Config) (config.Runtime, error) {
	if cfg.Runtime != config.RuntimeAuto && cfg.Runtime != "" {
		return cfg.Runtime, nil
	}
	if _, err := exec.LookPath("proton"); err == nil {
		return config.RuntimeProton, nil
	}
	// Check for CrossOver on macOS
	if runtime.GOOS == "darwin" {
		if isCrossOverInstalled() {
			return config.RuntimeCrossOver, nil
		}
	}
	if _, err := exec.LookPath("wine"); err == nil {
		return config.RuntimeWine, nil
	}
	if runtime.GOOS == "windows" {
		return config.RuntimeNative, nil
	}
	return "", errors.New("no supported runtime found")
}

func isCrossOverInstalled() bool {
	crossOverPath := "/Applications/CrossOver.app/Contents/SharedSupport/CrossOver/bin/wine"
	if _, err := os.Stat(crossOverPath); err == nil {
		return true
	}
	// Also check if wine is available via CrossOver's symlink
	if winePath, err := exec.LookPath("wine"); err == nil {
		// Check if it's CrossOver's wine (contains CrossOver in path)
		if absPath, err := filepath.EvalSymlinks(winePath); err == nil {
			if filepath.Base(filepath.Dir(filepath.Dir(absPath))) == "CrossOver" {
				return true
			}
		}
	}
	return false
}
