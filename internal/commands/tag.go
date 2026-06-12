package commands

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func NewTagCmd(deps *Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "管理订阅用户的主题标签",
	}
	cmd.AddCommand(newTagAddCmd(deps))
	cmd.AddCommand(newTagRemoveCmd(deps))
	cmd.AddCommand(newTagSetCmd(deps))
	cmd.AddCommand(newTagListCmd(deps))
	return cmd
}

func newTagAddCmd(deps *Deps) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "add <screen_name> <tag> [tag...]",
		Short: "为订阅用户添加主题标签",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			screenName := args[0]
			tags := args[1:]

			if err := deps.Store.AddUserTags(ctx, screenName, tags); err != nil {
				return err
			}

			user, err := deps.Store.GetWatchedUser(ctx, screenName)
			if err != nil {
				return err
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(user, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("已为 @%s 添加标签：%s\n", screenName, strings.Join(tags, ", "))
			fmt.Printf("当前标签：%s\n", strings.Join(user.Tags, ", "))
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

func newTagRemoveCmd(deps *Deps) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "remove <screen_name> <tag> [tag...]",
		Short: "移除订阅用户的主题标签",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			screenName := args[0]
			tags := args[1:]

			if err := deps.Store.RemoveUserTags(ctx, screenName, tags); err != nil {
				return err
			}

			user, err := deps.Store.GetWatchedUser(ctx, screenName)
			if err != nil {
				return err
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(user, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("已从 @%s 移除标签：%s\n", screenName, strings.Join(tags, ", "))
			if len(user.Tags) == 0 {
				fmt.Println("当前无标签")
			} else {
				fmt.Printf("当前标签：%s\n", strings.Join(user.Tags, ", "))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

func newTagSetCmd(deps *Deps) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "set <screen_name> <tag> [tag...]",
		Short: "设置订阅用户的主题标签（覆盖原有标签）",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			screenName := args[0]
			tags := args[1:]

			if err := deps.Store.SetUserTags(ctx, screenName, tags); err != nil {
				return err
			}

			user, err := deps.Store.GetWatchedUser(ctx, screenName)
			if err != nil {
				return err
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(user, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if len(user.Tags) == 0 {
				fmt.Printf("已清空 @%s 的标签\n", screenName)
			} else {
				fmt.Printf("已将 @%s 的标签设为：%s\n", screenName, strings.Join(user.Tags, ", "))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

func newTagListCmd(deps *Deps) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有主题标签及订阅人数",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			counts, err := deps.Store.ListTags(ctx)
			if err != nil {
				return fmt.Errorf("获取标签列表失败: %w", err)
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(counts, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if len(counts) == 0 {
				fmt.Println("当前没有设置任何主题标签")
				fmt.Println("使用 xwatch tag add <screen_name> <tag> 来添加")
				return nil
			}

			tags := make([]string, 0, len(counts))
			for tag := range counts {
				tags = append(tags, tag)
			}
			sort.Strings(tags)

			fmt.Println("主题标签：")
			for _, tag := range tags {
				fmt.Printf("  %-20s %d 位用户\n", tag, counts[tag])
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}
