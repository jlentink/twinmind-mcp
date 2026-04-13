package config

import (
	"os"
	"path/filepath"
)

const appName = "twinmind"

func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", appName)
}

func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}
