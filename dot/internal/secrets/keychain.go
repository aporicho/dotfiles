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
