package secrets

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestHasChanged(t *testing.T) {
	dir := t.TempDir()
	plainPath := filepath.Join(dir, "secrets.env")
	encPath := filepath.Join(dir, "secrets.env.age")
	passphrase := "test"

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
