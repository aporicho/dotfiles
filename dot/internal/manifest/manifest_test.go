package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManifest(t *testing.T) {
	m := New("/home/user/dotfiles")

	if m.Version != 1 {
		t.Errorf("expected Version=1, got %d", m.Version)
	}
	if m.DotfilesPath != "/home/user/dotfiles" {
		t.Errorf("expected DotfilesPath=/home/user/dotfiles, got %s", m.DotfilesPath)
	}
	if m.Modules == nil {
		t.Fatal("expected Modules to be initialized, got nil")
	}
	if len(m.Modules) != 0 {
		t.Errorf("expected empty Modules map, got %d entries", len(m.Modules))
	}
}

func TestManifest_AddAndRemoveModule(t *testing.T) {
	m := New("/home/user/dotfiles")

	links := []LinkRecord{
		{Source: "/home/user/dotfiles/zsh/.zshrc", Target: "/home/user/.zshrc"},
		{Source: "/home/user/dotfiles/zsh/.zshenv", Target: "/home/user/.zshenv", Backup: "/home/user/.zshenv.bak"},
	}

	m.AddModule("zsh", links)

	rec, ok := m.Modules["zsh"]
	if !ok {
		t.Fatal("expected module 'zsh' to exist after AddModule")
	}
	if rec.InstalledAt == "" {
		t.Error("expected InstalledAt to be set")
	}
	if len(rec.Links) != 2 {
		t.Errorf("expected 2 links, got %d", len(rec.Links))
	}
	if rec.Links[0].Source != "/home/user/dotfiles/zsh/.zshrc" {
		t.Errorf("unexpected source: %s", rec.Links[0].Source)
	}
	if rec.Links[1].Backup != "/home/user/.zshenv.bak" {
		t.Errorf("unexpected backup: %s", rec.Links[1].Backup)
	}

	m.RemoveModule("zsh")

	if _, ok := m.Modules["zsh"]; ok {
		t.Error("expected module 'zsh' to be removed")
	}

	// RemoveModule on non-existent module should not panic
	m.RemoveModule("nonexistent")
}

func TestManifest_IsInstalled(t *testing.T) {
	m := New("/home/user/dotfiles")

	if m.IsInstalled("git") {
		t.Error("expected IsInstalled to return false for unknown module")
	}

	m.AddModule("git", []LinkRecord{
		{Source: "/home/user/dotfiles/git/.gitconfig", Target: "/home/user/.gitconfig"},
	})

	if !m.IsInstalled("git") {
		t.Error("expected IsInstalled to return true after AddModule")
	}
}

func TestManifest_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "manifest.json")

	original := New("/home/user/dotfiles")
	original.AddModule("vim", []LinkRecord{
		{Source: "/home/user/dotfiles/vim/.vimrc", Target: "/home/user/.vimrc"},
		{Source: "/home/user/dotfiles/vim/.vim", Target: "/home/user/.vim", Backup: "/home/user/.vim.bak"},
	})
	original.AddModule("tmux", []LinkRecord{
		{Source: "/home/user/dotfiles/tmux/.tmux.conf", Target: "/home/user/.tmux.conf"},
	})

	if err := Save(original, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file was created (including parent directory)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected manifest file to exist: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != original.Version {
		t.Errorf("Version mismatch: got %d, want %d", loaded.Version, original.Version)
	}
	if loaded.DotfilesPath != original.DotfilesPath {
		t.Errorf("DotfilesPath mismatch: got %s, want %s", loaded.DotfilesPath, original.DotfilesPath)
	}
	if len(loaded.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(loaded.Modules))
	}

	vim := loaded.Modules["vim"]
	if vim == nil {
		t.Fatal("expected vim module to exist")
	}
	if len(vim.Links) != 2 {
		t.Errorf("expected 2 vim links, got %d", len(vim.Links))
	}
	if vim.Links[1].Backup != "/home/user/.vim.bak" {
		t.Errorf("expected backup path preserved, got %s", vim.Links[1].Backup)
	}
	if vim.InstalledAt != original.Modules["vim"].InstalledAt {
		t.Errorf("InstalledAt mismatch for vim")
	}

	tmux := loaded.Modules["tmux"]
	if tmux == nil {
		t.Fatal("expected tmux module to exist")
	}
	if len(tmux.Links) != 1 {
		t.Errorf("expected 1 tmux link, got %d", len(tmux.Links))
	}
}

func TestManifest_LoadMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent", "manifest.json")

	m, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil manifest for missing file")
	}
	if m.Version != 1 {
		t.Errorf("expected Version=1 for missing file, got %d", m.Version)
	}
	if m.Modules == nil {
		t.Fatal("expected Modules to be initialized for missing file")
	}
	if len(m.Modules) != 0 {
		t.Errorf("expected empty Modules for missing file, got %d", len(m.Modules))
	}
}
