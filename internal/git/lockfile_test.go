package git

import "testing"

func TestIsLockFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		// Lock files that should be detected
		{"uv.lock", "uv.lock", true},
		{"poetry.lock", "poetry.lock", true},
		{"package-lock.json", "package-lock.json", true},
		{"yarn.lock", "yarn.lock", true},
		{"pnpm-lock.yaml", "pnpm-lock.yaml", true},
		{"Gemfile.lock", "Gemfile.lock", true},
		{"Cargo.lock", "Cargo.lock", true},
		{"go.sum", "go.sum", true},
		{"composer.lock", "composer.lock", true},

		// Lock files in subdirectories
		{"nested uv.lock", "packages/uv.lock", true},
		{"nested package-lock.json", "frontend/package-lock.json", true},

		// Non-lock files
		{"regular go file", "main.go", false},
		{"regular json", "config.json", false},
		{"go.mod", "go.mod", false},
		{"package.json", "package.json", false},
		{"Gemfile", "Gemfile", false},
		{"Cargo.toml", "Cargo.toml", false},
		{"pyproject.toml", "pyproject.toml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLockFile(tt.filename)
			if got != tt.want {
				t.Errorf("IsLockFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
