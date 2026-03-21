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

	os.Remove(plain)

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
