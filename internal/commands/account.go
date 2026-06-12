package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mr9esx/xwatch/internal/store"
	"github.com/spf13/cobra"
)

func NewAccountCmd(deps *Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "管理账号池",
	}
	cmd.AddCommand(newAccountAddCmd(deps))
	cmd.AddCommand(newAccountRemoveCmd(deps))
	cmd.AddCommand(newAccountListCmd(deps))
	return cmd
}

func newAccountAddCmd(deps *Deps) *cobra.Command {
	var (
		authToken   string
		csrfToken   string
		cfClearance string
		jsonOutput  bool
	)
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "添加账号到账号池",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return fmt.Errorf("name 不能为空")
			}
			if authToken == "" || csrfToken == "" {
				return fmt.Errorf("--auth-token 和 --csrf-token 必须提供")
			}

			acct := store.Account{
				Name:        name,
				AuthToken:   authToken,
				CSRFToken:   csrfToken,
				CFClearance: cfClearance,
			}

			ctx := cmd.Context()
			if err := deps.Store.AddAccount(ctx, acct); err != nil {
				return fmt.Errorf("添加账号失败: %w", err)
			}

			if jsonOutput {
				data, _ := json.MarshalIndent(map[string]string{"status": "ok", "name": name}, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("✅ 账号 %q 已添加到池中\n", name)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&authToken, "auth-token", "", "x_auth_token (必须)")
	cmd.Flags().StringVar(&csrfToken, "csrf-token", "", "x_ct0 (必须)")
	cmd.Flags().StringVar(&cfClearance, "cf-clearance", "", "cf_clearance (可选)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

func newAccountRemoveCmd(deps *Deps) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "从账号池移除账号",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ctx := cmd.Context()
			if err := deps.Store.RemoveAccount(ctx, name); err != nil {
				return err
			}
			if jsonOutput {
				data, _ := json.MarshalIndent(map[string]string{"status": "ok", "removed": name}, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("✅ 账号 %q 已移除\n", name)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

func newAccountListCmd(deps *Deps) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出账号池中的所有账号",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			accounts, err := deps.Store.GetAccounts(ctx)
			if err != nil {
				return err
			}
			if len(accounts) == 0 {
				if jsonOutput {
					fmt.Println("[]")
				} else {
					fmt.Println("账号池为空，使用 xwatch account add 添加账号")
				}
				return nil
			}

			if jsonOutput {
				type entry struct {
					Name          string `json:"name"`
					Status        string `json:"status"`
					CooldownUntil string `json:"cooldown_until,omitempty"`
					TokenPreview  string `json:"token_preview"`
					CreatedAt     string `json:"created_at"`
					LastUsedAt    string `json:"last_used_at"`
				}
				var entries []entry
				for _, a := range accounts {
					e := entry{
						Name:         a.Name,
						Status:       a.Status,
						TokenPreview: maskToken(a.AuthToken),
						CreatedAt:    a.CreatedAt.Local().Format("2006-01-02 15:04"),
						LastUsedAt:   a.LastUsedAt.Local().Format("2006-01-02 15:04"),
					}
					if a.Status == "rate_limited" || a.Status == "disabled" {
						e.CooldownUntil = a.CooldownUntil.Local().Format("2006-01-02 15:04:05")
					}
					entries = append(entries, e)
				}
				data, _ := json.MarshalIndent(entries, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("账号池 (%d 个账号):\n\n", len(accounts))
			for _, a := range accounts {
				statusIcon := "🟢"
				switch a.Status {
				case "rate_limited":
					statusIcon = "🟡"
				case "disabled":
					statusIcon = "🔴"
				}
				fmt.Printf("  %s %s  token: %s  状态: %s\n",
					statusIcon, a.Name, maskToken(a.AuthToken), a.Status)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
