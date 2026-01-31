package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const MasterListFile = "master_lists.json"

type MasterList struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

type MasterLists struct {
	Lists []MasterList `json:"lists"`
}

func MasterListPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, MasterListFile), nil
}

func LoadMasterLists() (MasterLists, error) {
	path, err := MasterListPath()
	if err != nil {
		return MasterLists{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default with Open.MP
			return MasterLists{
				Lists: []MasterList{
					{
						Name:        "Open.MP Official",
						Host:        "https://api.open.mp/servers",
						Description: "Official Open.MP master server list",
						Active:      true,
					},
				},
			}, nil
		}
		return MasterLists{}, err
	}

	var lists MasterLists
	if err := json.Unmarshal(data, &lists); err != nil {
		return MasterLists{}, err
	}

	return lists, nil
}

func SaveMasterLists(lists MasterLists) error {
	path, err := MasterListPath()
	if err != nil {
		return err
	}

	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, DefaultPerms); err != nil {
		return err
	}

	data, err := json.MarshalIndent(lists, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func GetActiveMasterList() (string, error) {
	lists, err := LoadMasterLists()
	if err != nil {
		return "", err
	}

	for _, list := range lists.Lists {
		if list.Active {
			return list.Host, nil
		}
	}

	// Fallback to default
	return "https://api.open.mp/servers", nil
}

func GetActiveMasterListName() (string, error) {
	lists, err := LoadMasterLists()
	if err != nil {
		return "", err
	}

	for _, list := range lists.Lists {
		if list.Active {
			return list.Name, nil
		}
	}

	// Fallback to default
	return "Open.MP Official", nil
}
