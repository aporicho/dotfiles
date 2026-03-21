package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initTestRepo creates a temporary git repository with one committed file.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "commit.gpgSign", "false"},
	}
	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("initTestRepo %v: %s: %v", args, out, err)
		}
	}

	// Create and commit a file so we have a valid HEAD.
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("initTestRepo %v: %s: %v", args, out, err)
		}
	}

	return dir
}

func TestStatus_Clean(t *testing.T) {
	dir := initTestRepo(t)

	lines, err := Status(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected empty status, got %v", lines)
	}
}

func TestStatus_Modified(t *testing.T) {
	dir := initTestRepo(t)

	// Modify the committed file so the working tree is dirty.
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	lines, err := Status(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("expected non-empty status for modified file")
	}
}

func TestAddCommit(t *testing.T) {
	dir := initTestRepo(t)

	// Create a new file.
	newFile := filepath.Join(dir, "new.txt")
	if err := os.WriteFile(newFile, []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Add and commit.
	if err := AddAndCommit(dir, []string{"new.txt"}, "add new file"); err != nil {
		t.Fatalf("AddAndCommit: %v", err)
	}

	// Status should be clean now.
	lines, err := Status(dir)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected clean status after commit, got %v", lines)
	}
}

func TestPull_NoRemote(t *testing.T) {
	dir := initTestRepo(t)

	// Pull without a remote configured — should return an error but not panic.
	err := Pull(dir)
	if err == nil {
		// Some git versions may not error; that's acceptable too.
		return
	}

	// If it did error, ensure it's a *GitError.
	gitErr, ok := err.(*GitError)
	if !ok {
		t.Fatalf("expected *GitError, got %T: %v", err, err)
	}
	if gitErr.Op != "pull" {
		t.Fatalf("expected op \"pull\", got %q", gitErr.Op)
	}
}
