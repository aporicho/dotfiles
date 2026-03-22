package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitError represents an error from a git operation.
type GitError struct {
	Op     string
	Output string
	Err    error
}

func (e *GitError) Error() string { return "git " + e.Op + ": " + e.Output }
func (e *GitError) Unwrap() error { return e.Err }

// run executes a git command in the given directory and returns combined output.
func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Status runs `git status --porcelain` and returns a list of changed file lines.
// A clean repo returns an empty slice.
func Status(repoDir string) ([]string, error) {
	out, err := run(repoDir, "status", "--porcelain")
	if err != nil {
		return nil, &GitError{Op: "status", Output: strings.TrimSpace(out), Err: err}
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// AddAndCommit stages the given paths and creates a commit with the given message.
func AddAndCommit(repoDir string, paths []string, message string) error {
	// git add <paths>
	args := append([]string{"add"}, paths...)
	out, err := run(repoDir, args...)
	if err != nil {
		return &GitError{Op: "add", Output: strings.TrimSpace(out), Err: err}
	}

	// git commit -m <message>
	out, err = run(repoDir, "commit", "-m", message)
	if err != nil {
		return &GitError{Op: "commit", Output: strings.TrimSpace(out), Err: err}
	}
	return nil
}

// Push runs `git push` in the given repository directory.
func Push(repoDir string) error {
	out, err := run(repoDir, "push")
	if err != nil {
		return &GitError{Op: "push", Output: strings.TrimSpace(out), Err: err}
	}
	return nil
}

// Pull runs `git pull` in the given repository directory.
func Pull(repoDir string) error {
	out, err := run(repoDir, "pull")
	if err != nil {
		return &GitError{Op: "pull", Output: strings.TrimSpace(out), Err: err}
	}
	return nil
}

// AheadBehind returns how many commits HEAD is ahead of and behind its upstream.
// Returns (0, 0, nil) when the branch has no upstream or on any error.
func AheadBehind(repoDir string) (ahead, behind int, err error) {
	out, err := run(repoDir, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if err != nil {
		return 0, 0, nil
	}
	out = strings.TrimSpace(out)
	parts := strings.Fields(out)
	if len(parts) != 2 {
		return 0, 0, nil
	}
	fmt.Sscanf(parts[0], "%d", &ahead)
	fmt.Sscanf(parts[1], "%d", &behind)
	return ahead, behind, nil
}
