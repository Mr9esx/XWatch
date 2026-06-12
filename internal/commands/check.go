package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/mr9esx/xwatch/internal/fetcher"
	"github.com/mr9esx/xwatch/internal/store"
	"github.com/spf13/cobra"
)

func NewCheckCmd(deps *Deps) *cobra.Command {
	var (
		force bool
		tag   string
	)
	cmd := &cobra.Command{
		Use:   "check",
		Short: "检查订阅用户的新推文",
		Long:  "遍历订阅用户，拉取新推文并存储。供 Hermes cron 调用，Hermes 负责翻译和格式化。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(deps, force, tag)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "忽略检查间隔，强制检查所有用户")
	cmd.Flags().StringVar(&tag, "tag", "", "仅检查带有该主题标签的用户")
	return cmd
}

func runCheck(deps *Deps, force bool, tag string) error {
	if deps.Fetcher == nil {
		return fmt.Errorf("Twitter 认证未配置，请先运行: xwatch config set x_auth_token <token>")
	}

	ctx := context.Background()
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

	if len(users) == 0 {
		fmt.Println("[SILENT]")
		return nil
	}

	defaultInterval := deps.Config.GetDefaultCheckInterval(ctx)
	now := time.Now()
	var allTweets []tweetOutput

	rateLimitHits := 0
	const maxRateLimitHits = 3

	for i, user := range users {
		interval := user.CheckInterval
		if interval == 0 {
			interval = defaultInterval
		}

		if !force && now.Sub(user.LastCheckedAt) < time.Duration(interval)*time.Second {
			continue
		}

		if rateLimitHits >= maxRateLimitHits {
			fmt.Fprintf(os.Stderr, "警告: 连续触发 %d 次限流，跳过剩余 %d 位用户\n", rateLimitHits, len(users)-i)
			break
		}

		tweets, err := deps.Fetcher.GetUserTweets(ctx, user.ScreenName, user.LatestTweetID, 20)
		if err != nil {
			if fetcher.IsRateLimitError(err) {
				rateLimitHits++
				wait := time.Duration(30*(1<<rateLimitHits)) * time.Second
				fmt.Fprintf(os.Stderr, "警告: 获取 @%s 触发限流 (429)，等待 %s 后继续\n", user.ScreenName, wait)
				select {
				case <-time.After(wait):
				case <-ctx.Done():
					return ctx.Err()
				}
				continue
			}
			if fetcher.IsAuthError(err) {
				return fmt.Errorf("认证失败 (获取 @%s 时): %w\n请检查 x_auth_token / x_ct0 是否过期", user.ScreenName, err)
			}
			fmt.Fprintf(os.Stderr, "警告: 获取 @%s 的推文失败: %v\n", user.ScreenName, err)
			continue
		}

		rateLimitHits = 0

		if len(tweets) == 0 {
			deps.Store.UpdateLastChecked(ctx, user.ScreenName, user.LatestTweetID)
			jitter := time.Duration(1000+rand.Intn(2000)) * time.Millisecond
			time.Sleep(jitter)
			continue
		}

		var latestID string
		for _, tweet := range tweets {
			mediaURLsJSON, _ := json.Marshal(tweet.MediaURLs)
			deps.Store.SaveTweet(ctx, store.StoredTweet{
				TweetID:      tweet.ID,
				UserID:       tweet.UserID,
				ScreenName:   user.ScreenName,
				OriginalText: tweet.Text,
				TweetURL:     tweet.URL,
				MediaURLs:    string(mediaURLsJSON),
				CreatedAt:    tweet.CreatedAt,
			})

			allTweets = append(allTweets, tweetOutput{
				ScreenName:          user.ScreenName,
				Name:                user.Name,
				Text:                tweet.Text,
				URL:                 tweet.URL,
				CreatedAt:           tweet.CreatedAt.Local().Format("2006-01-02 15:04"),
				IsRetweet:           tweet.IsRetweet,
				RetweetedScreenName: tweet.RetweetedScreenName,
			})

			if latestID == "" || tweet.ID > latestID {
				latestID = tweet.ID
			}
		}

		if latestID != "" {
			deps.Store.UpdateLastChecked(ctx, user.ScreenName, latestID)
		} else {
			deps.Store.UpdateLastChecked(ctx, user.ScreenName, user.LatestTweetID)
		}

		if i < len(users)-1 {
			jitter := time.Duration(1000+rand.Intn(2000)) * time.Millisecond
			time.Sleep(jitter)
		}
	}

	if len(allTweets) == 0 {
		fmt.Println("[SILENT]")
	} else {
		var parts []string
		for _, t := range allTweets {
			if t.IsRetweet {
				parts = append(parts, fmt.Sprintf("🔁 @%s (%s) 于 %s 转推了 @%s：\n%s\n🔗 %s",
					t.ScreenName, t.Name, t.CreatedAt, t.RetweetedScreenName, t.Text, t.URL))
			} else {
				parts = append(parts, fmt.Sprintf("@%s (%s) 于 %s 发布：\n%s\n🔗 %s",
					t.ScreenName, t.Name, t.CreatedAt, t.Text, t.URL))
			}
		}
		fmt.Println(strings.Join(parts, "\n---\n"))
	}

	return nil
}

type tweetOutput struct {
	ScreenName          string
	Name                string
	Text                string
	URL                 string
	CreatedAt           string
	IsRetweet           bool
	RetweetedScreenName string
}
