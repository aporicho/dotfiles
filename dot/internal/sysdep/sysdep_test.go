package sysdep

import (
	"runtime"
	"testing"
)

func TestBuildCommand_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping: not darwin")
	}

	cmd := buildCommand("brew", []string{"git", "curl"})
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	args := cmd.Args
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d: %v", len(args), args)
	}
	if args[0] != "brew" || args[1] != "install" || args[2] != "git" || args[3] != "curl" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestBuildCommand_Empty(t *testing.T) {
	if cmd := buildCommand("brew", nil); cmd != nil {
		t.Fatalf("expected nil for nil packages, got %v", cmd)
	}
	if cmd := buildCommand("brew", []string{}); cmd != nil {
		t.Fatalf("expected nil for empty packages, got %v", cmd)
	}
}
