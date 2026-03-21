package module

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Link represents a symlink mapping from source to target.
type Link struct {
	Source    string   `toml:"source"`
	Target   string   `toml:"target"`
	Platforms []string `toml:"platforms,omitempty"`
}

// PlatformDeps holds package manager dependencies for a specific platform.
type PlatformDeps struct {
	Brew []string `toml:"brew,omitempty"`
	Apt  []string `toml:"apt,omitempty"`
}

// Deps holds platform-specific dependency declarations.
type Deps struct {
	Darwin PlatformDeps `toml:"darwin"`
	Linux  PlatformDeps `toml:"linux"`
}

// Hooks holds lifecycle hook commands.
type Hooks struct {
	PreInstall  string `toml:"pre_install"`
	PostInstall string `toml:"post_install"`
	PreRemove   string `toml:"pre_remove"`
	PostRemove  string `toml:"post_remove"`
}

// Module represents a parsed module.toml configuration.
type Module struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Platforms   []string `toml:"platforms,omitempty"`
	Requires    []string `toml:"requires,omitempty"`
	Exclude     []string `toml:"exclude,omitempty"`
	Links       []Link   `toml:"links"`
	Deps        Deps     `toml:"deps"`
	Hooks       Hooks    `toml:"hooks"`
	Dir         string   `toml:"-"`
}

// Parse reads and validates a module.toml file from the given directory.
// It returns an error if the file is missing, malformed, or fails validation.
func Parse(dir string) (*Module, error) {
	path := filepath.Join(dir, "module.toml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading module.toml: %w", err)
	}

	var m Module
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing module.toml: %w", err)
	}

	if m.Name == "" {
		return nil, fmt.Errorf("module.toml: name is required")
	}
	if len(m.Links) == 0 {
		return nil, fmt.Errorf("module.toml: at least one [[links]] is required")
	}

	m.Dir = dir
	return &m, nil
}
