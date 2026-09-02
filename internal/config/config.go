package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config — конфігурація користувача.
// Зберігається у ~/.fire-station/config.yaml з правами 0600.
type Config struct {
	DBPath    string `yaml:"db_path"`
	FontPath  string `yaml:"font_path"`
	AIKey     string `yaml:"ai_key"`
	AIModel   string `yaml:"ai_model"`
	AIBaseURL string `yaml:"ai_base_url"` // OpenAI-сумісний endpoint (Groq, aimlapi, Ollama…)

	path string `yaml:"-"`
}

func defaults() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		DBPath:   "data/fire_station.accdb",
		FontPath: "assets/fonts/DejaVuSans.ttf",
		path:     filepath.Join(home, ".fire-station", "config.yaml"),
	}
}

// Load читає конфіг з диска поверх дефолтів; відсутній файл — не помилка.
func Load() (*Config, error) {
	c := defaults()
	data, err := os.ReadFile(c.path)
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Save записує конфіг на диск (0600).
func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0600)
}
