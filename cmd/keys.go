package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/api"
	"github.com/wanepia/wnp/internal/table"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage API keys",
}

var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var keys []api.APIKey
		if err := c.Get("/v1/auth/keys", &keys); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(keys)
			return
		}
		rows := make([][]string, 0, len(keys))
		for _, k := range keys {
			rows = append(rows, []string{
				k.ID,
				k.Label,
				k.Prefix + "…",
				table.Bool(k.Active),
				shortTime(k.LastUsedAt),
				shortTime(k.CreatedAt),
			})
		}
		table.Print([]string{"ID", "LABEL", "PREFIX", "ACTIVE", "LAST USED", "CREATED"}, rows)
		fmt.Println()
	},
}

var keysCreateCmd = &cobra.Command{
	Use:   "create <label>",
	Short: "Create a new API key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var result struct {
			Key    string `json:"key"`
			ID     string `json:"id"`
			Prefix string `json:"prefix"`
		}
		if err := c.Post("/v1/auth/keys", map[string]string{"label": args[0]}, &result); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(result)
			return
		}
		fmt.Printf("key created — copy it now, it won't be shown again:\n\n  %s\n\n", table.Bold(result.Key))
	},
}

var keysDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an API key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		if err := c.Delete("/v1/auth/keys/" + args[0]); err != nil {
			die(err)
		}
		fmt.Println("key deleted")
	},
}

func init() {
	keysCmd.AddCommand(keysListCmd, keysCreateCmd, keysDeleteCmd)
}
