package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DefaultEngine  string
	SystemPrompt   string
	PromptStrategy string
	Engines        map[string]EngineConfig
}

type EngineConfig struct {
	Command string
	Args    []string
}

const defaultSystemPrompt = `Generate a single-line commit message following the Conventional Commits specification.

Rules:
- IMPORTANT: Output only the commit message
- Format: type: short summary
- One line only
- Types allowed: feat, fix, docs, style, refactor, test, chore
- English only
- No trailing period
- Summary under 72 characters
- Do not wrap the message in quotes or backticks
- Make sure with version bump if needed

Write the commit message based on the following git diff:`

func Default() Config {
	return Config{
		DefaultEngine:  "codex",
		SystemPrompt:   defaultSystemPrompt,
		PromptStrategy: "append",
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
	if v, ok := values["system_prompt"]; ok {
		cfg.SystemPrompt = v.str
	}
	if v, ok := values["prompt_strategy"]; ok {
		cfg.PromptStrategy = v.str
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
