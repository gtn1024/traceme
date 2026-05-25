package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	IntervalSeconds int    `toml:"interval_seconds"`
	Model           Model  `toml:"model"`
	Storage         Storage `toml:"storage"`
	Log             Log    `toml:"log"`
}

type Model struct {
	BaseURL         string `toml:"base_url"`
	Model           string `toml:"model"`
	APIKey          string `toml:"api_key"`
	TimeoutSeconds  int    `toml:"timeout_seconds"`
}

type Storage struct {
	Root string `toml:"storage_root"`
}

type Log struct {
	Level string `toml:"level"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		IntervalSeconds: 300,
		Model: Model{
			BaseURL:        "http://127.0.0.1:1234/v1",
			Model:          "gemma-4-e4b",
			APIKey:         "",
			TimeoutSeconds: 60,
		},
		Storage: Storage{
			Root: filepath.Join(home, ".traceme"),
		},
		Log: Log{
			Level: "info",
		},
	}
}

func (c *Config) LogsDir() string {
	return filepath.Join(c.Storage.Root, "logs")
}

func (c *Config) ConfigPath() string {
	return filepath.Join(c.Storage.Root, "config.toml")
}

func (c *Config) Save() error {
	if err := os.MkdirAll(c.Storage.Root, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	f, err := os.Create(c.ConfigPath())
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	return enc.Encode(c)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func LoadDefault() (*Config, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".traceme", "config.toml")
	return Load(path)
}
