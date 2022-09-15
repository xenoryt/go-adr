package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
)

const configFileName = ".go-adr.json"

type Config struct {
	// The directory of ADR files
	Dir string `json:"dir"`

	// The directory of the config file. This is used to calculate relative paths stored in this config file.
	cfgDir string `json:"-"`
}

// AbsDir returns the absolute path to cfg.Dir
func (cfg Config) AbsDir() string {
	if path.IsAbs(cfg.Dir) {
		return cfg.Dir
	}
	return path.Join(cfg.cfgDir, cfg.Dir)
}

// ConfigFilePath returns rooted path to the config file.
// Errors if unable to find config file.
func ConfigFilePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	filepath := configFilePathFinder(cwd)
	if filepath == "" {
		return "", errors.New("Failed to find config file")
	}
	return filepath, nil
}

// ReadConfig reads the config file.
func ReadConfig() (cfg *Config, err error) {
	filepath, err := ConfigFilePath()
	if err != nil {
		return
	}

	contents, err := os.ReadFile(filepath)
	if err != nil {
		return
	}

	err = json.Unmarshal(contents, &cfg)
	if err != nil {
		return
	}
	cfg.cfgDir = path.Dir(filepath)
	return
}

// InitConfigFile initializes the config file.
func InitConfigFile(cfg *Config) error {
	// Check if file already exists
	if _, err := os.Stat(configFileName); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("Error writing config file: %w", fs.ErrExist)
	}
	return writeConfigFile(cfg, configFileName)
}

// UpdateConfig writes the config to the existing config file.
// Will error if config file does not exist.
func UpdateConfig(cfg *Config) error {
	filepath, err := ConfigFilePath()
	if err != nil {
		return err
	}
	return writeConfigFile(cfg, filepath)
}

// writeConfigFile writes the config to a file.
func writeConfigFile(cfg *Config, filepath string) error {
	contents, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, contents, 0644)
}

func configFilePathFinder(dir string) string {
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		if !f.IsDir() && f.Name() == configFileName {
			return path.Join(dir, f.Name())
		}
	}
	parent := path.Dir(dir)
	if parent == "/" || parent == "." {
		// Hit the root and unable to go further up.
		return ""
	}
	return configFilePathFinder(parent)
}
