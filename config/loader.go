package config

import (
	"os"
)

// ConfigLoader defines the interface for loading configuration
type ConfigLoader interface {
	Load() (*Config, error)
}

// FileConfigLoader loads configuration from environment variables with file fallback
type FileConfigLoader struct{}

// NewFileConfigLoader creates a new file-based config loader
func NewFileConfigLoader() ConfigLoader {
	return &FileConfigLoader{}
}

// Load loads configuration from environment variables, falling back to defaults
func (l *FileConfigLoader) Load() (*Config, error) {
	cfg := DefaultConfig()

	// Load from environment variables if set
	if userSeedPath := os.Getenv("USER_SEED_DATA_PATH"); userSeedPath != "" {
		cfg.UserSeedDataPath = userSeedPath
	}

	if callDataPath := os.Getenv("CALL_DATA_PATH"); callDataPath != "" {
		cfg.CallDataPath = callDataPath
	}

	if serverPort := os.Getenv("SERVER_PORT"); serverPort != "" {
		cfg.ServerPort = serverPort
	}

	// Note: For simplicity, we're using defaults for numeric values
	// In production, you might want to parse env vars for these too

	return cfg, nil
}

// Load is a convenience function that loads configuration using the default loader
func Load() *Config {
	loader := NewFileConfigLoader()
	cfg, err := loader.Load()
	if err != nil {
		// Return defaults if loading fails
		return DefaultConfig()
	}
	return cfg
}
