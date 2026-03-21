package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
)

var pushMessage string

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Commit and push module changes to remote",
	RunE:  runPush,
}

func init() {
	pushCmd.Flags().StringVarP(&pushMessage, "message", "m", "", "Custom commit message")
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	dfPath, err := DotfilesPath()
	if err != nil {
		return err
	}

	// Step 1: check git status
	changes, err := gitpkg.Status(dfPath)
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if len(changes) == 0 {
		fmt.Println("无变更")
		return nil
	}

	// Step 2: group changes by module
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

	// Step 3: display grouped changes
	var modNames []string
	for name := range grouped {
		modNames = append(modNames, name)
	}
	sort.Strings(modNames)

	fmt.Println("变更摘要：")
	for _, name := range modNames {
		lines := grouped[name]
		fmt.Printf("  [%s] (%d 个文件)\n", name, len(lines))
		for _, line := range lines {
			fmt.Printf("    %s\n", line)
		}
	}

	// Step 4: ask confirmation
	fmt.Print("\n确认提交并推送？(Y/n) ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "n" || answer == "no" {
		fmt.Println("取消操作")
		return nil
	}

	// Step 5: build commit message
	msg := pushMessage
	if msg == "" {
		// Auto-generate: "update <mod1>, <mod2> config"
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

	// Step 6: commit and push
	fmt.Printf("提交：%s\n", msg)
	if err := gitpkg.AddAndCommit(dfPath, []string{"modules/"}, msg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	fmt.Println("推送到远程...")
	if err := gitpkg.Push(dfPath); err != nil {
		return fmt.Errorf("git push: %w", err)
	}

	fmt.Println("完成。")
	return nil
}
