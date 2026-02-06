package server

import (
	"context"
	"net"
	"strconv"
	"time"

	sampquery "github.com/Southclaws/go-samp-query"
)

const (
	queryTimeout = 1500 * time.Millisecond
)

func QueryServer(ctx context.Context, host string, port int) (Server, error) {
	addr := net.JoinHostPort(host, itoa(port))
	query, err := sampquery.NewQuery(addr)
	if err != nil {
		return Server{}, err
	}
	defer query.Close()

	deadline := time.Now().Add(queryTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	info, err := query.GetInfo(ctx, true)
	if err != nil {
		return Server{}, err
	}
	ping, err := query.GetPing(ctx)
	if err != nil {
		ping = 0
	}

	server := Server{
		Name:        info.Hostname,
		Host:        host,
		Port:        port,
		Players:     info.Players,
		MaxPlayers:  info.MaxPlayers,
		Passworded:  info.Password,
		Ping:        ping,
		Loading:     false,
		LastUpdated: time.Now(),
	}
	return server, nil
}

// QueryServerWithRules queries server info, ping, and rules in one call
func QueryServerWithRules(ctx context.Context, host string, port int) (Server, error) {
	addr := net.JoinHostPort(host, itoa(port))
	query, err := sampquery.NewQuery(addr)
	if err != nil {
		return Server{}, err
	}
	defer query.Close()

	deadline := time.Now().Add(queryTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	info, err := query.GetInfo(ctx, true)
	if err != nil {
		return Server{}, err
	}
	ping, err := query.GetPing(ctx)
	if err != nil {
		ping = 0
	}

	// Fetch rules
	rules, err := query.GetRules(ctx)
	if err != nil {
		// Continue without rules if they fail to fetch
		rules = nil
	}

	server := Server{
		Name:        info.Hostname,
		Host:        host,
		Port:        port,
		Players:     info.Players,
		MaxPlayers:  info.MaxPlayers,
		Passworded:  info.Password,
		Ping:        ping,
		Loading:     false,
		LastUpdated: time.Now(),
		Rules:       rules,
	}
	return server, nil
}

func QueryServerRules(ctx context.Context, host string, port int) (map[string]string, error) {
	addr := net.JoinHostPort(host, itoa(port))
	query, err := sampquery.NewQuery(addr)
	if err != nil {
		return nil, err
	}
	defer query.Close()

	deadline := time.Now().Add(queryTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	rules, err := query.GetRules(ctx)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func QueryServerPlayers(ctx context.Context, host string, port int) ([]string, error) {
	addr := net.JoinHostPort(host, itoa(port))
	query, err := sampquery.NewQuery(addr)
	if err != nil {
		return nil, err
	}
	defer query.Close()

	deadline := time.Now().Add(queryTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	players, err := query.GetPlayers(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
