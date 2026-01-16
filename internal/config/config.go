package config

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DefaultEngine  string                  `toml:"engine"`
	SystemPrompt   string                  `toml:"system_prompt"`
	PromptStrategy string                  `toml:"prompt_strategy"`
	PromptPreset   string                  `toml:"prompt_preset"`
	Engines        map[string]EngineConfig `toml:"engines"`
}

type EngineConfig struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

//go:embed assets/*.md
var promptFS embed.FS

const defaultPromptPreset = "default"

func Default() Config {
	prompt, _ := LoadPromptPreset(defaultPromptPreset)
	return Config{
		DefaultEngine:  "codex",
		SystemPrompt:   prompt,
		PromptStrategy: "append",
		PromptPreset:   defaultPromptPreset,
		Engines:        map[string]EngineConfig{},
	}
}

func Load() (Config, error) {
	cfg := Default()
	path, err := configPath()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Engines == nil {
		cfg.Engines = map[string]EngineConfig{}
	}
	if strings.TrimSpace(cfg.PromptPreset) != "" && strings.TrimSpace(cfg.SystemPrompt) == "" {
		prompt, err := LoadPromptPreset(cfg.PromptPreset)
		if err != nil {
			return cfg, err
		}
		cfg.SystemPrompt = prompt
	}
	return cfg, nil
}

func configPath() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "git-ai-commit", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".config", "git-ai-commit", "config.toml"), nil
}

func LoadPromptPreset(name string) (string, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "", fmt.Errorf("prompt preset is empty")
	}
	path := fmt.Sprintf("assets/%s.md", name)
	data, err := promptFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("prompt preset %q not found", name)
	}
	return strings.TrimSpace(string(data)), nil
}
