package server

import (
	"encoding/json"
	"os"
	"time"

	"github.com/rsetiawan7/omp-launcher-tui/internal/config"
)

type ServerCache struct {
	Servers   []Server  `json:"servers"`
	UpdatedAt time.Time `json:"updated_at"`
}

func LoadCache() ([]Server, error) {
	path, err := config.CachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cache ServerCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	// Return cached servers if less than 1 hour old
	if time.Since(cache.UpdatedAt) < 1*time.Hour {
		return cache.Servers, nil
	}

	return nil, nil
}

func SaveCache(servers []Server) error {
	path, err := config.CachePath()
	if err != nil {
		return err
	}

	cache := ServerCache{
		Servers:   servers,
		UpdatedAt: time.Now(),
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	dir, err := config.ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, config.DefaultPerms); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
