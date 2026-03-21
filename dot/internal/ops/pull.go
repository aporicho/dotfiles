package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aporicho/dotfiles/dot/internal/deps"
	"github.com/aporicho/dotfiles/dot/internal/hook"
	"github.com/aporicho/dotfiles/dot/internal/linker"
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/platform"
	"github.com/aporicho/dotfiles/dot/internal/secrets"
	"github.com/aporicho/dotfiles/dot/internal/sysdep"
)

// PullModule installs a single named module and its dependencies.
// It writes progress output to w. Unlike cmd/pull.go, it has no interactive
// TUI elements: passphrases are retrieved from the keychain only; if none is
// available, encrypted secrets are skipped with a warning.
func PullModule(dfPath, modName string, w io.Writer) error {
	// Step 1: load all modules from modules/ directory
	allModules, err := module.LoadAll(filepath.Join(dfPath, "modules"))
	if err != nil {
		return fmt.Errorf("loading modules: %w", err)
	}

	// Step 2: build maps, filtering by platform
	modMap := make(map[string]*module.Module)
	depsMap := make(map[string][]string)

	for _, mod := range allModules {
		if !platform.MatchesPlatform(mod.Platforms) {
			continue
		}
		modMap[mod.Name] = mod
		depsMap[mod.Name] = mod.Requires
	}

	// Step 3: validate the requested module exists
	if _, ok := modMap[modName]; !ok {
		return fmt.Errorf("模块 %q 不存在或不适用于当前平台", modName)
	}

	// Step 4: resolve dependency order
	order, err := deps.Resolve([]string{modName}, depsMap)
	if err != nil {
		return fmt.Errorf("dependency resolution: %w", err)
	}

	// Step 5: load manifest
	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}
	mf.DotfilesPath = dfPath

	// Step 6: install each module in dependency order
	for _, name := range order {
		mod := modMap[name]
		fmt.Fprintf(w, "→ 安装 %s ...\n", name)
		if err := installModuleOps(dfPath, mod, mf, w); err != nil {
			return fmt.Errorf("installing %s: %w", name, err)
		}
		fmt.Fprintf(w, "  ✓ %s 安装完成\n", name)
	}

	// Step 7: save manifest
	if err := manifest.Save(mf, mfPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	return nil
}

// installModuleOps runs hooks, installs system deps, creates symlinks, handles
// secrets (keychain only, no interactive prompt), and records in manifest.
// On failure it rolls back created links.
func installModuleOps(dfPath string, mod *module.Module, mf *manifest.Manifest, w io.Writer) error {
	modDir := mod.Dir
	var createdLinks []*linker.LinkResult

	// rollback helper
	rollback := func() {
		for _, lr := range createdLinks {
			_ = linker.RemoveLink(lr.Target)
			if lr.BackupPath != "" {
				_ = os.Rename(lr.BackupPath, lr.Target)
			}
		}
	}

	// Step 1: pre_install hook
	if err := hook.Run(mod.Hooks.PreInstall, modDir); err != nil {
		rollback()
		return fmt.Errorf("pre_install hook: %w", err)
	}

	// Step 2: system deps (warn on failure, don't abort)
	if err := sysdep.Install(mod.Deps.Darwin, mod.Deps.Linux); err != nil {
		fmt.Fprintf(w, "  警告：系统依赖安装失败：%v\n", err)
	}

	// Step 3: create symlinks
	for _, link := range mod.Links {
		// Skip links not matching current platform
		if !platform.MatchesPlatform(link.Platforms) {
			continue
		}

		target := expandHomeOps(link.Target)

		if link.Source == "." {
			// Expand directory entries
			exclude := mod.Exclude
			if exclude == nil {
				exclude = []string{"module.toml", "README*", "LICENSE*", ".gitignore"}
			}
			entries, err := linker.ExpandDirEntries(modDir, exclude)
			if err != nil {
				rollback()
				return fmt.Errorf("expanding dir entries: %w", err)
			}
			for _, entry := range entries {
				entryTarget := filepath.Join(target, filepath.Base(entry))
				lr, err := linker.CreateLink(entry, entryTarget)
				if err != nil {
					rollback()
					return fmt.Errorf("linking %s: %w", entry, err)
				}
				createdLinks = append(createdLinks, lr)
			}
		} else {
			source := filepath.Join(modDir, link.Source)
			lr, err := linker.CreateLink(source, target)
			if err != nil {
				rollback()
				return fmt.Errorf("linking %s -> %s: %w", source, target, err)
			}
			createdLinks = append(createdLinks, lr)
		}
	}

	// Handle secrets: try keychain passphrase silently; skip if unavailable
	for _, sec := range mod.Secrets {
		encPath := filepath.Join(modDir, sec.Encrypted)
		plainPath := filepath.Join(modDir, sec.Source)
		target := expandHomeOps(sec.Target)

		// Skip if no encrypted file exists
		if _, err := os.Stat(encPath); os.IsNotExist(err) {
			continue
		}

		// Decrypt if plaintext doesn't exist
		if _, err := os.Stat(plainPath); os.IsNotExist(err) {
			// Try keychain silently; skip with warning if not available
			if !secrets.KeychainAvailable() {
				fmt.Fprintf(w, "  警告：跳过 %s（无 Keychain 密码短语）\n", sec.Encrypted)
				continue
			}
			passphrase, err := secrets.LoadPassphrase()
			if err != nil || passphrase == "" {
				fmt.Fprintf(w, "  警告：跳过 %s（无法从 Keychain 获取密码短语）\n", sec.Encrypted)
				continue
			}
			fmt.Fprintf(w, "  解密 %s\n", sec.Encrypted)
			if err := secrets.DecryptFile(encPath, plainPath, passphrase); err != nil {
				return fmt.Errorf("decrypting %s: %w", sec.Encrypted, err)
			}
		}

		// Create symlink for secret
		lr, err := linker.CreateLink(plainPath, target)
		if err != nil {
			return fmt.Errorf("linking secret %s: %w", sec.Source, err)
		}
		createdLinks = append(createdLinks, lr)
	}

	// Step 4: post_install hook
	if err := hook.Run(mod.Hooks.PostInstall, modDir); err != nil {
		rollback()
		return fmt.Errorf("post_install hook: %w", err)
	}

	// Step 5: record in manifest
	var records []manifest.LinkRecord
	for _, lr := range createdLinks {
		records = append(records, manifest.LinkRecord{
			Source: lr.Source,
			Target: lr.Target,
			Backup: lr.BackupPath,
		})
	}
	mf.AddModule(mod.Name, records)

	return nil
}

// expandHomeOps replaces a leading "~/" with the user's home directory.
func expandHomeOps(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
