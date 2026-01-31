package config

import (
	"os"
	"path/filepath"
)

const (
	AppName      = "omp-tui"
	ConfigFile   = "config.json"
	CacheFile    = "servers_cache.json"
	DefaultPerms = 0o755
)

func ConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, AppName), nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFile), nil
}

func CachePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, CacheFile), nil
}
