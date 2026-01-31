package launcher

import (
	"errors"
	"os/exec"
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
	if _, err := exec.LookPath("wine"); err == nil {
		return config.RuntimeWine, nil
	}
	if runtime.GOOS == "windows" {
		return config.RuntimeNative, nil
	}
	return "", errors.New("no supported runtime found")
}
