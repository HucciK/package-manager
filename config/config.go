package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	User       string `json:"user"`
	Password   string `json:"password"`
	Address    string `json:"address"`
	PacketsDir string `json:"packets_path"`
}

func NewConfig() (Config, error) {
	var cfg Config

	content, err := os.ReadFile("./config/config.json")
	if err != nil {
		return cfg, fmt.Errorf("can't read config file: %w", err)
	}

	if err := json.Unmarshal(content, &cfg); err != nil {
		return cfg, fmt.Errorf("can't unmarshall config file content: %w", err)
	}

	return cfg, nil
}
