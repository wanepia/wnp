package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/config"
	"github.com/wanepia/wnp/internal/table"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			die(err)
		}
		token := cfg.Token
		if len(token) > 8 {
			token = token[:8] + "…"
		}
		fmt.Printf("\nconfig: %s\n\n", table.Dim(config.Path()))
		rows := [][]string{
			{"url", cfg.URL},
			{"token", token},
		}
		table.Print([]string{"KEY", "VALUE"}, rows)
		fmt.Println()
	},
}

var configSetURLCmd = &cobra.Command{
	Use:   "set-url <url>",
	Short: "Set the API base URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			die(err)
		}
		u := args[0]
		if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
			u = "https://" + u
		}
		cfg.URL = strings.TrimRight(u, "/")
		if err := config.Save(cfg); err != nil {
			die(err)
		}
		fmt.Printf("url set to %s\n", cfg.URL)
	},
}

var configSetTokenCmd = &cobra.Command{
	Use:   "set-token <token>",
	Short: "Set the API key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			die(err)
		}
		cfg.Token = args[0]
		if err := config.Save(cfg); err != nil {
			die(err)
		}
		n := len(args[0])
		if n > 8 {
			n = 8
		}
		fmt.Printf("token saved (prefix: %s…)\n", args[0][:n])
	},
}

func init() {
	configCmd.AddCommand(configShowCmd, configSetURLCmd, configSetTokenCmd)
}
