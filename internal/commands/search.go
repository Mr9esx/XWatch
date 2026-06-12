package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func NewSearchCmd(deps *Deps) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "搜索 X 用户",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps.Fetcher == nil {
				return fmt.Errorf("Twitter 认证未配置")
			}

			users, err := deps.Fetcher.SearchUsers(cmd.Context(), args[0], 10)
			if err != nil {
				return fmt.Errorf("搜索失败: %w", err)
			}

			if len(users) == 0 {
				fmt.Println("未找到匹配的用户")
				return nil
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(users, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("找到 %d 个匹配用户：\n\n", len(users))
			for i, u := range users {
				fmt.Printf("%d. @%s (%s) — %s followers\n", i+1, u.ScreenName, u.Name, formatCount(u.FollowersCount))
				if u.Description != "" {
					desc := u.Description
					if len(desc) > 80 {
						desc = desc[:80] + "..."
					}
					fmt.Printf("   %s\n", desc)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

func formatCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
