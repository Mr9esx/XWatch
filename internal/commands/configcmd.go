package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewConfigCmd(deps *Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "查看或修改配置",
	}

	cmd.AddCommand(newConfigGetCmd(deps))
	cmd.AddCommand(newConfigSetCmd(deps))
	cmd.AddCommand(newConfigListCmd(deps))

	return cmd
}

func newConfigGetCmd(deps *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "获取配置值",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			val, err := deps.Config.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if val == "" {
				fmt.Printf("%s: (未设置)\n", args[0])
			} else {
				display := val
				if isSensitiveKey(args[0]) {
					display = maskValue(val)
				}
				fmt.Printf("%s: %s\n", args[0], display)
			}
			return nil
		},
	}
}

func newConfigSetCmd(deps *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "设置配置值",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := deps.Config.Set(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}
			display := args[1]
			if isSensitiveKey(args[0]) {
				display = maskValue(args[1])
			}
			fmt.Printf("已设置 %s = %s\n", args[0], display)
			return nil
		},
	}
}

func newConfigListCmd(deps *Deps) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有配置",
		RunE: func(cmd *cobra.Command, args []string) error {
			all, err := deps.Config.GetAll(cmd.Context())
			if err != nil {
				return err
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(all, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			for k, v := range all {
				display := v
				if isSensitiveKey(k) {
					display = maskValue(v)
				}
				fmt.Printf("%-25s = %s\n", k, display)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	return strings.Contains(lower, "token") ||
		strings.Contains(lower, "password") ||
		strings.Contains(lower, "api_key") ||
		strings.Contains(lower, "ct0")
}

func maskValue(v string) string {
	if len(v) <= 8 {
		return "***"
	}
	return v[:4] + "..." + v[len(v)-4:]
}
