package manifest

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// LinkRecord describes a single symlink created during module installation.
type LinkRecord struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Backup string `json:"backup,omitempty"` // path to backup file, if any
}

// ModuleRecord tracks when a module was installed and the links it created.
type ModuleRecord struct {
	InstalledAt string       `json:"installed_at"`
	Links       []LinkRecord `json:"links"`
}

// Manifest is the persistent state file for all installed modules.
type Manifest struct {
	Version      int                      `json:"version"`
	DotfilesPath string                   `json:"dotfiles_path"`
	Modules      map[string]*ModuleRecord `json:"modules"`
}

// New creates an empty manifest with Version=1.
func New(dotfilesPath string) *Manifest {
	return &Manifest{
		Version:      1,
		DotfilesPath: dotfilesPath,
		Modules:      make(map[string]*ModuleRecord),
	}
}

// AddModule records the installation of a module with a timestamp and its links.
func (m *Manifest) AddModule(name string, links []LinkRecord) {
	m.Modules[name] = &ModuleRecord{
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
		Links:       links,
	}
}

// RemoveModule deletes the record for the named module. It is a no-op if the
// module is not present.
func (m *Manifest) RemoveModule(name string) {
	delete(m.Modules, name)
}

// IsInstalled reports whether the named module has a record in the manifest.
func (m *Manifest) IsInstalled(name string) bool {
	_, ok := m.Modules[name]
	return ok
}

// Save writes the manifest as indented JSON to the given path, creating parent
// directories as needed.
func Save(m *Manifest, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Load reads a manifest from the given path. If the file does not exist, it
// returns a new empty manifest (not an error).
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return New(""), nil
		}
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Modules == nil {
		m.Modules = make(map[string]*ModuleRecord)
	}
	return &m, nil
}

// DefaultPath returns the default manifest file location:
// ~/.local/share/dot/manifest.json
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "dot", "manifest.json"), nil
}
