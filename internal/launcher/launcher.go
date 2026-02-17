package launcher

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	// For CrossOver, use CrossOverLauncher if specified, otherwise fall back to OMPLauncher
	if runtimeChoice == config.RuntimeCrossOver {
		if cfg.CrossOverLauncher == "" {
			return errors.New("CrossOverLauncher path not configured")
		}
		return launchViaCrossOver(cfg, opts)
	}

	if cfg.OMPLauncher == "" {
		return errors.New("OMPLauncher path not configured")
	}

	// Resolve launcher executable path
	launcherPath := resolveLauncherPath(cfg.OMPLauncher)
	if launcherPath == "" {
		return errors.New("unable to find Open.MP launcher executable")
	}

	// Build command arguments for Open.MP launcher
	args := []string{"-h", opts.Host, "-p", itoa(opts.Port), "-n", opts.Nickname, "-g", opts.GTAPath}
	if opts.Password != "" {
		args = append(args, "-z", opts.Password)
	}

	cmd, err := buildCommand(runtimeChoice, cfg, launcherPath, args)
	if err != nil {
		return err
	}

	// Print the command that will be executed
	printCommand(cmd, opts.Password)

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

func resolveLauncherPath(ompLauncher string) string {
	if ompLauncher == "" {
		return ""
	}

	// If it's already a file, use it directly
	if info, err := os.Stat(ompLauncher); err == nil && !info.IsDir() {
		return ompLauncher
	}

	// If it's a directory, look for the launcher executable inside
	candidates := []string{
		"omp-launcher.exe",
		"omp-launcher",
	}
	for _, candidate := range candidates {
		path := filepath.Join(ompLauncher, candidate)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}

func printCommand(cmd *exec.Cmd, password string) {
	cmdStr := cmd.Path
	for _, arg := range cmd.Args[1:] {
		// Mask password if it follows the -z flag
		if strings.Contains(cmdStr, "-z") && len(password) > 0 && arg == password {
			cmdStr += " " + strings.Repeat("*", 10)
		} else {
			// Quote arguments with spaces
			if strings.Contains(arg, " ") {
				cmdStr += fmt.Sprintf(" \"%s\"", arg)
			} else {
				cmdStr += " " + arg
			}
		}
	}
	fmt.Printf("Executing: %s\n", cmdStr)
}
func itoa(v int) string {
	return strconv.Itoa(v)
}

func launchViaCrossOver(cfg config.Config, opts LaunchOptions) error {
	winePath := "/Applications/CrossOver.app/Contents/SharedSupport/CrossOver/bin/wine"

	// Check if wine exists
	if _, err := os.Stat(winePath); err != nil {
		return fmt.Errorf("CrossOver wine not found at %s: %w", winePath, err)
	}

	// Note: CrossOverLauncher is a Windows path (e.g., Z:/path/to/file.exe)
	// We can't validate it from macOS, wine will handle the path resolution
	if cfg.CrossOverLauncher == "" {
		return errors.New("CrossOverLauncher path is empty")
	}

	// Build command: wine omp-launcher-tui.exe connect -h <host> -p <port> -n <nickname>
	cmdArgs := []string{cfg.CrossOverLauncher, "connect", "-nickname", opts.Nickname, fmt.Sprintf("%s:%d", opts.Host, opts.Port)}

	cmd := exec.Command(winePath, cmdArgs...)

	// Set CrossOver bottle if specified
	if cfg.CrossOverBottle != "" {
		cmd.Env = append(os.Environ(), "CX_BOTTLE="+cfg.CrossOverBottle)
	}

	// Print the command that will be executed
	fmt.Printf("Executing via CrossOver: %s %s connect -nickname %s %s:%d",
		winePath, cfg.CrossOverLauncher, opts.Nickname, opts.Host, opts.Port)
	if cfg.CrossOverBottle != "" {
		fmt.Printf(" (bottle: %s)", cfg.CrossOverBottle)
	}
	fmt.Println()
	fmt.Println("Note: Password prompt (if needed) will appear from the Windows executable")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
