package config

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

func DefaultConfig() Config {
	return Config{
		Nickname:     "Player",
		GTAPath:      "",
		OMPLauncher:  "",
		Runtime:      RuntimeAuto,
		MasterServer: "https://api.open.mp/servers",
		BrowseOnly:   false,
	}
}
