package tui

import (
	"strings"
	"testing"
)

func TestPassphraseModelMasking(t *testing.T) {
	m := NewPassphraseModel("Enter passphrase:", false)
	m.value = "hello"
	view := m.View()
	if strings.Contains(view, "hello") {
		t.Error("passphrase should be masked in view")
	}
	if !strings.Contains(view, "•••••") {
		t.Error("view should contain mask dots")
	}
}
