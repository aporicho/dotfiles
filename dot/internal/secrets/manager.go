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

// GetPassphrase tries keychain first, falls back to askFunc.
func GetPassphrase(askFunc func(string, bool) (string, error)) (string, error) {
	if KeychainAvailable() {
		if p, _ := LoadPassphrase(); p != "" {
			return p, nil
		}
	}
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

// OfferSaveToKeychain saves passphrase to keychain if available.
func OfferSaveToKeychain(passphrase string) {
	if !KeychainAvailable() {
		return
	}
	SavePassphrase(passphrase)
}
