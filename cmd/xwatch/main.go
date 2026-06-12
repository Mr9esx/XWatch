package main

import (
	"fmt"
	"os"

	"github.com/mr9esx/xwatch/internal/commands"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "xwatch",
		Short: "X/Twitter 推文订阅与推送工具",
		Long:  "订阅 X (Twitter) 用户的推文，自动检测新推文并翻译推送。配合 Hermes Agent 实现对话式管理。",
	}

	var dbPath string
	root.PersistentFlags().StringVar(&dbPath, "db", "", "数据库路径 (默认 ~/.xwatch/xwatch.db)")

	deps := &commands.Deps{}
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" {
			return nil
		}
		d, err := commands.InitDeps(dbPath)
		if err != nil {
			return err
		}
		*deps = *d
		return nil
	}
	root.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if deps.Store != nil {
			deps.Store.Close()
		}
	}

	root.AddCommand(
		commands.NewVersionCmd(),
		commands.NewInitCmd(),
		commands.NewCheckCmd(deps),
		commands.NewSearchCmd(deps),
		commands.NewSubscribeCmd(deps),
		commands.NewUnsubscribeCmd(deps),
		commands.NewListCmd(deps),
		commands.NewTagCmd(deps),
		commands.NewTweetsCmd(deps),
		commands.NewConfigCmd(deps),
		commands.NewAccountCmd(deps),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
