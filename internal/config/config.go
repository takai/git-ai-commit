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
	return Config{
		DefaultEngine:  "",
		SystemPrompt:   "",
		PromptStrategy: "append",
		PromptPreset:   defaultPromptPreset,
		Engines:        map[string]EngineConfig{},
	}
}

func Load() (Config, error) {
	cfg := Default()

	// 1. Load user config
	userPath, err := configPath()
	if err != nil {
		return cfg, err
	}
	if data, err := os.ReadFile(userPath); err == nil {
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parse user config: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return cfg, fmt.Errorf("read user config: %w", err)
	}

	// 2. Load repo config (merges on top of user config)
	repoPath, err := repoConfigPath()
	if err != nil {
		return cfg, err
	}
	if repoPath != "" {
		if data, err := os.ReadFile(repoPath); err == nil {
			if err := toml.Unmarshal(data, &cfg); err != nil {
				return cfg, fmt.Errorf("parse repo config: %w", err)
			}
		} else {
			return cfg, fmt.Errorf("read repo config: %w", err)
		}
	}

	// 3. Auto-detect engine if still empty
	if strings.TrimSpace(cfg.DefaultEngine) == "" {
		if auto := autodetectEngine(); auto != "" {
			cfg.DefaultEngine = auto
		}
	}

	// 4. Ensure Engines map is initialized
	if cfg.Engines == nil {
		cfg.Engines = map[string]EngineConfig{}
	}

	// 5. Load prompt from preset if not explicitly set
	if strings.TrimSpace(cfg.SystemPrompt) == "" {
		preset := cfg.PromptPreset
		if strings.TrimSpace(preset) == "" {
			preset = defaultPromptPreset
		}
		prompt, err := LoadPromptPreset(preset)
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
