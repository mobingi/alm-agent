package util

import "github.com/BurntSushi/toml"

var allowedPlatforms = []string{"aws"}

// UpdateSetting includes self update information
type updateSetting struct {
	Platform        string
	VersionURL      string
	UpdateScriptURL string
}

// Updater includes self update information
type Updater struct {
	Setting updateSetting
}

func LoadSetting(b []byte) (*updateSetting, error) {
	var setting updateSetting
	_, err := toml.Decode(string(b), &setting)
	if err != nil {
		//
		return &setting, err
	}

	return &setting, nil
}
