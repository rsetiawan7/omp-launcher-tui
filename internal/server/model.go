package server

import (
	"fmt"
	"net"
	"time"
)

type Server struct {
	Name        string
	Host        string
	Port        int
	Players     int
	MaxPlayers  int
	Ping        time.Duration
	Passworded  bool
	LastUpdated time.Time
	Loading     bool
}

func (s Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s Server) UDPAddr() (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", s.Addr())
}
