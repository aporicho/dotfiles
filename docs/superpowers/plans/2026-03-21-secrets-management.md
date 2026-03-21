# Secrets Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add age-encrypted secrets management to dot CLI, so sensitive env vars can be safely committed to git and synced across machines.

**Architecture:** Extend `module.toml` with `[[secrets]]` sections. New `internal/secrets/` package handles age encryption/decryption and system keychain access. `pull`, `push`, `doctor` commands gain secrets awareness. TUI passphrase input uses bubbletea.

**Tech Stack:** Go, `filippo.io/age` (scrypt passphrase mode), macOS Keychain (`security` CLI), Linux secret-service (`secret-tool` CLI), bubbletea TUI.

**Spec:** `docs/superpowers/specs/2026-03-21-secrets-management-design.md`

---

### Task 1: Extend Module struct with Secrets

**Files:**
- Modify: `dot/internal/module/module.go` (add Secret struct, update Module, relax validation)
- Modify: `dot/internal/module/module_test.go` (add test for secrets parsing)

- [ ] **Step 1: Write failing test for secrets parsing**

In `dot/internal/module/module_test.go`, add:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd dot && go test ./internal/module/ -run TestParseWithSecrets -v`
Expected: FAIL — `mod.Secrets` field doesn't exist.

- [ ] **Step 3: Add Secret struct and update Module**

In `dot/internal/module/module.go`, add the `Secret` struct after `Link`:

```go
type Secret struct {
	Source    string `toml:"source"`
	Encrypted string `toml:"encrypted"`
	Target    string `toml:"target"`
}
```

Add `Secrets` field to `Module` struct:

```go
type Module struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Platforms   []string `toml:"platforms,omitempty"`
	Requires    []string `toml:"requires,omitempty"`
	Exclude     []string `toml:"exclude,omitempty"`
	Links       []Link   `toml:"links"`
	Secrets     []Secret `toml:"secrets"`
	Deps        Deps     `toml:"deps"`
	Hooks       Hooks    `toml:"hooks"`
	Dir         string   `toml:"-"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd dot && go test ./internal/module/ -run TestParseWithSecrets -v`
Expected: PASS

- [ ] **Step 5: Run all module tests**

Run: `cd dot && go test ./internal/module/ -v`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add dot/internal/module/module.go dot/internal/module/module_test.go
git commit -m "feat(module): add Secret struct and [[secrets]] parsing"
```

---

### Task 2: Create secrets crypto package (age encrypt/decrypt)

**Files:**
- Create: `dot/internal/secrets/crypto.go`
- Create: `dot/internal/secrets/crypto_test.go`
- Modify: `dot/go.mod` (add `filippo.io/age`)

- [ ] **Step 1: Add age dependency**

Run: `cd dot && go get filippo.io/age`

- [ ] **Step 2: Write failing tests**

Create `dot/internal/secrets/crypto_test.go`:

```go
package secrets

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plaintext := []byte("export TOKEN=\"secret123\"\n")
	passphrase := "test-passphrase"

	encrypted, err := Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if bytes.Equal(encrypted, plaintext) {
		t.Fatal("encrypted should differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted, passphrase)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("roundtrip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	plaintext := []byte("export TOKEN=\"secret123\"\n")
	encrypted, err := Encrypt(plaintext, "correct")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = Decrypt(encrypted, "wrong")
	if err == nil {
		t.Fatal("expected error with wrong passphrase")
	}
}

func TestEncryptNonDeterministic(t *testing.T) {
	plaintext := []byte("same content")
	passphrase := "same-pass"

	enc1, _ := Encrypt(plaintext, passphrase)
	enc2, _ := Encrypt(plaintext, passphrase)

	if bytes.Equal(enc1, enc2) {
		t.Fatal("age encryption should be non-deterministic")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd dot && go test ./internal/secrets/ -v`
Expected: FAIL — package doesn't exist.

- [ ] **Step 4: Implement crypto.go**

Create `dot/internal/secrets/crypto.go`:

```go
package secrets

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
)

// Encrypt encrypts plaintext using age scrypt (passphrase-based).
func Encrypt(plaintext []byte, passphrase string) ([]byte, error) {
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipient)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(plaintext); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decrypt decrypts age-encrypted ciphertext using passphrase.
func Decrypt(ciphertext []byte, passphrase string) ([]byte, error) {
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, err
	}

	r, err := age.Decrypt(bytes.NewReader(ciphertext), identity)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

// HasChanged checks if plaintext differs from encrypted content.
// Returns true if they differ or if decryption fails.
func HasChanged(plaintextPath, encryptedPath, passphrase string) (bool, error) {
	plaintext, err := os.ReadFile(plaintextPath)
	if err != nil {
		return false, err
	}
	ciphertext, err := os.ReadFile(encryptedPath)
	if err != nil {
		// No encrypted file yet → changed
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	decrypted, err := Decrypt(ciphertext, passphrase)
	if err != nil {
		return false, fmt.Errorf("decrypting %s: %w", encryptedPath, err)
	}
	return !bytes.Equal(plaintext, decrypted), nil
}
```

- [ ] **Step 5: Write test for HasChanged**

Append to `crypto_test.go`:

```go
func TestHasChanged(t *testing.T) {
	dir := t.TempDir()
	plainPath := filepath.Join(dir, "secrets.env")
	encPath := filepath.Join(dir, "secrets.env.age")
	passphrase := "test"

	// Write plaintext
	content := []byte("export A=1\n")
	os.WriteFile(plainPath, content, 0o600)

	// No encrypted file → changed
	changed, err := HasChanged(plainPath, encPath, passphrase)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("should be changed when no .age exists")
	}

	// Encrypt and write
	enc, _ := Encrypt(content, passphrase)
	os.WriteFile(encPath, enc, 0o644)

	// Same content → not changed
	changed, err = HasChanged(plainPath, encPath, passphrase)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("should not be changed when content matches")
	}

	// Modify plaintext → changed
	os.WriteFile(plainPath, []byte("export A=2\n"), 0o600)
	changed, err = HasChanged(plainPath, encPath, passphrase)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("should be changed after modification")
	}
}
```

Add `"os"` and `"path/filepath"` to test imports.

- [ ] **Step 6: Run all tests**

Run: `cd dot && go test ./internal/secrets/ -v`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add dot/internal/secrets/ dot/go.mod dot/go.sum
git commit -m "feat(secrets): add age encrypt/decrypt package"
```

---

### Task 3: Create keychain package

**Files:**
- Create: `dot/internal/secrets/keychain.go`
- Create: `dot/internal/secrets/keychain_test.go`

- [ ] **Step 1: Write failing test**

Create `dot/internal/secrets/keychain_test.go`:

```go
package secrets

import (
	"testing"
)

func TestKeychainCommands(t *testing.T) {
	// Unit test: verify command construction, not actual keychain access
	save := keychainSaveCmd("my-passphrase")
	if save == nil {
		t.Fatal("save command should not be nil")
	}

	load := keychainLoadCmd()
	if load == nil {
		t.Fatal("load command should not be nil")
	}

	del := keychainDeleteCmd()
	if del == nil {
		t.Fatal("delete command should not be nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd dot && go test ./internal/secrets/ -run TestKeychain -v`
Expected: FAIL — functions don't exist.

- [ ] **Step 3: Implement keychain.go**

Create `dot/internal/secrets/keychain.go`:

```go
package secrets

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	keychainService = "dot-secrets"
)

// keychainSaveCmd returns the command to save passphrase to system keychain.
func keychainSaveCmd(passphrase string) *exec.Cmd {
	user := currentUser()
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("security", "add-generic-password",
			"-s", keychainService, "-a", user, "-w", passphrase, "-U")
	case "linux":
		cmd := exec.Command("secret-tool", "store",
			"--label", keychainService, "app", "dot")
		cmd.Stdin = strings.NewReader(passphrase)
		return cmd
	}
	return nil
}

// keychainLoadCmd returns the command to load passphrase from system keychain.
func keychainLoadCmd() *exec.Cmd {
	user := currentUser()
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("security", "find-generic-password",
			"-s", keychainService, "-a", user, "-w")
	case "linux":
		return exec.Command("secret-tool", "lookup", "app", "dot")
	}
	return nil
}

// keychainDeleteCmd returns the command to delete passphrase from system keychain.
func keychainDeleteCmd() *exec.Cmd {
	user := currentUser()
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("security", "delete-generic-password",
			"-s", keychainService, "-a", user)
	case "linux":
		return exec.Command("secret-tool", "clear", "app", "dot")
	}
	return nil
}

// LoadPassphrase retrieves the passphrase from system keychain.
// Returns empty string and nil error if keychain is unavailable.
func LoadPassphrase() (string, error) {
	cmd := keychainLoadCmd()
	if cmd == nil {
		return "", nil
	}
	out, err := cmd.Output()
	if err != nil {
		return "", nil // keychain miss or unavailable
	}
	return strings.TrimSpace(string(out)), nil
}

// SavePassphrase stores the passphrase in system keychain.
func SavePassphrase(passphrase string) error {
	cmd := keychainSaveCmd(passphrase)
	if cmd == nil {
		return fmt.Errorf("keychain not supported on %s", runtime.GOOS)
	}
	return cmd.Run()
}

// DeletePassphrase removes the passphrase from system keychain.
func DeletePassphrase() error {
	cmd := keychainDeleteCmd()
	if cmd == nil {
		return nil
	}
	return cmd.Run()
}

// KeychainAvailable returns true if a system keychain is available.
func KeychainAvailable() bool {
	switch runtime.GOOS {
	case "darwin":
		_, err := exec.LookPath("security")
		return err == nil
	case "linux":
		_, err := exec.LookPath("secret-tool")
		return err == nil
	}
	return false
}

func currentUser() string {
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	return "default"
}
```

- [ ] **Step 4: Run tests**

Run: `cd dot && go test ./internal/secrets/ -run TestKeychain -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add dot/internal/secrets/keychain.go dot/internal/secrets/keychain_test.go
git commit -m "feat(secrets): add keychain integration for passphrase storage"
```

---

### Task 4: Create TUI passphrase input

**Files:**
- Create: `dot/internal/tui/passphrase.go`
- Create: `dot/internal/tui/passphrase_test.go`

- [ ] **Step 1: Write failing test**

Create `dot/internal/tui/passphrase_test.go`:

```go
package tui

import "testing"

func TestPassphraseModelMasking(t *testing.T) {
	m := NewPassphraseModel("Enter passphrase:", false)
	// Simulate typing
	m.value = "hello"
	view := m.View()
	if contains(view, "hello") {
		t.Error("passphrase should be masked in view")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && stringContains(s, substr)
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd dot && go test ./internal/tui/ -run TestPassphraseModel -v`
Expected: FAIL — `NewPassphraseModel` doesn't exist.

- [ ] **Step 3: Implement passphrase.go**

Create `dot/internal/tui/passphrase.go`:

```go
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// PassphraseModel is a bubbletea model for passphrase input.
type PassphraseModel struct {
	prompt  string
	value   string
	confirm bool // if true, asks for confirmation
	phase   int  // 0=first input, 1=confirm input
	first   string
	err     string
	done    bool
	aborted bool
}

// NewPassphraseModel creates a passphrase input model.
// If confirm is true, the user must enter the passphrase twice.
func NewPassphraseModel(prompt string, confirm bool) PassphraseModel {
	return PassphraseModel{prompt: prompt, confirm: confirm}
}

func (m PassphraseModel) Init() tea.Cmd { return nil }

// Update handles key messages using bubbletea v1.x API (msg.String() pattern).
func (m PassphraseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			m.done = true
			return m, tea.Quit
		case "enter":
			if len(m.value) == 0 {
				m.err = "passphrase 不能为空"
				return m, nil
			}
			if m.confirm && m.phase == 0 {
				m.first = m.value
				m.value = ""
				m.phase = 1
				m.err = ""
				return m, nil
			}
			if m.confirm && m.phase == 1 && m.value != m.first {
				m.err = "两次输入不一致，请重新输入"
				m.value = ""
				m.phase = 0
				m.first = ""
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		case "backspace":
			if len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
			}
			return m, nil
		default:
			// Single character input (runes)
			if len(msg.String()) == 1 {
				m.value += msg.String()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m PassphraseModel) View() string {
	var b strings.Builder
	if m.confirm && m.phase == 1 {
		b.WriteString("  确认 Passphrase: ")
	} else {
		b.WriteString(fmt.Sprintf("  %s ", m.prompt))
	}
	b.WriteString(strings.Repeat("•", len(m.value)))
	if m.err != "" {
		b.WriteString(fmt.Sprintf("\n  ✗ %s", m.err))
	}
	b.WriteString("\n")
	return b.String()
}

// Value returns the entered passphrase.
func (m PassphraseModel) Value() string { return m.value }

// Aborted returns true if the user cancelled.
func (m PassphraseModel) Aborted() bool { return m.aborted }

// RunPassphraseInput runs the TUI passphrase input and returns the passphrase.
func RunPassphraseInput(prompt string, confirm bool) (string, error) {
	m := NewPassphraseModel(prompt, confirm)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	final := result.(PassphraseModel)
	if final.Aborted() {
		return "", fmt.Errorf("已取消")
	}
	return final.Value(), nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd dot && go test ./internal/tui/ -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add dot/internal/tui/passphrase.go dot/internal/tui/passphrase_test.go
git commit -m "feat(tui): add passphrase input model"
```

---

### Task 5: Create high-level secrets manager

**Files:**
- Create: `dot/internal/secrets/manager.go`
- Create: `dot/internal/secrets/manager_test.go`

This ties together crypto + keychain + TUI into a single interface used by commands.

- [ ] **Step 1: Write failing test**

Create `dot/internal/secrets/manager_test.go`:

```go
package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptFile(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "secrets.env")
	enc := filepath.Join(dir, "secrets.env.age")

	os.WriteFile(plain, []byte("export A=1\n"), 0o600)

	err := EncryptFile(plain, enc, "test-pass")
	if err != nil {
		t.Fatalf("EncryptFile: %v", err)
	}

	// Encrypted file should exist
	if _, err := os.Stat(enc); err != nil {
		t.Fatalf("encrypted file not created: %v", err)
	}
}

func TestDecryptFile(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "secrets.env")
	enc := filepath.Join(dir, "secrets.env.age")

	content := []byte("export A=1\n")
	os.WriteFile(plain, content, 0o600)
	EncryptFile(plain, enc, "test-pass")

	// Remove plaintext
	os.Remove(plain)

	// Decrypt
	err := DecryptFile(enc, plain, "test-pass")
	if err != nil {
		t.Fatalf("DecryptFile: %v", err)
	}

	got, _ := os.ReadFile(plain)
	if string(got) != string(content) {
		t.Errorf("got %q, want %q", got, content)
	}

	// Check permissions
	info, _ := os.Stat(plain)
	if info.Mode().Perm() != 0o600 {
		t.Errorf("permissions = %o, want 0600", info.Mode().Perm())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd dot && go test ./internal/secrets/ -run "TestEncryptFile|TestDecryptFile" -v`
Expected: FAIL — functions don't exist.

- [ ] **Step 3: Implement manager.go**

Create `dot/internal/secrets/manager.go`:

```go
package secrets

import (
	"fmt"
	"os"
)

// EncryptFile encrypts a plaintext file and writes the result to encPath.
func EncryptFile(plainPath, encPath, passphrase string) error {
	plaintext, err := os.ReadFile(plainPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", plainPath, err)
	}
	ciphertext, err := Encrypt(plaintext, passphrase)
	if err != nil {
		return fmt.Errorf("encrypting: %w", err)
	}
	return os.WriteFile(encPath, ciphertext, 0o644)
}

// DecryptFile decrypts an age file and writes plaintext with 0600 permissions.
func DecryptFile(encPath, plainPath, passphrase string) error {
	ciphertext, err := os.ReadFile(encPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", encPath, err)
	}
	plaintext, err := Decrypt(ciphertext, passphrase)
	if err != nil {
		return fmt.Errorf("decrypting: %w", err)
	}
	return os.WriteFile(plainPath, plaintext, 0o600)
}

// GetPassphrase tries keychain first, falls back to TUI prompt.
// askFunc is the TUI callback: func(prompt string, confirm bool) (string, error)
// If saveFunc is non-nil and keychain is available, offers to save after TUI input.
func GetPassphrase(askFunc func(string, bool) (string, error)) (string, error) {
	// Try keychain first
	if KeychainAvailable() {
		if p, _ := LoadPassphrase(); p != "" {
			return p, nil
		}
	}

	// Fall back to interactive input
	p, err := askFunc("Passphrase:", false)
	if err != nil {
		return "", err
	}
	return p, nil
}

// GetNewPassphrase prompts for a new passphrase with confirmation.
func GetNewPassphrase(askFunc func(string, bool) (string, error)) (string, error) {
	return askFunc("Passphrase:", true)
}

// OfferSaveToKeychain asks if the user wants to save to keychain and does it.
func OfferSaveToKeychain(passphrase string) {
	if !KeychainAvailable() {
		return
	}
	// Save silently — the TUI prompt for "save to keychain?" is handled by the caller
	SavePassphrase(passphrase)
}
```

- [ ] **Step 4: Run all secrets tests**

Run: `cd dot && go test ./internal/secrets/ -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add dot/internal/secrets/manager.go dot/internal/secrets/manager_test.go
git commit -m "feat(secrets): add file encrypt/decrypt and passphrase manager"
```

---

### Task 6: Integrate secrets into `dot pull`

**Files:**
- Modify: `dot/cmd/pull.go` (add secret handling after symlink install)

- [ ] **Step 1: Add secrets handling to installModule**

In `dot/cmd/pull.go`, insert secrets handling **between the symlink creation loop (ending at line ~211) and the post_install hook (line ~213)**. The secrets symlinks are appended to `createdLinks` (same as regular links), so they get included in the manifest records built at line ~220.

```go
	// Handle secrets (insert after line 211, before "Step 4: post_install hook")
	for _, sec := range mod.Secrets {
		encPath := filepath.Join(modDir, sec.Encrypted)
		plainPath := filepath.Join(modDir, sec.Source)
		target := expandHome(sec.Target)

		// Skip if no encrypted file exists
		if _, err := os.Stat(encPath); os.IsNotExist(err) {
			continue
		}

		// Decrypt if plaintext doesn't exist
		if _, err := os.Stat(plainPath); os.IsNotExist(err) {
			passphrase, err := getPassphrase()
			if err != nil {
				return fmt.Errorf("getting passphrase: %w", err)
			}
			fmt.Printf("  🔐 解密 %s\n", sec.Encrypted)
			if err := secrets.DecryptFile(encPath, plainPath, passphrase); err != nil {
				return fmt.Errorf("decrypting %s: %w", sec.Encrypted, err)
			}
			// Offer to save to keychain on first successful decrypt
			if secrets.KeychainAvailable() {
				secrets.OfferSaveToKeychain(passphrase)
				fmt.Println("  ✓ 已保存到系统 Keychain")
			}
		}

		// Create symlink for secrets — append to createdLinks (same as regular links)
		lr, err := linker.CreateLink(plainPath, target)
		if err != nil {
			return fmt.Errorf("linking secret %s: %w", sec.Source, err)
		}
		createdLinks = append(createdLinks, lr)
	}
```

- [ ] **Step 2: Add getPassphrase helper and import**

At the top of `pull.go`, add import for `secrets` package. Add a file-level helper with **retry logic (max 3 attempts)**:

```go
import (
	// ... existing imports
	"github.com/aporicho/dotfiles/dot/internal/secrets"
	"github.com/aporicho/dotfiles/dot/internal/tui"
)

// Cached passphrase for the current session
var cachedPassphrase string

func getPassphrase() (string, error) {
	if cachedPassphrase != "" {
		return cachedPassphrase, nil
	}

	// Try keychain first
	if secrets.KeychainAvailable() {
		if p, _ := secrets.LoadPassphrase(); p != "" {
			cachedPassphrase = p
			return p, nil
		}
	}

	// Interactive input with retry (max 3 attempts)
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		p, err := tui.RunPassphraseInput("Passphrase:", false)
		if err != nil {
			return "", err
		}
		cachedPassphrase = p
		return p, nil
	}
	return "", fmt.Errorf("超过最大重试次数")
}
```

Note: `expandHome` already exists in `pull.go` at line ~233. No need to add it again.

- [ ] **Step 3: Test manually**

Run: `cd dot && go build -o dot . && ./dot pull --help`
Expected: Builds without error.

- [ ] **Step 4: Commit**

```bash
git add dot/cmd/pull.go
git commit -m "feat(pull): decrypt secrets on module install"
```

---

### Task 7: Integrate secrets into `dot push`

**Files:**
- Modify: `dot/cmd/push.go` (encrypt secrets before commit)

- [ ] **Step 1: Add secrets encryption before git add**

In `dot/cmd/push.go`, before the `gitpkg.AddAndCommit(...)` call, add:

```go
// Encrypt secrets if changed
allModules, _ := module.LoadAll(filepath.Join(dfPath, "modules"))
for _, mod := range allModules {
	if len(mod.Secrets) == 0 {
		continue
	}
	modDir := mod.Dir
	for _, sec := range mod.Secrets {
		plainPath := filepath.Join(modDir, sec.Source)
		encPath := filepath.Join(modDir, sec.Encrypted)

		// Skip if no plaintext exists
		if _, err := os.Stat(plainPath); os.IsNotExist(err) {
			continue
		}

		passphrase, err := getPushPassphrase(encPath)
		if err != nil {
			return fmt.Errorf("getting passphrase: %w", err)
		}

		changed, err := secrets.HasChanged(plainPath, encPath, passphrase)
		if err != nil {
			return fmt.Errorf("checking secrets: %w", err)
		}

		if changed {
			fmt.Printf("  🔐 加密 %s → %s\n", sec.Source, sec.Encrypted)
			if err := secrets.EncryptFile(plainPath, encPath, passphrase); err != nil {
				return fmt.Errorf("encrypting %s: %w", sec.Source, err)
			}
		}
	}
}
```

- [ ] **Step 2: Add getPushPassphrase helper**

```go
var cachedPushPassphrase string

func getPushPassphrase(encPath string) (string, error) {
	if cachedPushPassphrase != "" {
		return cachedPushPassphrase, nil
	}

	// If no .age file exists, this is first-time encryption → ask with confirmation
	_, err := os.Stat(encPath)
	isNew := os.IsNotExist(err)

	var p string
	if isNew {
		fmt.Println("  🔐 首次加密，请设置 passphrase")
		p, err = secrets.GetNewPassphrase(tui.RunPassphraseInput)
	} else {
		p, err = secrets.GetPassphrase(tui.RunPassphraseInput)
	}
	if err != nil {
		return "", err
	}

	cachedPushPassphrase = p

	// Offer to save to keychain
	if isNew && secrets.KeychainAvailable() {
		secrets.OfferSaveToKeychain(p)
		fmt.Println("  ✓ 已保存到系统 Keychain")
	}

	return p, nil
}
```

Add imports for `secrets`, `tui`, `module` packages.

- [ ] **Step 3: Update git add paths to include .gitignore**

In `push.go`, find the `AddAndCommit` call. Change `[]string{"modules/"}` to `[]string{"modules/", ".gitignore"}` to ensure gitignore changes are also committed.

- [ ] **Step 4: Test manually**

Run: `cd dot && go build -o dot .`
Expected: Builds without error.

- [ ] **Step 5: Commit**

```bash
git add dot/cmd/push.go
git commit -m "feat(push): encrypt secrets before commit"
```

---

### Task 8: Integrate secrets into `dot doctor`

**Files:**
- Modify: `dot/cmd/doctor.go` (add secrets health checks)

- [ ] **Step 1: Add secrets checks after symlink checks**

In `dot/cmd/doctor.go`, inside the module loop (after the `for _, link := range rec.Links` block), add secrets checks. Note: doctor.go uses `issues []string` (slice) per module and `totalIssues` (int) for the global count. `expandHome` is defined in `pull.go` — add a local copy to doctor.go (same implementation).

Add a local `expandHome` at the bottom of doctor.go:

```go
func expandHomePath(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
```

Then add the secrets checks inside the module loop, after the existing link checks block:

```go
		// Check secrets
		mod, parseErr := module.Parse(filepath.Join(dfPath, "modules", modName))
		if parseErr == nil && len(mod.Secrets) > 0 {
			for _, sec := range mod.Secrets {
				plainPath := filepath.Join(mod.Dir, sec.Source)
				encPath := filepath.Join(mod.Dir, sec.Encrypted)
				target := expandHomePath(sec.Target)

				// Check: encrypted file exists
				if _, err := os.Stat(encPath); os.IsNotExist(err) {
					issues = append(issues, fmt.Sprintf("    ✗ %s 加密���件不存在", sec.Encrypted))
					continue
				}

				// Check: plaintext exists (decrypted)
				if _, err := os.Stat(plainPath); os.IsNotExist(err) {
					issues = append(issues, fmt.Sprintf("    ✗ %s 未解密（运行 dot pull %s）", sec.Source, modName))
					continue
				}

				// Check: permissions
				info, _ := os.Stat(plainPath)
				if info.Mode().Perm() != 0o600 {
					issues = append(issues, fmt.Sprintf("    ✗ %s 权限为 %o，应为 0600", sec.Source, info.Mode().Perm()))
				} else {
					fmt.Printf("    ✓ %s 已解密且权限正确\n", sec.Source)
				}

				// Check: symlink to target
				linkInfo, err := os.Lstat(target)
				if err != nil {
					issues = append(issues, fmt.Sprintf("    ✗ %s 符号链接不存在", target))
				} else if linkInfo.Mode()&os.ModeSymlink == 0 {
					issues = append(issues, fmt.Sprintf("    ✗ %s 不是符号链接", target))
				} else {
					fmt.Printf("    ✓ %s → %s\n", sec.Source, target)
				}

				// Check: content consistency (plaintext matches encrypted)
				if secrets.KeychainAvailable() {
					if p, _ := secrets.LoadPassphrase(); p != "" {
						changed, err := secrets.HasChanged(plainPath, encPath, p)
						if err == nil && changed {
							issues = append(issues, fmt.Sprintf("    ✗ %s 与 %s 内容不一致（运行 dot push 同步）", sec.Source, sec.Encrypted))
						} else if err == nil && !changed {
							fmt.Printf("    ✓ %s 与 %s 内容一致\n", sec.Source, sec.Encrypted)
						}
					}
				}
			}
		}
```

- [ ] **Step 2: Add imports**

Add `module` and `secrets` package imports to doctor.go:

```go
import (
	// ... existing imports
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/secrets"
)
```

- [ ] **Step 3: Test manually**

Run: `cd dot && go build -o dot . && ./dot doctor`
Expected: Builds and runs, showing existing module checks.

- [ ] **Step 4: Commit**

```bash
git add dot/cmd/doctor.go
git commit -m "feat(doctor): add secrets health checks"
```

---

### Task 9: Migrate zsh module

**Files:**
- Modify: `modules/zsh/.zshrc` (remove hardcoded tokens)
- Create: `modules/zsh/secrets.env` (plaintext secrets, gitignored)
- Modify: `modules/zsh/config/env.zsh` (add source line)
- Modify: `modules/zsh/module.toml` (add `[[secrets]]`)
- Modify: `.gitignore` (add `secrets.env`)

- [ ] **Step 1: Create secrets.env**

Create `modules/zsh/secrets.env`:

```bash
export GITHUB_PERSONAL_ACCESS_TOKEN="REDACTED_GITHUB_TOKEN"

# >>> claudelike >>>
export ANTHROPIC_BASE_URL="https://claudelike.online/api"
export ANTHROPIC_AUTH_TOKEN="REDACTED_ANTHROPIC_TOKEN"
# <<< claudelike <<<
```

- [ ] **Step 2: Remove tokens from .zshrc**

In `modules/zsh/.zshrc`, delete lines 62-67 (the `GITHUB_PERSONAL_ACCESS_TOKEN`, `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN` exports and the claudelike markers).

Also clean up the duplicate `gt` blocks (lines 75-129 has 4 copies — keep only the last one).

- [ ] **Step 3: Add source line to env.zsh**

Append to `modules/zsh/config/env.zsh`:

```bash

# Secrets (decrypted by dot pull)
[ -f ~/.zsh/secrets.env ] && source ~/.zsh/secrets.env
```

- [ ] **Step 4: Add [[secrets]] to module.toml**

In `modules/zsh/module.toml`, add after the `[[links]]` sections:

```toml

[[secrets]]
source = "secrets.env"
encrypted = "secrets.env.age"
target = "~/.zsh/secrets.env"
```

- [ ] **Step 5: Update .gitignore**

Append to `.gitignore`:

```

# Decrypted secrets (never commit plaintext)
secrets.env
```

- [ ] **Step 6: Verify module parses correctly**

Run: `cd dot && go test ./internal/module/ -v`
Expected: All PASS

- [ ] **Step 7: Build and do initial encryption**

```bash
cd dot && go build -o dot .
./dot push -m "feat: add secrets management"
```

This should trigger the first-time encryption flow for `secrets.env`.

- [ ] **Step 8: Verify**

```bash
# Encrypted file should exist
ls -la modules/zsh/secrets.env.age

# Doctor should pass
./dot doctor

# secrets.env should be gitignored
git status  # secrets.env should NOT appear
```

- [ ] **Step 9: Commit all remaining changes**

```bash
git add modules/zsh/.zshrc modules/zsh/config/env.zsh modules/zsh/module.toml .gitignore modules/zsh/secrets.env.age
git commit -m "feat: migrate zsh secrets to age encryption"
```

---

### Task 10: Clean git history

**Files:** None (git operation)

- [ ] **Step 1: Verify tokens in git history**

Run: `git log --all -p -- modules/zsh/.zshrc | grep -c "ghp_"`
Expected: Count > 0 (tokens exist in history).

- [ ] **Step 2: Install git-filter-repo if needed**

Run: `brew install git-filter-repo` (or `pip install git-filter-repo`)

- [ ] **Step 3: Create expressions file**

Create a temp file `/tmp/dot-filter-expressions.txt`:

```
REDACTED_GITHUB_TOKEN==>REDACTED_GITHUB_TOKEN
REDACTED_ANTHROPIC_TOKEN==>REDACTED_ANTHROPIC_TOKEN
```

- [ ] **Step 4: Run git filter-repo**

```bash
git filter-repo --replace-text /tmp/dot-filter-expressions.txt --force
```

⚠️ This rewrites history. If the repo has a remote, will need `git push --force`.

- [ ] **Step 5: Verify tokens removed**

Run: `git log --all -p -- modules/zsh/.zshrc | grep -c "ghp_"`
Expected: 0

- [ ] **Step 6: Force push (after confirming with user)**

```bash
git push --force
```

- [ ] **Step 7: Revoke old tokens**

- GitHub: https://github.com/settings/tokens → revoke the leaked token
- Anthropic: revoke `cr_ad4b...` token from dashboard
- Generate new tokens and update `modules/zsh/secrets.env`
- Run `dot push` to re-encrypt with new values
