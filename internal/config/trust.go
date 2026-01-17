package config

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type trustedRepoList struct {
	Entries []trustedRepoEntry `json:"entries"`
}

type trustedRepoEntry struct {
	RepoRoot   string `json:"repo_root"`
	ConfigPath string `json:"config_path"`
	Hash       string `json:"hash"`
}

func loadTrustedRepoConfig(repoRoot, repoConfigPath string) ([]byte, bool, error) {
	repoRootReal, err := realPath(repoRoot)
	if err != nil {
		return nil, false, fmt.Errorf("resolve repo root: %w", err)
	}
	repoConfigReal, err := realPath(repoConfigPath)
	if err != nil {
		return nil, false, fmt.Errorf("resolve repo config: %w", err)
	}

	data, err := os.ReadFile(repoConfigPath)
	if err != nil {
		return nil, false, fmt.Errorf("read repo config: %w", err)
	}
	hash := computeHash(data)

	trustPath, err := trustedRepoListPath()
	if err != nil {
		return nil, false, err
	}
	list, err := loadTrustedRepoList(trustPath)
	if err != nil {
		return nil, false, err
	}
	entry, index, ok := findTrustedEntry(list, repoRootReal, repoConfigReal)
	changed := ok && entry.Hash != hash
	if ok && !changed {
		return data, true, nil
	}
	if !isInteractiveStdin() {
		if changed {
			return nil, false, fmt.Errorf("repo config changed: %s", repoConfigPath)
		}
		return nil, false, fmt.Errorf("untrusted repo config: %s", repoConfigPath)
	}

	if !promptTrust(repoConfigPath, data, changed) {
		return nil, false, fmt.Errorf("repo config not trusted: %s", repoConfigPath)
	}
	newEntry := trustedRepoEntry{
		RepoRoot:   repoRootReal,
		ConfigPath: repoConfigReal,
		Hash:       hash,
	}
	list = upsertTrustedEntry(list, newEntry, index, ok)
	if err := saveTrustedRepoList(trustPath, list); err != nil {
		return nil, false, err
	}
	return data, true, nil
}

func loadTrustedRepoList(path string) (trustedRepoList, error) {
	list := trustedRepoList{Entries: []trustedRepoEntry{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return list, nil
		}
		return list, fmt.Errorf("read trusted repos: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return list, nil
	}
	if err := json.Unmarshal(data, &list); err == nil {
		return list, nil
	}
	var legacy struct {
		Entries map[string]string `json:"entries"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return list, fmt.Errorf("parse trusted repos: %w", err)
	}
	for key, hash := range legacy.Entries {
		parts := strings.SplitN(key, "\n", 2)
		if len(parts) != 2 {
			continue
		}
		list.Entries = append(list.Entries, trustedRepoEntry{
			RepoRoot:   parts[0],
			ConfigPath: parts[1],
			Hash:       hash,
		})
	}
	return list, nil
}

func saveTrustedRepoList(path string, list trustedRepoList) error {
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("encode trusted repos: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir trusted repos dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write trusted repos: %w", err)
	}
	return nil
}

func trustedRepoListPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "trusted_repos.json"), nil
}

func computeHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func isInteractiveStdin() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func promptTrust(path string, data []byte, changed bool) bool {
	if changed {
		fmt.Fprintf(os.Stderr, "Repo config changed: %s\n", path)
	} else {
		fmt.Fprintf(os.Stderr, "Untrusted repo config detected: %s\n", path)
	}
	fmt.Fprintln(os.Stderr, "----")
	fmt.Fprint(os.Stderr, string(data))
	if len(data) == 0 || data[len(data)-1] != '\n' {
		fmt.Fprintln(os.Stderr)
	}
	fmt.Fprintln(os.Stderr, "----")
	fmt.Fprint(os.Stderr, "Trust this config? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	return strings.EqualFold(line, "y") || strings.EqualFold(line, "yes")
}

func findTrustedEntry(list trustedRepoList, repoRoot, repoConfigPath string) (trustedRepoEntry, int, bool) {
	for i, entry := range list.Entries {
		if entry.RepoRoot == repoRoot && entry.ConfigPath == repoConfigPath {
			return entry, i, true
		}
	}
	return trustedRepoEntry{}, -1, false
}

func upsertTrustedEntry(list trustedRepoList, entry trustedRepoEntry, index int, ok bool) trustedRepoList {
	if ok && index >= 0 && index < len(list.Entries) {
		list.Entries[index] = entry
		return list
	}
	list.Entries = append(list.Entries, entry)
	return list
}
