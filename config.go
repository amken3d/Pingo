package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig holds non-secret settings that are persisted to disk.
// API keys are NEVER stored here — use environment variables instead.
type AppConfig struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	OllamaHost  string  `json:"ollama_host,omitempty"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

func defaultConfig() AppConfig {
	return AppConfig{
		Provider:    "auto",
		Temperature: 0.7,
	}
}

// configDir returns ~/.config/pingo, creating it if needed.
func configDir() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	dir = filepath.Join(dir, "pingo")
	return dir, os.MkdirAll(dir, 0700)
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// loadConfig reads settings from disk. Returns defaults if file doesn't exist.
func loadConfig() AppConfig {
	cfg := defaultConfig()
	path, err := configPath()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	return cfg
}

// saveConfig writes non-secret settings to disk with restrictive permissions.
func saveConfig(cfg AppConfig) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
