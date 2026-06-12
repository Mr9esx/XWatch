package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mr9esx/xwatch/internal/store"
	"github.com/spf13/cobra"
)

func NewListCmd(deps *Deps) *cobra.Command {
	var tag string
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有订阅的用户",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var users []store.WatchedUser
			var err error
			if tag != "" {
				users, err = deps.Store.GetWatchedUsersByTag(ctx, tag)
			} else {
				users, err = deps.Store.GetWatchedUsers(ctx)
			}
			if err != nil {
				return fmt.Errorf("获取订阅列表失败: %w", err)
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(users, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if len(users) == 0 {
				if tag != "" {
					fmt.Printf("主题「%s」下没有订阅用户\n", tag)
				} else {
					fmt.Println("当前没有订阅任何用户")
					fmt.Println("使用 xwatch subscribe <screen_name> 来订阅")
				}
				return nil
			}

			defaultInterval := deps.Config.GetDefaultCheckInterval(ctx)
			if tag != "" {
				fmt.Printf("主题「%s」共 %d 位用户：\n\n", tag, len(users))
			} else {
				fmt.Printf("当前订阅 %d 位用户：\n\n", len(users))
			}

			for _, u := range users {
				interval := u.CheckInterval
				intervalStr := ""
				if interval == 0 {
					interval = defaultInterval
					intervalStr = fmt.Sprintf("默认(%s)", formatDuration(interval))
				} else {
					intervalStr = formatDuration(interval)
				}

				lastChecked := "从未"
				if !u.LastCheckedAt.IsZero() && u.LastCheckedAt.Year() > 1970 {
					ago := time.Since(u.LastCheckedAt)
					lastChecked = formatAgo(ago)
				}

				tagStr := ""
				if len(u.Tags) > 0 {
					tagStr = fmt.Sprintf("  [%s]", strings.Join(u.Tags, ", "))
				}
				fmt.Printf("  @%-15s (%s)%s  每%s  最后检查: %s\n",
					u.ScreenName, u.Name, tagStr, intervalStr, lastChecked)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&tag, "tag", "", "按主题标签筛选")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

func formatDuration(seconds int) string {
	switch {
	case seconds >= 3600:
		return fmt.Sprintf("%d小时", seconds/3600)
	case seconds >= 60:
		return fmt.Sprintf("%d分钟", seconds/60)
	default:
		return fmt.Sprintf("%d秒", seconds)
	}
}

func formatAgo(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%d秒前", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%d分钟前", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d小时前", int(d.Hours()))
	default:
		return fmt.Sprintf("%d天前", int(d.Hours()/24))
	}
}
