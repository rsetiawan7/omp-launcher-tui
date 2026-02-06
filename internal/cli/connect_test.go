package cli

import (
	"testing"
)

func TestParseAddress(t *testing.T) {
	tests := []struct {
		input       string
		wantHost    string
		wantPort    int
		wantErr     bool
		description string
	}{
		{"127.0.0.1", "127.0.0.1", 7777, false, "IP without port should default to 7777"},
		{"127.0.0.1:7777", "127.0.0.1", 7777, false, "IP with port 7777"},
		{"127.0.0.1:8888", "127.0.0.1", 8888, false, "IP with custom port"},
		{"example.com", "example.com", 7777, false, "Hostname without port should default to 7777"},
		{"example.com:9999", "example.com", 9999, false, "Hostname with custom port"},
		{"", "", 0, true, "Empty string should error"},
		{"127.0.0.1:invalid", "", 0, true, "Invalid port should error"},
		{"127.0.0.1:99999", "", 0, true, "Port out of range should error"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			host, port, err := ParseAddress(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseAddress(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseAddress(%q) unexpected error: %v", tt.input, err)
				}
				if host != tt.wantHost {
					t.Errorf("ParseAddress(%q) host = %q, want %q", tt.input, host, tt.wantHost)
				}
				if port != tt.wantPort {
					t.Errorf("ParseAddress(%q) port = %d, want %d", tt.input, port, tt.wantPort)
				}
			}
		})
	}
}
