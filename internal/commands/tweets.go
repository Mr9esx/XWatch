package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func NewTweetsCmd(deps *Deps) *cobra.Command {
	var (
		user       string
		tag        string
		since      string
		limit      int
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "tweets",
		Short: "查询已缓存的推文",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			sinceTime, err := parseDuration(since)
			if err != nil {
				return fmt.Errorf("无效的时间范围: %w", err)
			}

			if user != "" && tag != "" {
				return fmt.Errorf("--user 和 --tag 不能同时使用")
			}

			var tweets interface{}
			switch {
			case user != "":
				tweets, err = deps.Store.GetTweets(ctx, user, sinceTime, limit)
			case tag != "":
				tweets, err = deps.Store.GetTweetsByTag(ctx, tag, sinceTime, limit)
			default:
				tweets, err = deps.Store.GetAllTweets(ctx, sinceTime, limit)
			}
			if err != nil {
				return fmt.Errorf("查询推文失败: %w", err)
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(tweets, "", "  ")
				fmt.Println(string(data))
			} else {
				data, _ := json.MarshalIndent(tweets, "", "  ")
				fmt.Println(string(data))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&user, "user", "", "按用户筛选 (screen_name)")
	cmd.Flags().StringVar(&tag, "tag", "", "按主题标签筛选")
	cmd.Flags().StringVar(&since, "since", "24h", "时间范围 (如 1h, 24h, 7d)")
	cmd.Flags().IntVar(&limit, "limit", 100, "最大返回条数")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

func parseDuration(s string) (time.Time, error) {
	if len(s) == 0 {
		return time.Now().Add(-24 * time.Hour), nil
	}

	last := s[len(s)-1]
	numStr := s[:len(s)-1]

	var multiplier time.Duration
	switch last {
	case 's':
		multiplier = time.Second
	case 'm':
		multiplier = time.Minute
	case 'h':
		multiplier = time.Hour
	case 'd':
		multiplier = 24 * time.Hour
	case 'w':
		multiplier = 7 * 24 * time.Hour
	default:
		return time.Time{}, fmt.Errorf("未知的时间单位: %c (支持 s/m/h/d/w)", last)
	}

	var num int
	if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
		return time.Time{}, fmt.Errorf("无效的数字: %s", numStr)
	}

	return time.Now().Add(-time.Duration(num) * multiplier), nil
}
