package config

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
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
	Args []string `toml:"args"`
}

//go:embed assets/*.md
var promptFS embed.FS

const defaultPromptPreset = "default"

// DefaultEngineArgs provides default CLI arguments for known engines.
var DefaultEngineArgs = map[string][]string{
	"codex":        {"exec"},
	"claude":       {"-p", "--model", "haiku"},
	"cursor-agent": {"-p"},
	"gemini":       {"-m", "gemini-2.5-flash", "-p", "{{prompt}}"},
}

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
	if repoPath, err := repoConfigPath(); err != nil {
		return cfg, err
	} else if repoPath != "" {
		return loadFromPath(cfg, repoPath)
	}
	path, err := configPath()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if auto := autodetectEngine(); auto != "" {
				cfg.DefaultEngine = auto
			}
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

func repoConfigPath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", nil
	}
	root := strings.TrimSpace(stdout.String())
	if root == "" {
		return "", nil
	}
	path := filepath.Join(root, ".git-ai-commit.toml")
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("stat repo config: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("repo config is a directory: %s", path)
	}
	return path, nil
}

func loadFromPath(cfg Config, path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
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

func autodetectEngine() string {
	candidates := []string{"claude", "gemini", "codex"}
	for _, name := range candidates {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return ""
}
