package cmd

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/config"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login <email>",
	Short: "Authenticate and save the API token to config",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			die(err)
		}
		u := cfg.URL
		if flagURL != "" {
			u = flagURL
		}
		if u == "" {
			die(fmt.Errorf("no API URL configured\n  run: wnp config set-url <url>"))
		}

		fmt.Fprintf(os.Stderr, "password: ")
		raw, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			die(fmt.Errorf("could not read password: %w", err))
		}
		password := strings.TrimSpace(string(raw))
		if password == "" {
			die(fmt.Errorf("password cannot be empty"))
		}

		c := newClientRaw(u, "")
		var result struct {
			Key    string `json:"key"`
			Prefix string `json:"prefix"`
		}
		body := map[string]string{"email": args[0], "password": password}
		if err := c.Post("/v1/auth/login", body, &result); err != nil {
			die(err)
		}

		cfg.Token = result.Key
		if err := config.Save(cfg); err != nil {
			die(err)
		}
		fmt.Printf("logged in — token saved (prefix: %s…)\n", result.Prefix)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
