package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func NewUnsubscribeCmd(deps *Deps) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "unsubscribe <screen_name>",
		Short: "取消订阅 X 用户",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			screenName := args[0]
			if err := deps.Store.RemoveWatchedUser(cmd.Context(), screenName); err != nil {
				return fmt.Errorf("取消订阅失败: %w", err)
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(map[string]interface{}{
					"success":     true,
					"screen_name": screenName,
				}, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("已取消订阅 @%s\n", screenName)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}
