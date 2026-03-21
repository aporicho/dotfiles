package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_Success(t *testing.T) {
	err := Run("echo hello", t.TempDir())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRun_Empty(t *testing.T) {
	// Empty string should be a no-op.
	for _, cmd := range []string{"", "   ", "\t\n"} {
		if err := Run(cmd, t.TempDir()); err != nil {
			t.Fatalf("expected no error for empty/whitespace command %q, got %v", cmd, err)
		}
	}
}

func TestRun_Failure(t *testing.T) {
	err := Run("exit 1", t.TempDir())
	if err == nil {
		t.Fatal("expected error for exit 1, got nil")
	}
}

func TestRun_WorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	err := Run("pwd > out.txt", dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "out.txt"))
	if err != nil {
		t.Fatalf("failed to read out.txt: %v", err)
	}

	got := strings.TrimSpace(string(data))
	if got != dir {
		t.Fatalf("expected working directory %q, got %q", dir, got)
	}
}
