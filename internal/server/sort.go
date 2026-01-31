package server

import "sort"

type SortMode int

const (
	SortNone SortMode = iota
	SortPing
	SortPlayers
)

func SortServers(servers []Server, mode SortMode) {
	switch mode {
	case SortPing:
		sort.SliceStable(servers, func(i, j int) bool {
			// Push servers with 0 ping to the bottom
			if servers[i].Ping == 0 && servers[j].Ping == 0 {
				return false
			}
			if servers[i].Ping == 0 {
				return false
			}
			if servers[j].Ping == 0 {
				return true
			}
			return servers[i].Ping < servers[j].Ping
		})
	case SortPlayers:
		sort.SliceStable(servers, func(i, j int) bool {
			return servers[i].Players > servers[j].Players
		})
	default:
		return
	}
}
