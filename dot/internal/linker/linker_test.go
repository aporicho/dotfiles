package linker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateLink_SingleFile(t *testing.T) {
	tmp := t.TempDir()

	source := filepath.Join(tmp, "source.conf")
	target := filepath.Join(tmp, "target.conf")

	if err := os.WriteFile(source, []byte("hello"), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result, err := CreateLink(source, target)
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}

	if result.Source != source {
		t.Errorf("Source = %q, want %q", result.Source, source)
	}
	if result.Target != target {
		t.Errorf("Target = %q, want %q", result.Target, target)
	}
	if result.BackupPath != "" {
		t.Errorf("BackupPath = %q, want empty", result.BackupPath)
	}

	// Verify it's a symlink pointing to source
	got, err := os.Readlink(target)
	if err != nil {
		t.Fatalf("Readlink: %v", err)
	}
	if got != source {
		t.Errorf("symlink target = %q, want %q", got, source)
	}
}

func TestCreateLink_ParentDirCreated(t *testing.T) {
	tmp := t.TempDir()

	source := filepath.Join(tmp, "source.conf")
	target := filepath.Join(tmp, "a", "b", "c", "target.conf")

	if err := os.WriteFile(source, []byte("deep"), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result, err := CreateLink(source, target)
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}

	if result.Target != target {
		t.Errorf("Target = %q, want %q", result.Target, target)
	}

	// Verify parent directories were created
	info, err := os.Lstat(filepath.Dir(target))
	if err != nil {
		t.Fatalf("stat parent dir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("parent is not a directory")
	}

	// Verify symlink works
	got, err := os.Readlink(target)
	if err != nil {
		t.Fatalf("Readlink: %v", err)
	}
	if got != source {
		t.Errorf("symlink target = %q, want %q", got, source)
	}
}

func TestCreateLink_BackupExisting(t *testing.T) {
	tmp := t.TempDir()

	source := filepath.Join(tmp, "source.conf")
	target := filepath.Join(tmp, "target.conf")

	originalContent := "original content"

	if err := os.WriteFile(source, []byte("new"), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(target, []byte(originalContent), 0644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	result, err := CreateLink(source, target)
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}

	// BackupPath should be set
	if result.BackupPath == "" {
		t.Fatal("BackupPath is empty, expected backup to be made")
	}
	if !strings.HasPrefix(result.BackupPath, target+".backup.") {
		t.Errorf("BackupPath = %q, want prefix %q", result.BackupPath, target+".backup.")
	}

	// Verify backup content matches original
	data, err := os.ReadFile(result.BackupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(data) != originalContent {
		t.Errorf("backup content = %q, want %q", string(data), originalContent)
	}

	// Verify target is now a symlink to source
	got, err := os.Readlink(target)
	if err != nil {
		t.Fatalf("Readlink: %v", err)
	}
	if got != source {
		t.Errorf("symlink target = %q, want %q", got, source)
	}
}

func TestCreateLink_ReplaceExistingSymlink(t *testing.T) {
	tmp := t.TempDir()

	oldSource := filepath.Join(tmp, "old.conf")
	newSource := filepath.Join(tmp, "new.conf")
	target := filepath.Join(tmp, "target.conf")

	if err := os.WriteFile(oldSource, []byte("old"), 0644); err != nil {
		t.Fatalf("write old source: %v", err)
	}
	if err := os.WriteFile(newSource, []byte("new"), 0644); err != nil {
		t.Fatalf("write new source: %v", err)
	}

	// Create existing symlink to old source
	if err := os.Symlink(oldSource, target); err != nil {
		t.Fatalf("create initial symlink: %v", err)
	}

	result, err := CreateLink(newSource, target)
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}

	// No backup for symlink replacement
	if result.BackupPath != "" {
		t.Errorf("BackupPath = %q, want empty (no backup for symlink)", result.BackupPath)
	}

	// Verify target now points to new source
	got, err := os.Readlink(target)
	if err != nil {
		t.Fatalf("Readlink: %v", err)
	}
	if got != newSource {
		t.Errorf("symlink target = %q, want %q", got, newSource)
	}
}

func TestExpandDirEntries(t *testing.T) {
	tmp := t.TempDir()

	// Create test entries
	files := []string{"module.toml", "README.md", ".zshrc", "config", "aliases"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tmp, f), []byte("x"), 0644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}
	// Also create a subdirectory
	if err := os.Mkdir(filepath.Join(tmp, "snippets"), 0755); err != nil {
		t.Fatalf("mkdir snippets: %v", err)
	}

	entries, err := ExpandDirEntries(tmp, []string{"module.toml", "README*"})
	if err != nil {
		t.Fatalf("ExpandDirEntries: %v", err)
	}

	// Should include: .zshrc, aliases, config, snippets
	// Should exclude: module.toml, README.md
	expected := map[string]bool{
		".zshrc":   false,
		"config":   false,
		"aliases":  false,
		"snippets": false,
	}

	for _, e := range entries {
		name := filepath.Base(e)
		if name == "module.toml" || strings.HasPrefix(name, "README") {
			t.Errorf("entry %q should have been excluded", name)
		}
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected entry %q not found in results", name)
		}
	}
}

func TestRemoveLinks(t *testing.T) {
	tmp := t.TempDir()

	source := filepath.Join(tmp, "source.conf")
	target := filepath.Join(tmp, "target.conf")

	if err := os.WriteFile(source, []byte("data"), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.Symlink(source, target); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	// Remove should succeed
	if err := RemoveLink(target); err != nil {
		t.Fatalf("RemoveLink: %v", err)
	}

	// Verify target is gone
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Errorf("target should not exist after removal, got err: %v", err)
	}

	// Remove again — already gone, should return nil
	if err := RemoveLink(target); err != nil {
		t.Errorf("RemoveLink on missing target: %v", err)
	}

	// Remove a regular file — should return error
	regular := filepath.Join(tmp, "regular.txt")
	if err := os.WriteFile(regular, []byte("x"), 0644); err != nil {
		t.Fatalf("write regular: %v", err)
	}
	if err := RemoveLink(regular); err == nil {
		t.Error("RemoveLink on regular file should return error")
	}
}
