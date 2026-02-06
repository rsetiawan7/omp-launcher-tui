package config

import (
	"fmt"
	"math/rand"
	"time"
)

type Runtime string

const (
	RuntimeAuto   Runtime = "auto"
	RuntimeWine   Runtime = "wine"
	RuntimeProton Runtime = "proton"
	RuntimeNative Runtime = "native"
)

type Config struct {
	Nickname     string  `json:"nickname"`
	GTAPath      string  `json:"gta_path"`
	OMPLauncher  string  `json:"omp_launcher"`
	Runtime      Runtime `json:"runtime"`
	MasterServer string  `json:"master_server"`
	BrowseOnly   bool    `json:"browse_only"`
}

// generateRandomNickname generates a random nickname following SA-MP rules:
// - Length: 3-20 characters
// - Can contain: letters (a-z, A-Z), numbers (0-9), underscores (_), brackets ([]), dots (.)
// - Must start with a letter
func generateRandomNickname() string {
	rand.Seed(time.Now().UnixNano())

	// Common prefixes for player names
	prefixes := []string{
		"Player", "Gamer", "Pro", "Noob", "Epic", "Cool", "Dark", "Shadow",
		"Fire", "Ice", "Thunder", "Storm", "Night", "Day", "Mega", "Super",
		"Alpha", "Beta", "Delta", "Omega", "Phantom", "Ghost", "Specter",
	}

	// Common suffixes
	suffixes := []string{
		"", "X", "Z", "YT", "TV", "Pro", "HD", "4K", "Gaming", "Plays",
	}

	prefix := prefixes[rand.Intn(len(prefixes))]
	suffix := suffixes[rand.Intn(len(suffixes))]
	number := rand.Intn(1000)

	// Generate different formats
	formats := []string{
		prefix,
		fmt.Sprintf("%s%d", prefix, number),
		fmt.Sprintf("%s_%d", prefix, number),
		fmt.Sprintf("%s%s", prefix, suffix),
		fmt.Sprintf("%s_%s", prefix, suffix),
		fmt.Sprintf("%s%s%d", prefix, suffix, rand.Intn(100)),
		fmt.Sprintf("[%s]%d", prefix, number),
	}

	name := formats[rand.Intn(len(formats))]

	// Ensure it's between 3-20 characters
	if len(name) > 20 {
		name = name[:20]
	}
	if len(name) < 3 {
		name = fmt.Sprintf("Player%d", rand.Intn(1000))
	}

	return name
}

func DefaultConfig() Config {
	return Config{
		Nickname:     generateRandomNickname(),
		GTAPath:      "",
		OMPLauncher:  "",
		Runtime:      RuntimeAuto,
		MasterServer: "https://api.open.mp/servers",
		BrowseOnly:   false,
	}
}
