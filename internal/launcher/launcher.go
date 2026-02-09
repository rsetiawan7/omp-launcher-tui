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
