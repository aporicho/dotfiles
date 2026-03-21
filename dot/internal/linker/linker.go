package linker

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LinkResult holds the outcome of a symlink creation.
type LinkResult struct {
	Source     string
	Target     string
	BackupPath string // empty if no backup was made
}

// CreateLink creates a symlink from source to target.
// It auto-creates parent directories for the target path.
// If target exists as a symlink, it is removed and replaced (no backup).
// If target exists as a regular file or directory, it is backed up
// to {target}.backup.{timestamp} before creating the symlink.
func CreateLink(source, target string) (*LinkResult, error) {
	// Ensure parent directory exists
	parentDir := filepath.Dir(target)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir parent %s: %w", parentDir, err)
	}

	result := &LinkResult{
		Source: source,
		Target: target,
	}

	// Check if target already exists
	info, err := os.Lstat(target)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			// Existing symlink → remove and replace, no backup
			if err := os.Remove(target); err != nil {
				return nil, fmt.Errorf("remove existing symlink %s: %w", target, err)
			}
		} else {
			// Regular file or directory → backup then replace
			timestamp := time.Now().Format("20060102150405")
			backupPath := fmt.Sprintf("%s.backup.%s", target, timestamp)
			if err := os.Rename(target, backupPath); err != nil {
				return nil, fmt.Errorf("backup %s to %s: %w", target, backupPath, err)
			}
			result.BackupPath = backupPath
		}
	}

	// Create symlink
	if err := os.Symlink(source, target); err != nil {
		return nil, fmt.Errorf("symlink %s -> %s: %w", source, target, err)
	}

	return result, nil
}

// ExpandDirEntries returns first-level entries in dir, excluding files whose
// names match any of the given patterns. Pattern matching uses filepath.Match
// against the base name of each entry.
func ExpandDirEntries(dir string, exclude []string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	var entries []string
	for _, entry := range dirEntries {
		name := entry.Name()
		excluded := false
		for _, pattern := range exclude {
			matched, err := filepath.Match(pattern, name)
			if err != nil {
				return nil, fmt.Errorf("bad pattern %q: %w", pattern, err)
			}
			if matched {
				excluded = true
				break
			}
		}
		if !excluded {
			entries = append(entries, filepath.Join(dir, name))
		}
	}

	return entries, nil
}

// RemoveLink removes a symlink at target.
// If target does not exist, it returns nil (already gone).
// If target exists but is not a symlink, it returns an error.
func RemoveLink(target string) error {
	info, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // already gone
		}
		return fmt.Errorf("stat %s: %w", target, err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("refusing to remove %s: not a symlink", target)
	}

	if err := os.Remove(target); err != nil {
		return fmt.Errorf("remove symlink %s: %w", target, err)
	}
	return nil
}
