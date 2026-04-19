package config

import (
	"bytes"
	"embed"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DefaultEngine string                  `toml:"engine"`
	Prompt        string                  `toml:"prompt"`
	PromptFile    string                  `toml:"prompt_file"`
	Engines       map[string]EngineConfig `toml:"engines"`
	Filter        FilterConfig            `toml:"filter"`

	// ResolvedPrompt holds the final prompt text after loading from preset or file.
	// This is not read from config files directly.
	ResolvedPrompt string `toml:"-"`
}

// FilterConfig holds diff filtering configuration.
type FilterConfig struct {
	MaxFileLines           int      `toml:"max_file_lines"`           // Max lines per file (0 = use default)
	DefaultExcludePatterns []string `toml:"default_exclude_patterns"` // Override built-in defaults
	ExcludePatterns        []string `toml:"exclude_patterns"`         // Additional patterns to exclude
}

// rawConfig is the TOML structure used to detect mutual exclusivity in a single layer.
type rawConfig struct {
	DefaultEngine string                  `toml:"engine"`
	Prompt        string                  `toml:"prompt"`
	PromptFile    string                  `toml:"prompt_file"`
	Engines       map[string]EngineConfig `toml:"engines"`
	Filter        FilterConfig            `toml:"filter"`
}

type EngineConfig struct {
	Args []string `toml:"args"`
}

//go:embed assets/*.md
var promptFS embed.FS

const defaultPromptPreset = "default"

// DefaultEngineArgs provides default CLI arguments for known engines.
var DefaultEngineArgs = map[string][]string{
	"codex":        {"exec", "--model", "gpt-5.4-mini"},
	"claude":       {"-p", "--model", "haiku", "--settings", "{\"attribution\":{\"commit\":\"\",\"pr\":\"\"}}", "--no-session-persistence"},
	"cursor-agent": {"-p"},
	"gemini":       {"-m", "gemini-2.5-flash", "-p", "{{prompt}}"},
}

func Default() Config {
	return Config{
		DefaultEngine: "",
		Prompt:        "",
		PromptFile:    "",
		Engines:       map[string]EngineConfig{},
	}
}

func Load() (Config, error) {
	cfg := Default()
	var promptFilePath string // resolved path used by resolvePromptFromPath
	var promptFileRepoRoot string

	// Read all git config scopes in a single call.
	gitScopes, err := readGitConfigScopes()
	if err != nil {
		return cfg, err
	}

	// Determine repo root once (needed for local/worktree promptFile resolution
	// and for loading the repo TOML).
	repoRoot, repoPath, err := repoConfigPath()
	if err != nil {
		return cfg, err
	}

	// Determine home directory (needed for global promptFile resolution).
	homeDir, _ := os.UserHomeDir()

	// 1. System git config (lowest priority)
	if err := applyGitConfigScope(&cfg, gitScopes.system, "system git config", ""); err != nil {
		return cfg, err
	}
	promptFilePath, promptFileRepoRoot = gitScopePromptPaths(gitScopes.system, promptFilePath, promptFileRepoRoot, cfg)

	// 2. Global git config
	if err := applyGitConfigScope(&cfg, gitScopes.global, "user git config", homeDir); err != nil {
		return cfg, err
	}
	promptFilePath, promptFileRepoRoot = gitScopePromptPaths(gitScopes.global, promptFilePath, promptFileRepoRoot, cfg)

	// 3. User TOML config
	userPath, err := configPath()
	if err != nil {
		return cfg, err
	}
	if data, err := os.ReadFile(userPath); err == nil {
		if err := loadConfigLayer(data, &cfg, "user config"); err != nil {
			return cfg, err
		}
		if cfg.PromptFile != "" {
			promptFilePath = userPath
			promptFileRepoRoot = ""
		} else if cfg.Prompt != "" {
			promptFilePath = ""
			promptFileRepoRoot = ""
		}
	} else if !os.IsNotExist(err) {
		return cfg, fmt.Errorf("read user config: %w", err)
	}

	// 4. Repo TOML config
	if repoPath != "" {
		data, trusted, err := loadTrustedRepoConfig(repoRoot, repoPath)
		if err != nil {
			return cfg, err
		}
		if trusted {
			var repoCfg rawConfig
			if err := toml.Unmarshal(data, &repoCfg); err != nil {
				return cfg, fmt.Errorf("parse repo config: %w", err)
			}
			if err := validatePromptExclusivity(repoCfg.Prompt, repoCfg.PromptFile, "repo config"); err != nil {
				return cfg, err
			}
			if repoCfg.DefaultEngine != "" {
				cfg.DefaultEngine = repoCfg.DefaultEngine
			}
			if repoCfg.Prompt != "" {
				cfg.Prompt = repoCfg.Prompt
				cfg.PromptFile = ""
				promptFilePath = ""
				promptFileRepoRoot = ""
			}
			if repoCfg.PromptFile != "" {
				cfg.PromptFile = repoCfg.PromptFile
				cfg.Prompt = ""
				promptFilePath = repoPath
				promptFileRepoRoot = repoRoot
			}
			if repoCfg.Engines != nil {
				if cfg.Engines == nil {
					cfg.Engines = map[string]EngineConfig{}
				}
				maps.Copy(cfg.Engines, repoCfg.Engines)
			}
			if repoCfg.Filter.MaxFileLines != 0 {
				cfg.Filter.MaxFileLines = repoCfg.Filter.MaxFileLines
			}
			if len(repoCfg.Filter.DefaultExcludePatterns) > 0 {
				cfg.Filter.DefaultExcludePatterns = repoCfg.Filter.DefaultExcludePatterns
			}
			if len(repoCfg.Filter.ExcludePatterns) > 0 {
				cfg.Filter.ExcludePatterns = append(cfg.Filter.ExcludePatterns, repoCfg.Filter.ExcludePatterns...)
			}
		}
	}

	// 5. Local git config (overrides repo TOML)
	localPromptFileBase := repoRoot // may be "" if not in a repo
	if err := applyGitConfigScope(&cfg, gitScopes.local, "repo git config", localPromptFileBase); err != nil {
		return cfg, err
	}
	promptFilePath, promptFileRepoRoot = gitScopePromptPaths(gitScopes.local, promptFilePath, promptFileRepoRoot, cfg)

	// 6. Worktree git config (highest git config priority)
	if err := applyGitConfigScope(&cfg, gitScopes.worktree, "worktree git config", localPromptFileBase); err != nil {
		return cfg, err
	}
	promptFilePath, promptFileRepoRoot = gitScopePromptPaths(gitScopes.worktree, promptFilePath, promptFileRepoRoot, cfg)

	// 7. Auto-detect engine if still empty
	if strings.TrimSpace(cfg.DefaultEngine) == "" {
		if auto := autodetectEngine(); auto != "" {
			cfg.DefaultEngine = auto
		}
	}

	// 8. Ensure Engines map is initialized
	if cfg.Engines == nil {
		cfg.Engines = map[string]EngineConfig{}
	}

	// 9. Resolve prompt from preset or file.
	if err := resolvePromptFromPath(&cfg, promptFilePath, promptFileRepoRoot); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func loadConfigLayer(data []byte, cfg *Config, source string) error {
	var raw rawConfig
	if err := toml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse %s: %w", source, err)
	}
	if err := validatePromptExclusivity(raw.Prompt, raw.PromptFile, source); err != nil {
		return err
	}
	// Merge into cfg
	if raw.DefaultEngine != "" {
		cfg.DefaultEngine = raw.DefaultEngine
	}
	if raw.Prompt != "" {
		cfg.Prompt = raw.Prompt
		cfg.PromptFile = ""
	}
	if raw.PromptFile != "" {
		cfg.PromptFile = raw.PromptFile
		cfg.Prompt = ""
	}
	if raw.Engines != nil {
		if cfg.Engines == nil {
			cfg.Engines = map[string]EngineConfig{}
		}
		maps.Copy(cfg.Engines, raw.Engines)
	}
	// Merge filter config
	if raw.Filter.MaxFileLines != 0 {
		cfg.Filter.MaxFileLines = raw.Filter.MaxFileLines
	}
	if len(raw.Filter.DefaultExcludePatterns) > 0 {
		cfg.Filter.DefaultExcludePatterns = raw.Filter.DefaultExcludePatterns
	}
	if len(raw.Filter.ExcludePatterns) > 0 {
		cfg.Filter.ExcludePatterns = append(cfg.Filter.ExcludePatterns, raw.Filter.ExcludePatterns...)
	}
	return nil
}

func validatePromptExclusivity(prompt, promptFile, source string) error {
	if strings.TrimSpace(prompt) != "" && strings.TrimSpace(promptFile) != "" {
		return fmt.Errorf("%s: cannot set both 'prompt' and 'prompt_file'", source)
	}
	return nil
}

func resolvePromptFromPath(cfg *Config, promptFilePath, promptFileRepoRoot string) error {
	// If prompt_file is set, load from file (relative to config file's directory)
	if strings.TrimSpace(cfg.PromptFile) != "" {
		promptText, err := loadPromptFile(cfg.PromptFile, promptFilePath, promptFileRepoRoot)
		if err != nil {
			return err
		}
		cfg.ResolvedPrompt = promptText
		return nil
	}

	// If prompt preset is set, load from embedded assets
	preset := cfg.Prompt
	if strings.TrimSpace(preset) == "" {
		preset = defaultPromptPreset
	}
	promptText, err := LoadPromptPreset(preset)
	if err != nil {
		return err
	}
	cfg.ResolvedPrompt = promptText
	return nil
}

func loadPromptFile(promptFile, promptFilePath, promptFileRepoRoot string) (string, error) {
	var basePath string
	if promptFilePath != "" {
		basePath = filepath.Dir(promptFilePath)
	} else {
		// Fallback to current directory
		basePath = "."
	}
	fullPath := promptFile
	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(basePath, promptFile)
	}
	if promptFileRepoRoot != "" {
		if filepath.IsAbs(promptFile) {
			return "", fmt.Errorf("prompt_file must be within repo root")
		}
		allowed, err := isPathWithinRoot(fullPath, promptFileRepoRoot)
		if err != nil {
			return "", err
		}
		if !allowed {
			return "", fmt.Errorf("prompt_file must be within repo root")
		}
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("read prompt file %q: %w", fullPath, err)
	}
	return strings.TrimSpace(string(data)), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

var repoConfigNames = []string{"git-ai-commit.toml", ".git-ai-commit.toml"}

func repoConfigPath() (string, string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", "", nil
	}
	root := strings.TrimSpace(stdout.String())
	if root == "" {
		return "", "", nil
	}
	for _, name := range repoConfigNames {
		path := filepath.Join(root, name)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", "", fmt.Errorf("stat repo config: %w", err)
		}
		if info.IsDir() {
			return "", "", fmt.Errorf("repo config is a directory: %s", path)
		}
		return root, path, nil
	}
	return root, "", nil
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

// ApplyCLIPrompt applies CLI-level prompt or prompt_file overrides.
// Both prompt and promptFile should not be set at the same time (caller must validate).
// If promptFile is set, it is resolved relative to the current working directory.
func ApplyCLIPrompt(cfg *Config, prompt, promptFile string) error {
	if strings.TrimSpace(prompt) != "" {
		cfg.Prompt = prompt
		cfg.PromptFile = ""
		promptText, err := LoadPromptPreset(prompt)
		if err != nil {
			return err
		}
		cfg.ResolvedPrompt = promptText
		return nil
	}
	if strings.TrimSpace(promptFile) != "" {
		cfg.PromptFile = promptFile
		cfg.Prompt = ""
		fullPath := promptFile
		if !filepath.IsAbs(fullPath) {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			fullPath = filepath.Join(cwd, promptFile)
		}
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read prompt file %q: %w", fullPath, err)
		}
		cfg.ResolvedPrompt = strings.TrimSpace(string(data))
		return nil
	}
	return nil
}

// ValidateCLIPromptExclusivity checks that prompt and prompt_file are not both set at CLI level.
func ValidateCLIPromptExclusivity(prompt, promptFile string) error {
	return validatePromptExclusivity(prompt, promptFile, "CLI")
}
