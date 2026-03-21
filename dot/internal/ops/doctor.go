package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/secrets"
)

// DoctorCheck checks the health of a single named module.
// It loads the manifest, verifies symlinks and secrets, and reports results to w.
// Issues are reported but are not fatal — the function returns nil unless a
// hard error (e.g. manifest load failure) occurs.
func DoctorCheck(dfPath, modName string, w io.Writer) error {
	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	rec, ok := mf.Modules[modName]
	if !ok {
		fmt.Fprintf(w, "  %s：未在 manifest 中找到\n", modName)
		return nil
	}

	var issues []string

	// Check symlinks recorded in the manifest
	for _, link := range rec.Links {
		target := link.Target
		expectedSource := link.Source
		if !filepath.IsAbs(expectedSource) {
			expectedSource = filepath.Join(dfPath, expectedSource)
		}

		info, err := os.Lstat(target)
		if err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("    ✗ %s：链接不存在", target))
			} else {
				issues = append(issues, fmt.Sprintf("    ✗ %s：无法访问: %v", target, err))
			}
			continue
		}

		if info.Mode()&os.ModeSymlink == 0 {
			issues = append(issues, fmt.Sprintf("    ✗ %s：不是符号链接（已被替换为普通文件）", target))
			continue
		}

		actual, err := os.Readlink(target)
		if err != nil {
			issues = append(issues, fmt.Sprintf("    ✗ %s：无法读取链接目标: %v", target, err))
			continue
		}

		if !filepath.IsAbs(actual) {
			actual = filepath.Join(filepath.Dir(target), actual)
		}

		actualClean := filepath.Clean(actual)
		expectedClean := filepath.Clean(expectedSource)

		if actualClean != expectedClean {
			issues = append(issues, fmt.Sprintf("    ✗ %s：指向 %s（应为 %s）", target, actualClean, expectedClean))
			continue
		}

		if _, err := os.Stat(expectedSource); err != nil {
			issues = append(issues, fmt.Sprintf("    ✗ %s：源文件不存在 %s", target, expectedSource))
			continue
		}
	}

	// Check secrets defined in the module definition
	mod, parseErr := module.Parse(filepath.Join(dfPath, "modules", modName))
	if parseErr == nil && len(mod.Secrets) > 0 {
		for _, sec := range mod.Secrets {
			plainPath := filepath.Join(mod.Dir, sec.Source)
			encPath := filepath.Join(mod.Dir, sec.Encrypted)
			target := expandHomeOps(sec.Target)

			if _, err := os.Stat(encPath); os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("    ✗ %s 加密文件不存在", sec.Encrypted))
				continue
			}

			if _, err := os.Stat(plainPath); os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("    ✗ %s 未解密（运行 dot pull %s）", sec.Source, modName))
				continue
			}

			info, _ := os.Stat(plainPath)
			if info.Mode().Perm() != 0o600 {
				issues = append(issues, fmt.Sprintf("    ✗ %s 权限为 %o，应为 0600", sec.Source, info.Mode().Perm()))
			} else {
				fmt.Fprintf(w, "    ✓ %s 已解密且权限正确\n", sec.Source)
			}

			linkInfo, err := os.Lstat(target)
			if err != nil {
				issues = append(issues, fmt.Sprintf("    ✗ %s 符号链接不存在", target))
			} else if linkInfo.Mode()&os.ModeSymlink == 0 {
				issues = append(issues, fmt.Sprintf("    ✗ %s 不是符号链接", target))
			} else {
				fmt.Fprintf(w, "    ✓ %s → %s\n", sec.Source, target)
			}

			// Content consistency check (keychain only)
			if secrets.KeychainAvailable() {
				if p, _ := secrets.LoadPassphrase(); p != "" {
					changed, err := secrets.HasChanged(plainPath, encPath, p)
					if err == nil && changed {
						issues = append(issues, fmt.Sprintf("    ✗ %s 与 %s 内容不一致（运行 dot push 同步）", sec.Source, sec.Encrypted))
					} else if err == nil && !changed {
						fmt.Fprintf(w, "    ✓ %s 与 %s 内容一致\n", sec.Source, sec.Encrypted)
					}
				}
			}
		}
	}

	if len(issues) == 0 {
		fmt.Fprintf(w, "  %s：✓ 正常\n", modName)
	} else {
		fmt.Fprintf(w, "  %s：发现 %d 个问题\n", modName, len(issues))
		for _, issue := range issues {
			fmt.Fprintln(w, issue)
		}
	}

	return nil
}
