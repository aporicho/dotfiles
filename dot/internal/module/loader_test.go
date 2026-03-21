package module

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func writeModuleToml(t *testing.T, dir, name, target string) {
	t.Helper()
	os.MkdirAll(dir, 0755)
	content := fmt.Sprintf(`
name = "%s"
description = "test module"

[[links]]
source = "."
target = "%s"
`, name, target)
	os.WriteFile(filepath.Join(dir, "module.toml"), []byte(content), 0644)
}

func TestLoadAll(t *testing.T) {
	root := t.TempDir()

	writeModuleToml(t, filepath.Join(root, "zsh"), "zsh", "~/.zshrc")
	writeModuleToml(t, filepath.Join(root, "git"), "git", "~/.gitconfig")
	// dir without module.toml — should be skipped
	os.MkdirAll(filepath.Join(root, "nomodule"), 0755)

	modules, err := LoadAll(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(modules) != 2 {
		t.Fatalf("got %d modules, want 2", len(modules))
	}

	// Should be sorted by name: git before zsh
	if modules[0].Name != "git" {
		t.Errorf("modules[0].Name = %q, want %q", modules[0].Name, "git")
	}
	if modules[1].Name != "zsh" {
		t.Errorf("modules[1].Name = %q, want %q", modules[1].Name, "zsh")
	}
}

func TestLoadAll_EmptyDir(t *testing.T) {
	root := t.TempDir()

	modules, err := LoadAll(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(modules) != 0 {
		t.Fatalf("got %d modules, want 0", len(modules))
	}
}

func TestLoadAll_MissingDir(t *testing.T) {
	_, err := LoadAll("/tmp/nonexistent-dotfiles-dir-xyz")
	if err == nil {
		t.Fatal("expected error for missing directory, got nil")
	}
}
