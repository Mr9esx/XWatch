package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mr9esx/xwatch/internal/store"
	"github.com/spf13/cobra"
)

func NewSubscribeCmd(deps *Deps) *cobra.Command {
	var interval int
	var tags []string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "subscribe <screen_name>",
		Short: "订阅 X 用户的推文",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps.Fetcher == nil {
				return fmt.Errorf("Twitter 认证未配置")
			}

			screenName := args[0]
			ctx := cmd.Context()

			existing, err := deps.Store.GetWatchedUser(ctx, screenName)
			if err != nil {
				return err
			}
			if existing != nil {
				return fmt.Errorf("已经订阅了 @%s", screenName)
			}

			user, err := deps.Fetcher.GetUserByScreenName(ctx, screenName)
			if err != nil {
				return fmt.Errorf("获取用户信息失败: %w", err)
			}

			wu := store.WatchedUser{
				UserID:        user.ID,
				ScreenName:    user.ScreenName,
				Name:          user.Name,
				Tags:          tags,
				CheckInterval: interval,
			}
			if err := deps.Store.AddWatchedUser(ctx, wu); err != nil {
				return fmt.Errorf("订阅失败: %w", err)
			}

			if jsonOutput {
				result := map[string]interface{}{
					"success":     true,
					"screen_name": user.ScreenName,
					"name":        user.Name,
					"user_id":     user.ID,
					"interval":    interval,
					"tags":        tags,
				}
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("已订阅 @%s (%s)\n", user.ScreenName, user.Name)
			if len(tags) > 0 {
				fmt.Printf("主题标签：%s\n", strings.Join(tags, ", "))
			}
			if interval > 0 {
				fmt.Printf("检查频率：每 %d 秒\n", interval)
			} else {
				defaultInterval := deps.Config.GetDefaultCheckInterval(ctx)
				fmt.Printf("检查频率：默认（每 %d 秒）\n", defaultInterval)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&interval, "interval", 0, "检查频率（秒），0 表示使用全局默认")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "主题标签（可多次指定或逗号分隔）")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}
