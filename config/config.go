package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Config is the user-editable application configuration persisted to
// %APPDATA%/grails/config.json (or the OS equivalent).
type Config struct {
	Namespaces             []string `json:"namespaces"`
	PortRangeStart         int      `json:"portRangeStart"`
	PortRangeEnd           int      `json:"portRangeEnd"`
	GrpcPorts              []int    `json:"grpcPorts"`
	NodePortHost           string   `json:"nodePortHost"`
	TokenEndpoint          string   `json:"tokenEndpoint"`
	ClientID               string   `json:"clientId"`
	ServiceExcludePatterns []string `json:"serviceExcludePatterns"`
}

// Default returns a Config initialized with the documented defaults.
func Default() Config {
	return Config{
		Namespaces:             []string{"default", "am-dev", "am-qa", "am-demo"},
		PortRangeStart:         35000,
		PortRangeEnd:           60000,
		GrpcPorts:              []int{5001, 5002},
		NodePortHost:           "127.0.0.1",
		TokenEndpoint:          "",
		ClientID:               "",
		ServiceExcludePatterns: []string{"*wassups"},
	}
}

// configPath returns the absolute path to the config file.
func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating user config dir: %w", err)
	}
	return filepath.Join(dir, "grails", "config.json"), nil
}

// Load reads the config from disk, creating it with defaults if missing.
func Load() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Default(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			if saveErr := Save(cfg); saveErr != nil {
				log.Printf("[config.Load] Could not seed default config at %s: %v", path, saveErr)
			} else {
				log.Printf("[config.Load] Created default config at %s", path)
			}
			return cfg, nil
		}
		return Default(), fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), fmt.Errorf("parsing config %s: %w", path, err)
	}

	// Backfill any missing fields from defaults so old config files keep working.
	d := Default()
	if len(cfg.Namespaces) == 0 {
		cfg.Namespaces = d.Namespaces
	}
	if cfg.PortRangeStart == 0 {
		cfg.PortRangeStart = d.PortRangeStart
	}
	if cfg.PortRangeEnd == 0 {
		cfg.PortRangeEnd = d.PortRangeEnd
	}
	if len(cfg.GrpcPorts) == 0 {
		cfg.GrpcPorts = d.GrpcPorts
	}
	if cfg.NodePortHost == "" {
		cfg.NodePortHost = d.NodePortHost
	}

	return cfg, nil
}

// Save writes the config to disk, creating parent directories as needed.
func Save(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	log.Printf("[config.Save] Wrote config to %s", path)
	return nil
}
