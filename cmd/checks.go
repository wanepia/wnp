package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/api"
	"github.com/wanepia/wnp/internal/table"
)

func checkTarget(ch api.Check) string {
	if ch.TargetURL != "" {
		return ch.TargetURL
	}
	switch ch.CheckType {
	case "dns":
		hostname, _ := ch.Config["hostname"].(string)
		recType, _ := ch.Config["record_type"].(string)
		if hostname != "" {
			if recType != "" {
				return hostname + " (" + recType + ")"
			}
			return hostname
		}
	case "tcp", "tls":
		host, _ := ch.Config["host"].(string)
		port, _ := ch.Config["port"].(float64)
		if host != "" {
			if port != 0 {
				return fmt.Sprintf("%s:%d", host, int(port))
			}
			return host
		}
	}
	return table.Dim("—")
}

var checksCmd = &cobra.Command{
	Use:     "checks",
	Aliases: []string{"chk"},
	Short:   "Manage health checks",
}

var chkEntityFilter string

var checksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all checks",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var checks []api.Check
		if err := c.Get("/v1/checks", &checks); err != nil {
			die(err)
		}
		if chkEntityFilter != "" {
			filter := strings.ToLower(chkEntityFilter)
			filtered := checks[:0]
			for _, ch := range checks {
				if strings.HasPrefix(strings.ToLower(ch.EntityID), filter) {
					filtered = append(filtered, ch)
				}
			}
			checks = filtered
		}
		if flagJSON {
			printJSON(checks)
			return
		}
		rows := make([][]string, 0, len(checks))
		for _, ch := range checks {
			rows = append(rows, []string{
				truncID(ch.ID, 8),
				ch.CheckType,
				checkTarget(ch),
				fmt.Sprintf("%ds", ch.IntervalSeconds),
				fmt.Sprintf("%d", ch.FailureThreshold),
				table.Bool(ch.Enabled),
			})
		}
		table.Print([]string{"ID", "TYPE", "TARGET", "INTERVAL", "THRESHOLD", "ON"}, rows)
		fmt.Println(table.Dim("  tip: IDs are abbreviated — use --json for full UUIDs, or pass any unambiguous prefix"))
		fmt.Println()
	},
}

var checksGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a check",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		id, err := resolveCheckID(c, args[0])
		if err != nil {
			die(err)
		}
		var ch api.Check
		if err := c.Get("/v1/checks/"+id, &ch); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(ch)
			return
		}
		fmt.Printf("\n%s check  %s\n\n", table.Bold(ch.CheckType), table.Dim(ch.ID))
		rows := [][]string{
			{"target", checkTarget(ch)},
			{"interval", fmt.Sprintf("%ds", ch.IntervalSeconds)},
			{"timeout", fmt.Sprintf("%dms", ch.TimeoutMs)},
			{"failure threshold", fmt.Sprintf("%d", ch.FailureThreshold)},
			{"enabled", table.Bool(ch.Enabled)},
		}
		if ch.ExpectedStatus != 0 {
			rows = append(rows, []string{"expected status", strconv.Itoa(ch.ExpectedStatus)})
		}
		if ch.BodyContains != "" {
			rows = append(rows, []string{"body contains", ch.BodyContains})
		}
		table.Print([]string{"FIELD", "VALUE"}, rows)
		fmt.Println()
	},
}

var (
	chkEntityID  string
	chkType      string
	chkURL       string
	chkName      string
	chkInterval  int
	chkTimeout   int
	chkStatus    int
	chkBody      string
	chkThreshold int
)

var checksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a health check",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		isPush := chkType == "push"
		if !isPush && chkURL == "" {
			die(fmt.Errorf("--url is required for %s checks", chkType))
		}
		c, err := newClient()
		if err != nil {
			die(err)
		}
		body := map[string]interface{}{
			"entity_id":         chkEntityID,
			"check_type":        chkType,
			"interval_seconds":  chkInterval,
			"timeout_ms":        chkTimeout,
			"failure_threshold": chkThreshold,
			"enabled":           true,
		}
		if isPush {
			body["execution_mode"] = "push"
		} else {
			body["target_url"] = chkURL
		}
		if chkName != "" {
			body["name"] = chkName
		}
		if chkStatus != 0 {
			body["expected_status"] = chkStatus
		}
		if chkBody != "" {
			body["body_contains"] = chkBody
		}
		var ch api.Check
		if err := c.Post("/v1/checks", body, &ch); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(ch)
			return
		}
		fmt.Printf("created %s check %s\n", ch.CheckType, table.Bold(ch.ID))
	},
}

var checksUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a check",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		id, err := resolveCheckID(c, args[0])
		if err != nil {
			die(err)
		}
		body := map[string]interface{}{}
		if cmd.Flags().Changed("url") {
			body["target_url"] = chkURL
		}
		if cmd.Flags().Changed("interval") {
			body["interval_seconds"] = chkInterval
		}
		if cmd.Flags().Changed("timeout") {
			body["timeout_ms"] = chkTimeout
		}
		if cmd.Flags().Changed("status") {
			body["expected_status"] = chkStatus
		}
		if cmd.Flags().Changed("body") {
			body["body_contains"] = chkBody
		}
		if cmd.Flags().Changed("threshold") {
			body["failure_threshold"] = chkThreshold
		}
		if len(body) == 0 {
			die(fmt.Errorf("no fields specified; use --url, --interval, --timeout, --status, --body, or --threshold"))
		}
		var ch api.Check
		if err := c.Put("/v1/checks/"+id, body, &ch); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(ch)
			return
		}
		fmt.Printf("updated check %s\n", ch.ID)
	},
}

var checksEnableCmd = &cobra.Command{
	Use:   "enable <id>",
	Short: "Enable a check",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setEnabled(args[0], true)
	},
}

var checksDisableCmd = &cobra.Command{
	Use:   "disable <id>",
	Short: "Disable a check (stops polling, keeps config)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setEnabled(args[0], false)
	},
}

func setEnabled(id string, enabled bool) {
	c, err := newClient()
	if err != nil {
		die(err)
	}
	fullID, err := resolveCheckID(c, id)
	if err != nil {
		die(err)
	}
	var ch api.Check
	if err := c.Put("/v1/checks/"+fullID, map[string]interface{}{"enabled": enabled}, &ch); err != nil {
		die(err)
	}
	state := "enabled"
	if !enabled {
		state = "disabled"
	}
	fmt.Printf("check %s %s\n", ch.ID, state)
}

var checksDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a check",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		id, err := resolveCheckID(c, args[0])
		if err != nil {
			die(err)
		}
		if err := c.Delete("/v1/checks/" + id); err != nil {
			die(err)
		}
		fmt.Println("deleted")
	},
}

var chkResultsLimit int

var checksResultsCmd = &cobra.Command{
	Use:   "results <id>",
	Short: "Show recent check results",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		id, err := resolveCheckID(c, args[0])
		if err != nil {
			die(err)
		}
		var resp api.CheckResultsResponse
		if err := c.Get(fmt.Sprintf("/v1/checks/%s/results?limit=%d", id, chkResultsLimit), &resp); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(resp)
			return
		}
		rows := make([][]string, 0, len(resp.Results))
		for _, r := range resp.Results {
			ok := table.Green("✓")
			if !r.Success {
				ok = table.Red("✗")
			}
			status := ""
			if r.StatusCode != 0 {
				status = strconv.Itoa(r.StatusCode)
			}
			msg := r.ErrorMessage
			if len(msg) > 60 {
				msg = msg[:60] + "…"
			}
			rows = append(rows, []string{
				ok,
				status,
				fmt.Sprintf("%dms", r.LatencyMs),
				shortTime(r.CheckedAt),
				msg,
			})
		}
		table.Print([]string{"", "STATUS", "LATENCY", "CHECKED AT", "ERROR"}, rows)
		fmt.Println()
	},
}

var checksTransitionsCmd = &cobra.Command{
	Use:   "transitions <id>",
	Short: "Show state transitions for a check",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		id, err := resolveCheckID(c, args[0])
		if err != nil {
			die(err)
		}
		var list []api.StateTransition
		if err := c.Get("/v1/checks/"+id+"/transitions", &list); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(list)
			return
		}
		rows := make([][]string, 0, len(list))
		for _, t := range list {
			rows = append(rows, []string{
				table.Gray(t.FromState) + " → " + table.StatusColor(t.ToState),
				t.TriggerReason,
				shortTime(t.TransitionedAt),
			})
		}
		table.Print([]string{"TRANSITION", "REASON", "WHEN"}, rows)
		fmt.Println()
	},
}

var (
	alertType   string
	alertConfig []string
)

var checksAlertCmd = &cobra.Command{
	Use:   "alert <check-id>",
	Short: "Add an alert channel to a check (creates policy if needed)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		id, err := resolveCheckID(c, args[0])
		if err != nil {
			die(err)
		}
		policyBody := map[string]interface{}{
			"cooldown_seconds":        300,
			"notify_on_recovery":      true,
			"silenced":                false,
			"repeat_interval_seconds": 0,
		}
		var policy api.NotifyPolicy
		if err := c.Post("/v1/checks/"+id+"/policy", policyBody, &policy); err != nil {
			die(err)
		}
		channelBody := map[string]interface{}{
			"channel_type": alertType,
			"config":       parseKV(alertConfig),
		}
		var ch api.NotifyChannel
		if err := c.Post("/v1/checks/"+id+"/policy/channels", channelBody, &ch); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(ch)
			return
		}
		fmt.Printf("alert added: %s channel %s on check %s\n", ch.ChannelType, truncID(ch.ID, 8), truncID(id, 8))
	},
}

func init() {
	checksCreateCmd.Flags().StringVar(&chkEntityID, "entity", "", "Entity ID (required)")
	checksCreateCmd.Flags().StringVar(&chkType, "type", "http", "Check type: http, tcp, tls, dns, push")
	checksCreateCmd.Flags().StringVar(&chkURL, "url", "", "Target URL or host:port (required for non-push checks)")
	checksCreateCmd.Flags().StringVar(&chkName, "name", "", "Display name for the check")
	checksCreateCmd.Flags().IntVar(&chkInterval, "interval", 60, "Poll interval in seconds")
	checksCreateCmd.Flags().IntVar(&chkTimeout, "timeout", 5000, "Timeout in milliseconds")
	checksCreateCmd.Flags().IntVar(&chkStatus, "status", 0, "Expected HTTP status code")
	checksCreateCmd.Flags().StringVar(&chkBody, "body", "", "Required body substring (HTTP only)")
	checksCreateCmd.Flags().IntVar(&chkThreshold, "threshold", 3, "Consecutive failures before state change")
	_ = checksCreateCmd.MarkFlagRequired("entity")

	checksUpdateCmd.Flags().StringVar(&chkURL, "url", "", "New target URL")
	checksUpdateCmd.Flags().IntVar(&chkInterval, "interval", 0, "New interval in seconds")
	checksUpdateCmd.Flags().IntVar(&chkTimeout, "timeout", 0, "New timeout in milliseconds")
	checksUpdateCmd.Flags().IntVar(&chkStatus, "status", 0, "New expected HTTP status")
	checksUpdateCmd.Flags().StringVar(&chkBody, "body", "", "New body substring")
	checksUpdateCmd.Flags().IntVar(&chkThreshold, "threshold", 0, "New failure threshold")

	checksListCmd.Flags().StringVar(&chkEntityFilter, "entity", "", "Filter by entity ID prefix")

	checksResultsCmd.Flags().IntVar(&chkResultsLimit, "limit", 50, "Number of results to fetch")

	checksAlertCmd.Flags().StringVar(&alertType, "type", "", "Channel type: slack, discord, webhook, nats (required)")
	checksAlertCmd.Flags().StringArrayVar(&alertConfig, "config", nil, "Config as key=value (e.g. --config url=https://...)")
	_ = checksAlertCmd.MarkFlagRequired("type")

	checksCmd.AddCommand(
		checksListCmd,
		checksGetCmd,
		checksCreateCmd,
		checksUpdateCmd,
		checksEnableCmd,
		checksDisableCmd,
		checksDeleteCmd,
		checksResultsCmd,
		checksTransitionsCmd,
		checksAlertCmd,
	)
}
