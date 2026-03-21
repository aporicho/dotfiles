package module

import (
	"os"
	"path/filepath"
	"testing"
)

func writeModuleFile(t *testing.T, dir, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, "module.toml"), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseModule(t *testing.T) {
	dir := t.TempDir()
	writeModuleFile(t, dir, `
name = "git"
description = "Git configuration"
platforms = ["darwin", "linux"]
requires = ["base"]
exclude = ["*.bak"]

[[links]]
source = "gitconfig"
target = "~/.gitconfig"

[[links]]
source = "gitignore"
target = "~/.gitignore_global"
platforms = ["darwin"]

[deps.darwin]
brew = ["git", "git-lfs"]

[deps.linux]
apt = ["git"]

[hooks]
pre_install = "echo pre"
post_install = "echo post"
pre_remove = "echo pre-rm"
post_remove = "echo post-rm"
`)

	m, err := Parse(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Name != "git" {
		t.Errorf("Name = %q, want %q", m.Name, "git")
	}
	if m.Description != "Git configuration" {
		t.Errorf("Description = %q, want %q", m.Description, "Git configuration")
	}
	if len(m.Platforms) != 2 || m.Platforms[0] != "darwin" || m.Platforms[1] != "linux" {
		t.Errorf("Platforms = %v, want [darwin linux]", m.Platforms)
	}
	if len(m.Requires) != 1 || m.Requires[0] != "base" {
		t.Errorf("Requires = %v, want [base]", m.Requires)
	}
	if len(m.Exclude) != 1 || m.Exclude[0] != "*.bak" {
		t.Errorf("Exclude = %v, want [*.bak]", m.Exclude)
	}
	if len(m.Links) != 2 {
		t.Fatalf("Links count = %d, want 2", len(m.Links))
	}
	if m.Links[0].Source != "gitconfig" || m.Links[0].Target != "~/.gitconfig" {
		t.Errorf("Links[0] = %+v, want source=gitconfig target=~/.gitconfig", m.Links[0])
	}
	if len(m.Links[1].Platforms) != 1 || m.Links[1].Platforms[0] != "darwin" {
		t.Errorf("Links[1].Platforms = %v, want [darwin]", m.Links[1].Platforms)
	}
	if len(m.Deps.Darwin.Brew) != 2 || m.Deps.Darwin.Brew[0] != "git" {
		t.Errorf("Deps.Darwin.Brew = %v, want [git git-lfs]", m.Deps.Darwin.Brew)
	}
	if len(m.Deps.Linux.Apt) != 1 || m.Deps.Linux.Apt[0] != "git" {
		t.Errorf("Deps.Linux.Apt = %v, want [git]", m.Deps.Linux.Apt)
	}
	if m.Hooks.PreInstall != "echo pre" {
		t.Errorf("Hooks.PreInstall = %q, want %q", m.Hooks.PreInstall, "echo pre")
	}
	if m.Hooks.PostInstall != "echo post" {
		t.Errorf("Hooks.PostInstall = %q, want %q", m.Hooks.PostInstall, "echo post")
	}
	if m.Hooks.PreRemove != "echo pre-rm" {
		t.Errorf("Hooks.PreRemove = %q, want %q", m.Hooks.PreRemove, "echo pre-rm")
	}
	if m.Hooks.PostRemove != "echo post-rm" {
		t.Errorf("Hooks.PostRemove = %q, want %q", m.Hooks.PostRemove, "echo post-rm")
	}
	if m.Dir != dir {
		t.Errorf("Dir = %q, want %q", m.Dir, dir)
	}
}

func TestParseModule_Minimal(t *testing.T) {
	dir := t.TempDir()
	writeModuleFile(t, dir, `
name = "simple"
description = "A simple module"

[[links]]
source = "config"
target = "~/.config"
`)

	m, err := Parse(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Name != "simple" {
		t.Errorf("Name = %q, want %q", m.Name, "simple")
	}
	if m.Description != "A simple module" {
		t.Errorf("Description = %q, want %q", m.Description, "A simple module")
	}
	if len(m.Links) != 1 {
		t.Fatalf("Links count = %d, want 1", len(m.Links))
	}
	if m.Links[0].Source != "config" || m.Links[0].Target != "~/.config" {
		t.Errorf("Links[0] = %+v, want source=config target=~/.config", m.Links[0])
	}
	if len(m.Platforms) != 0 {
		t.Errorf("Platforms = %v, want empty", m.Platforms)
	}
	if len(m.Requires) != 0 {
		t.Errorf("Requires = %v, want empty", m.Requires)
	}
}

func TestParseModule_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := Parse(dir)
	if err == nil {
		t.Fatal("expected error for missing module.toml, got nil")
	}
}

func TestParseModule_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	writeModuleFile(t, dir, `
name = "broken
this is not valid toml [[[
`)

	_, err := Parse(dir)
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

func TestParseModule_MissingName(t *testing.T) {
	dir := t.TempDir()
	writeModuleFile(t, dir, `
description = "no name"

[[links]]
source = "a"
target = "b"
`)

	_, err := Parse(dir)
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestParseModule_MissingLinks(t *testing.T) {
	dir := t.TempDir()
	writeModuleFile(t, dir, `
name = "no-links"
description = "module without links"
`)

	_, err := Parse(dir)
	if err == nil {
		t.Fatal("expected error for missing links, got nil")
	}
}

func TestParseWithSecrets(t *testing.T) {
	dir := t.TempDir()
	tomlContent := `
name = "zsh"
description = "Zsh config"

[[links]]
source = ".zshrc"
target = "~/.zshrc"

[[secrets]]
source = "secrets.env"
encrypted = "secrets.env.age"
target = "~/.zsh/secrets.env"
`
	os.WriteFile(filepath.Join(dir, "module.toml"), []byte(tomlContent), 0o644)
	os.WriteFile(filepath.Join(dir, ".zshrc"), []byte("# zshrc"), 0o644)

	mod, err := Parse(dir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(mod.Secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(mod.Secrets))
	}
	s := mod.Secrets[0]
	if s.Source != "secrets.env" {
		t.Errorf("source = %q, want %q", s.Source, "secrets.env")
	}
	if s.Encrypted != "secrets.env.age" {
		t.Errorf("encrypted = %q, want %q", s.Encrypted, "secrets.env.age")
	}
	if s.Target != "~/.zsh/secrets.env" {
		t.Errorf("target = %q, want %q", s.Target, "~/.zsh/secrets.env")
	}
}
