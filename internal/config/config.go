package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config — налаштування програми (~/.fire-station/config.yaml).
// Містить API-ключ, тому файл записується з правами 0600.
type Config struct {
	DBPath   string `yaml:"db_path"`
	FontPath string `yaml:"font_path"`
	AIKey    string `yaml:"ai_key"`
	AIModel  string `yaml:"ai_model"`
}

func (c *Config) applyDefaults() {
	if c.DBPath == "" {
		c.DBPath = filepath.Join("data", "fire_station.accdb")
	}
	if c.FontPath == "" {
		c.FontPath = filepath.Join("assets", "fonts", "DejaVuSans.ttf")
	}
	if c.AIModel == "" {
		c.AIModel = "openai/gpt-5-5"
	}
}

// Path повертає шлях до файлу конфігурації.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".fire-station", "config.yaml"), nil
}

// Load читає конфігурацію. Відсутній файл — не помилка (дефолти).
func Load() (*Config, error) {
	c := &Config{}
	p, err := Path()
	if err != nil {
		c.applyDefaults()
		return c, nil
	}
	data, err := os.ReadFile(p)
	if err == nil {
		_ = yaml.Unmarshal(data, c)
	}
	c.applyDefaults()
	return c, nil
}

// Save записує конфігурацію з правами 0600.
func (c *Config) Save() error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}
