package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/api"
	"github.com/wanepia/wnp/internal/config"
)

// truncID safely shortens a UUID-like ID to n chars for display.
func truncID(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// resolveCheckID resolves a full or prefix UUID to a full check ID.
func resolveCheckID(c interface {
	Get(string, interface{}) error
}, id string) (string, error) {
	if len(id) == 36 {
		return id, nil
	}
	type checkList []struct {
		ID string `json:"ID"`
	}
	var list checkList
	if err := c.Get("/v1/checks", &list); err != nil {
		return "", err
	}
	var matches []string
	for _, ch := range list {
		if strings.HasPrefix(ch.ID, id) {
			matches = append(matches, ch.ID)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no check found with prefix %q", id)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("prefix %q is ambiguous (%d matches) — use more characters", id, len(matches))
	}
}

var Version = "dev"

var (
	flagURL   string
	flagToken string
	flagJSON  bool
)

var rootCmd = &cobra.Command{
	Use:     "wnp",
	Short:   "Wanepia CLI — manage your service catalog and monitoring",
	Long:    "wnp lets you manage blueprints, entities, health checks, alerts, and team from the terminal.\n\nRun `wnp config set-url <url>` and `wnp config set-token <token>` to get started.",
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newClient() (*api.Client, error) {
	cfg, _ := config.Load()
	u := flagURL
	if u == "" && cfg != nil {
		u = cfg.URL
	}
	t := flagToken
	if t == "" && cfg != nil {
		t = cfg.Token
	}
	if u == "" {
		return nil, fmt.Errorf("no API URL configured\n  run: wnp config set-url <url>")
	}
	if t == "" {
		fmt.Fprintln(os.Stderr, "warning: no API token configured — run: wnp config set-token <token>")
	}
	return api.New(u, t), nil
}

func printJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func newClientRaw(url, token string) *api.Client {
	return api.New(url, token)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagURL, "url", "", "API base URL (overrides config)")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "API key (overrides config)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output raw JSON")

	rootCmd.AddCommand(
		statusCmd,
		blueprintsCmd,
		entitiesCmd,
		checksCmd,
		notifyCmd,
		relationsCmd,
		teamCmd,
		keysCmd,
		skillsCmd,
		configCmd,
	)
}
