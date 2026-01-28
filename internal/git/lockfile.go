package git

import (
	"path/filepath"
	"strings"
)

var lockFileNames = map[string]bool{
	"uv.lock":           true,
	"poetry.lock":       true,
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"Gemfile.lock":      true,
	"Cargo.lock":        true,
	"go.sum":            true,
	"composer.lock":     true,
}

var lockFileSuffixes = []string{
	".lock",
	"-lock.json",
	"-lock.yaml",
}

func IsLockFile(filename string) bool {
	base := filepath.Base(filename)

	if lockFileNames[base] {
		return true
	}

	lower := strings.ToLower(base)
	for _, suffix := range lockFileSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}

	return false
}
