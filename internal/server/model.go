package server

import (
	"fmt"
	"net"
	"time"
)

type Server struct {
	Name        string            `json:"name"`
	Alias       string            `json:"alias,omitempty"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Players     int               `json:"players"`
	MaxPlayers  int               `json:"max_players"`
	Ping        time.Duration     `json:"ping"`
	Passworded  bool              `json:"passworded"`
	LastUpdated time.Time         `json:"last_updated"`
	Loading     bool              `json:"-"`
	Rules       map[string]string `json:"rules,omitempty"`
}

func (s Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s Server) UDPAddr() (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", s.Addr())
}
