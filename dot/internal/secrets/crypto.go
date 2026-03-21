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
func HasChanged(plaintextPath, encryptedPath, passphrase string) (bool, error) {
	plaintext, err := os.ReadFile(plaintextPath)
	if err != nil {
		return false, err
	}
	ciphertext, err := os.ReadFile(encryptedPath)
	if err != nil {
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
