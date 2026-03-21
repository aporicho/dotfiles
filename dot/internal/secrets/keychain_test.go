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
