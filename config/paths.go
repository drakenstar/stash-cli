package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName     = "stash-cli"
	ConfigFile  = "config.json"
	SessionFile = "session.json"
)

type Paths struct {
	ConfigPath  string
	SessionPath string
}

func DefaultPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, err
	}

	stateDir, err := userStateDir(home, configDir)
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		ConfigPath:  filepath.Join(configDir, AppName, ConfigFile),
		SessionPath: filepath.Join(stateDir, AppName, SessionFile),
	}, nil
}

func userStateDir(home, configDir string) (string, error) {
	switch runtime.GOOS {
	case "windows":
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return local, nil
		}
		return configDir, nil
	case "darwin":
		return configDir, nil
	default:
		if state := os.Getenv("XDG_STATE_HOME"); state != "" {
			return state, nil
		}
		return filepath.Join(home, ".local", "state"), nil
	}
}

func ConfigPathExists(paths Paths) (bool, error) {
	if paths.ConfigPath == "" {
		return false, nil
	}
	if _, err := os.Stat(paths.ConfigPath); err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	return false, nil
}
