package commands

import (
	"fmt"

	"github.com/mr9esx/xwatch/internal/config"
	"github.com/mr9esx/xwatch/internal/store"
	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "初始化 xwatch 配置",
		Long:  "创建数据目录和数据库，引导完成基础配置。",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.EnsureDataDir(); err != nil {
				return fmt.Errorf("创建数据目录失败: %w", err)
			}

			dbPath := config.DefaultDBPath()
			s, err := store.NewSQLiteStore(dbPath)
			if err != nil {
				return fmt.Errorf("初始化数据库失败: %w", err)
			}
			defer s.Close()

			fmt.Println("xwatch 初始化完成！")
			fmt.Printf("数据库位置: %s\n", dbPath)
			fmt.Println()
			fmt.Println("接下来请配置 X (Twitter) 认证信息:")
			fmt.Println("  xwatch config set x_auth_token <your_auth_token>")
			fmt.Println("  xwatch config set x_ct0 <your_ct0_token>")
			fmt.Println()
			fmt.Println("如需代理:")
			fmt.Println("  xwatch config set proxy socks5://user:pass@host:port")
			fmt.Println()
			fmt.Println("翻译和总结能力由 Hermes Agent 提供，无需额外配置 LLM。")
			fmt.Println("安装 Hermes 插件:")
			fmt.Println("  cp -r hermes/xwatch-plugin/ ~/.hermes/plugins/xwatch/")

			return nil
		},
	}
}
