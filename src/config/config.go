package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Domain       string
	Email        string
	APIToken     string
	SpaceID      string
	ParentPageID string
}

func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &config, nil
}
