package config

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DefaultEngine  string
	SystemPrompt   string
	PromptStrategy string
	PromptPreset   string
	Engines        map[string]EngineConfig
}

type EngineConfig struct {
	Command string
	Args    []string
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
	values, err := parseTOML(string(data))
	if err != nil {
		return cfg, err
	}
	if v, ok := values["engine"]; ok {
		cfg.DefaultEngine = v.str
	}
	if v, ok := values["prompt_strategy"]; ok {
		cfg.PromptStrategy = v.str
	}
	presetSet := false
	if v, ok := values["prompt_preset"]; ok {
		cfg.PromptPreset = v.str
		presetSet = true
	}
	if presetSet {
		prompt, err := LoadPromptPreset(cfg.PromptPreset)
		if err != nil {
			return cfg, err
		}
		cfg.SystemPrompt = prompt
	}
	if v, ok := values["system_prompt"]; ok {
		cfg.SystemPrompt = v.str
	}
	for key, v := range values {
		if !strings.HasPrefix(key, "engines.") {
			continue
		}
		parts := strings.Split(key, ".")
		if len(parts) < 3 {
			continue
		}
		name := parts[1]
		field := parts[2]
		engine := cfg.Engines[name]
		switch field {
		case "command":
			engine.Command = v.str
		case "args":
			engine.Args = v.list
		}
		cfg.Engines[name] = engine
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

type tomlValue struct {
	str  string
	list []string
}

func parseTOML(data string) (map[string]tomlValue, error) {
	values := make(map[string]tomlValue)
	scanner := bufio.NewScanner(strings.NewReader(data))
	section := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		fullKey := key
		if section != "" {
			fullKey = section + "." + key
		}
		parsed, err := parseValue(value)
		if err != nil {
			return nil, err
		}
		values[fullKey] = parsed
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan config: %w", err)
	}
	return values, nil
}

func parseValue(value string) (tomlValue, error) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		list, err := parseList(value[1 : len(value)-1])
		if err != nil {
			return tomlValue{}, err
		}
		return tomlValue{list: list}, nil
	}
	str, err := parseString(value)
	if err != nil {
		return tomlValue{}, err
	}
	return tomlValue{str: str}, nil
}

func parseList(value string) ([]string, error) {
	var items []string
	var current strings.Builder
	inQuote := byte(0)
	for i := 0; i < len(value); i++ {
		ch := value[i]
		switch ch {
		case '\'', '"':
			if inQuote == 0 {
				inQuote = ch
				continue
			}
			if inQuote == ch {
				inQuote = 0
				continue
			}
		case ',':
			if inQuote == 0 {
				item := strings.TrimSpace(current.String())
				if item != "" {
					items = append(items, item)
				}
				current.Reset()
				continue
			}
		}
		current.WriteByte(ch)
	}
	item := strings.TrimSpace(current.String())
	if item != "" {
		items = append(items, item)
	}
	for i, item := range items {
		parsed, err := parseString(strings.TrimSpace(item))
		if err != nil {
			return nil, err
		}
		items[i] = parsed
	}
	return items, nil
}

func parseString(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) || (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return value[1 : len(value)-1], nil
	}
	return value, nil
}
