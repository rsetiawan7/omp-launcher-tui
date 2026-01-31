package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	masterTimeout = 5 * time.Second
)

type apiServerResponse struct {
	IP         string `json:"ip"`
	Hostname   string `json:"hn"`
	Players    int    `json:"pc"`
	MaxPlayers int    `json:"pm"`
	Gamemode   string `json:"gm"`
	Language   string `json:"la"`
	Password   bool   `json:"pa"`
}

// TestMasterServer tests if a master server URL is reachable and returns valid JSON
func TestMasterServer(ctx context.Context, masterURL string) error {
	if masterURL == "" {
		return errors.New("URL cannot be empty")
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, masterURL, nil)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Execute request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Try to parse as JSON array
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var servers []apiServerResponse
	if err := json.Unmarshal(body, &servers); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// Validate we got at least the expected structure
	if len(servers) == 0 {
		return errors.New("server list is empty (no servers returned)")
	}

	// Validate first server has required fields
	if servers[0].IP == "" {
		return errors.New("invalid format: missing 'ip' field")
	}

	return nil
}

// FetchFromMaster fetches the server list from Open.MP API.
// If it fails, it returns an error and the caller can fallback.
func FetchFromMaster(ctx context.Context, masterURL string) ([]Server, error) {
	// Validate URL
	if masterURL == "" {
		masterURL = "https://api.open.mp/servers"
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, masterURL, nil)
	if err != nil {
		return nil, err
	}

	// Set reasonable timeout
	deadline := time.Now().Add(masterTimeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	req = req.WithContext(ctx)

	// Execute request
	client := &http.Client{
		Timeout: masterTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Read response body with size limit
	body := io.LimitReader(resp.Body, 50*1024*1024) // 50MB limit
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var apiServers []apiServerResponse
	if err := json.Unmarshal(data, &apiServers); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if len(apiServers) == 0 {
		return nil, errors.New("API returned zero servers")
	}

	// Convert API response to Server objects
	servers := make([]Server, 0, len(apiServers))
	for _, s := range apiServers {
		if s.IP == "" {
			continue
		}
		host, port := splitHostPort(s.IP)
		servers = append(servers, Server{
			Name:        s.Hostname,
			Host:        host,
			Port:        port,
			Players:     s.Players,
			MaxPlayers:  s.MaxPlayers,
			Passworded:  s.Password,
			Loading:     true,
			LastUpdated: time.Now(),
		})
	}

	return servers, nil
}

type fallbackServer struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

func LoadFallback(path string) ([]Server, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw []fallbackServer
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	servers := make([]Server, 0, len(raw))
	for _, s := range raw {
		servers = append(servers, Server{
			Name:    s.Name,
			Host:    s.Host,
			Port:    s.Port,
			Loading: true,
		})
	}
	return servers, nil
}

func DefaultFallbackPath() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "servers.json")
	}
	return filepath.Join(".", "servers.json")
}

func FetchServers(ctx context.Context, masterAddr string) ([]Server, error) {
	servers, err := FetchFromMaster(ctx, masterAddr)
	if err == nil {
		return servers, nil
	}
	fallback, ferr := LoadFallback(DefaultFallbackPath())
	if ferr != nil {
		return nil, err
	}
	return fallback, nil
}

func splitHostPort(value string) (string, int) {
	const defaultPort = 7777
	host, portStr, err := net.SplitHostPort(value)
	if err == nil {
		if port, perr := strconv.Atoi(portStr); perr == nil {
			return host, port
		}
		return host, defaultPort
	}
	if host, portStr, ok := strings.Cut(value, ":"); ok {
		if port, perr := strconv.Atoi(portStr); perr == nil {
			return host, port
		}
		return host, defaultPort
	}
	return value, defaultPort
}
