package commands

import (
	"encoding/json"
	"fmt"

	"github.com/mr9esx/xwatch/internal/version"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "显示 xwatch 版本信息",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := map[string]string{
				"version":         version.Version,
				"scraper_fork":    version.ScraperFork,
				"scraper_version": version.ScraperVersion,
				"auth_check":      "cookie-only (no verify_credentials)",
			}
			if jsonOutput {
				data, _ := json.MarshalIndent(info, "", "  ")
				fmt.Println(string(data))
				return nil
			}
			fmt.Printf("xwatch %s\n", version.Version)
			fmt.Printf("twitter-scraper: %s %s\n", version.ScraperFork, version.ScraperVersion)
			fmt.Println("auth check: cookie-only (不使用 verify_credentials)")
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}
