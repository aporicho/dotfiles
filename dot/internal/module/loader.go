package module

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// LoadAll reads all subdirectories in modulesDir, parses each one that
// contains a module.toml, and returns the resulting modules sorted by name.
// Subdirectories without a module.toml are silently skipped.
// Returns an error if modulesDir does not exist or cannot be read.
func LoadAll(modulesDir string) ([]*Module, error) {
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return nil, fmt.Errorf("reading modules directory: %w", err)
	}

	var modules []*Module
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(modulesDir, entry.Name())
		tomlPath := filepath.Join(dir, "module.toml")

		if _, err := os.Stat(tomlPath); os.IsNotExist(err) {
			continue
		}

		m, err := Parse(dir)
		if err != nil {
			return nil, fmt.Errorf("loading module %s: %w", entry.Name(), err)
		}

		modules = append(modules, m)
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})

	return modules, nil
}
