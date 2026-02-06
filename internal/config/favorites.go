package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const FavoritesFile = "favorites.json"

// FavoriteServer represents a user-saved server
type FavoriteServer struct {
	Name        string            `json:"name"`
	Alias       string            `json:"alias,omitempty"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	LastUpdated string            `json:"last_updated,omitempty"`
	Rules       map[string]string `json:"rules,omitempty"`
}

// Favorites holds the list of user favorite servers
type Favorites struct {
	Servers []FavoriteServer `json:"servers"`
}

// FavoritesPath returns the path to the favorites file
func FavoritesPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, FavoritesFile), nil
}

// LoadFavorites loads the favorites from the config file
func LoadFavorites() (Favorites, error) {
	path, err := FavoritesPath()
	if err != nil {
		return Favorites{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Favorites{Servers: []FavoriteServer{}}, nil
		}
		return Favorites{}, err
	}

	var favorites Favorites
	if err := json.Unmarshal(data, &favorites); err != nil {
		return Favorites{}, err
	}
	return favorites, nil
}

// SaveFavorites saves the favorites to the config file
func SaveFavorites(favorites Favorites) error {
	path, err := FavoritesPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), DefaultPerms); err != nil {
		return err
	}

	data, err := json.MarshalIndent(favorites, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// AddFavorite adds a server to favorites if not already present
func AddFavorite(name, alias, host string, port int) error {
	favorites, err := LoadFavorites()
	if err != nil {
		return err
	}

	// Check if already exists
	for _, srv := range favorites.Servers {
		if srv.Host == host && srv.Port == port {
			return nil // Already exists
		}
	}

	// Check alias uniqueness if provided
	if alias != "" && !IsAliasUnique(alias, host, port) {
		return errors.New("alias already exists")
	}

	favorites.Servers = append(favorites.Servers, FavoriteServer{
		Name:  name,
		Alias: alias,
		Host:  host,
		Port:  port,
	})

	return SaveFavorites(favorites)
}

// RemoveFavorite removes a server from favorites
func RemoveFavorite(host string, port int) error {
	favorites, err := LoadFavorites()
	if err != nil {
		return err
	}

	newServers := make([]FavoriteServer, 0, len(favorites.Servers))
	for _, srv := range favorites.Servers {
		if srv.Host != host || srv.Port != port {
			newServers = append(newServers, srv)
		}
	}
	favorites.Servers = newServers

	return SaveFavorites(favorites)
}

// IsFavorite checks if a server is in favorites
func IsFavorite(host string, port int) bool {
	favorites, err := LoadFavorites()
	if err != nil {
		return false
	}

	for _, srv := range favorites.Servers {
		if srv.Host == host && srv.Port == port {
			return true
		}
	}
	return false
}

// IsAliasUnique checks if an alias is unique (not used by other servers)
// Empty aliases are always considered unique
func IsAliasUnique(alias, host string, port int) bool {
	if alias == "" {
		return true
	}

	favorites, err := LoadFavorites()
	if err != nil {
		return true
	}

	for _, srv := range favorites.Servers {
		// Skip the current server (same host:port)
		if srv.Host == host && srv.Port == port {
			continue
		}
		// Check if another server has this alias
		if srv.Alias == alias {
			return false
		}
	}
	return true
}
