package launcher

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/rsetiawan7/omp-launcher-tui/internal/config"
)

type LaunchOptions struct {
	Host     string
	Port     int
	Nickname string
	GTAPath  string
	Password string
}

func Launch(cfg config.Config, opts LaunchOptions) error {
	runtimeChoice, err := DetectRuntime(cfg)
	if err != nil {
		return err
	}

	clientPath := resolveClientPath(opts.GTAPath)
	if clientPath == "" {
		return errors.New("unable to find Open.MP client executable")
	}

	args := []string{"-h", opts.Host, "-p", itoa(opts.Port), "-n", opts.Nickname, "-g", opts.GTAPath}
	if opts.Password != "" {
		args = append(args, "-z", opts.Password)
	}

	cmd, err := buildCommand(runtimeChoice, cfg, clientPath, args)
	if err != nil {
		return err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func buildCommand(runtimeChoice config.Runtime, cfg config.Config, clientPath string, args []string) (*exec.Cmd, error) {
	switch runtimeChoice {
	case config.RuntimeProton:
		cmdArgs := []string{"run", clientPath}
		cmdArgs = append(cmdArgs, args...)
		cmd := exec.Command("proton", cmdArgs...)
		return cmd, nil
	case config.RuntimeWine:
		cmdArgs := append([]string{clientPath}, args...)
		cmd := exec.Command("wine", cmdArgs...)
		return cmd, nil
	case config.RuntimeNative:
		cmdArgs := append([]string{}, args...)
		return exec.Command(clientPath, cmdArgs...), nil
	default:
		return nil, errors.New("unsupported runtime")
	}
}

func resolveClientPath(gtaPath string) string {
	if gtaPath == "" {
		return ""
	}
	candidates := []string{
		"omp-launcher.exe",
		"omp-launcher",
		"omp-client.exe",
		"omp-client",
		"samp.exe",
	}
	for _, candidate := range candidates {
		path := filepath.Join(gtaPath, candidate)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	if runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(gtaPath), ".exe") {
		if _, err := os.Stat(gtaPath); err == nil {
			return gtaPath
		}
	}
	return ""
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
