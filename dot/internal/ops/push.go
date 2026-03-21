package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/secrets"
)

// PushChanges commits and pushes dotfiles changes without interactive confirmation.
// It writes progress output to w. Secrets are encrypted using a keychain passphrase
// silently; if unavailable, encryption is skipped with a warning.
func PushChanges(dfPath, msg string, w io.Writer) error {
	// Step 1: check git status, return early if no changes
	changes, err := gitpkg.Status(dfPath)
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if len(changes) == 0 {
		fmt.Fprintln(w, "无变更")
		return nil
	}

	// Step 2: group changes by module for display
	grouped := make(map[string][]string)
	for _, line := range changes {
		if len(line) < 4 {
			continue
		}
		filePath := strings.TrimSpace(line[3:])
		// Handle renames: "old -> new"
		if idx := strings.Index(filePath, " -> "); idx >= 0 {
			filePath = filePath[idx+4:]
		}

		if strings.HasPrefix(filePath, "modules/") {
			parts := strings.SplitN(filePath[len("modules/"):], "/", 2)
			modName := parts[0]
			grouped[modName] = append(grouped[modName], line)
		} else {
			grouped["(其他)"] = append(grouped["(其他)"], line)
		}
	}

	// Sort module names for deterministic output
	var modNames []string
	for name := range grouped {
		modNames = append(modNames, name)
	}
	sort.Strings(modNames)

	fmt.Fprintln(w, "变更摘要：")
	for _, name := range modNames {
		lines := grouped[name]
		fmt.Fprintf(w, "  [%s] (%d 个文件)\n", name, len(lines))
		for _, line := range lines {
			fmt.Fprintf(w, "    %s\n", line)
		}
	}

	// Step 3: encrypt secrets if changed — try keychain passphrase silently,
	// skip with warning if unavailable
	allModules, _ := module.LoadAll(filepath.Join(dfPath, "modules"))
	for _, mod := range allModules {
		if len(mod.Secrets) == 0 {
			continue
		}
		modDir := mod.Dir
		for _, sec := range mod.Secrets {
			plainPath := filepath.Join(modDir, sec.Source)
			encPath := filepath.Join(modDir, sec.Encrypted)

			if _, err := os.Stat(plainPath); os.IsNotExist(err) {
				continue
			}

			// Try keychain passphrase silently; skip with warning if unavailable
			if !secrets.KeychainAvailable() {
				fmt.Fprintf(w, "  警告：跳过加密 %s（无 Keychain）\n", sec.Source)
				continue
			}
			passphrase, err := secrets.LoadPassphrase()
			if err != nil || passphrase == "" {
				fmt.Fprintf(w, "  警告：跳过加密 %s（无法从 Keychain 获取密码短语）\n", sec.Source)
				continue
			}

			changed, err := secrets.HasChanged(plainPath, encPath, passphrase)
			if err != nil {
				return fmt.Errorf("checking secrets: %w", err)
			}

			if changed {
				fmt.Fprintf(w, "  加密 %s → %s\n", sec.Source, sec.Encrypted)
				if err := secrets.EncryptFile(plainPath, encPath, passphrase); err != nil {
					return fmt.Errorf("encrypting %s: %w", sec.Source, err)
				}
			}
		}
	}

	// Step 4: auto-generate commit message if empty
	if msg == "" {
		var mods []string
		for _, name := range modNames {
			if name != "(其他)" {
				mods = append(mods, name)
			}
		}
		if len(mods) > 0 {
			msg = "update " + strings.Join(mods, ", ") + " config"
		} else {
			msg = "update dotfiles"
		}
	}

	// Step 5: git add + commit + push
	fmt.Fprintf(w, "提交：%s\n", msg)
	if err := gitpkg.AddAndCommit(dfPath, []string{"modules/", ".gitignore"}, msg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	fmt.Fprintln(w, "推送到远程...")
	if err := gitpkg.Push(dfPath); err != nil {
		return fmt.Errorf("git push: %w", err)
	}

	fmt.Fprintln(w, "完成。")
	return nil
}
