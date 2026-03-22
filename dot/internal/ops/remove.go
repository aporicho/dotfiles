package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
)

// RemoveModuleFiles permanently deletes a module's directory from modules/.
// The module must not be installed (check manifest first).
func RemoveModuleFiles(dfPath, modName string, w io.Writer) error {
	// Step 1: load manifest and check module is not installed
	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	if mf.IsInstalled(modName) {
		return fmt.Errorf("模块 %q 当前已安装，请先运行 dot uninstall %s", modName, modName)
	}

	// Step 2: check that the module directory exists
	modDir := filepath.Join(dfPath, "modules", modName)
	if _, err := os.Stat(modDir); os.IsNotExist(err) {
		return fmt.Errorf("模块目录不存在：%s", modDir)
	}

	// Step 3: delete the module directory permanently
	fmt.Fprintf(w, "  正在删除 %s ...\n", modDir)
	if err := os.RemoveAll(modDir); err != nil {
		return fmt.Errorf("删除模块目录失败：%w", err)
	}

	fmt.Fprintf(w, "  ✓ %s：模块文件已永久删除\n", modName)

	return nil
}
