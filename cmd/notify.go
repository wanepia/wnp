package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/api"
	"github.com/wanepia/wnp/internal/table"
)

var notifyCmd = &cobra.Command{
	Use:     "notify",
	Aliases: []string{"n"},
	Short:   "Manage notification policies and channels",
}

var notifyPolicyCmd = &cobra.Command{
	Use:   "policy <check-id>",
	Short: "Show the notification policy for a check",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var p api.PolicyWithChannels
		if err := c.Get("/v1/checks/"+args[0]+"/policy", &p); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(p)
			return
		}
		fmt.Printf("\n%s\n\n", table.Bold("policy"))
		rows := [][]string{
			{"cooldown", fmt.Sprintf("%ds", p.Policy.CooldownSeconds)},
			{"on recovery", table.Bool(p.Policy.NotifyOnRecovery)},
			{"silenced", table.Bool(p.Policy.Silenced)},
			{"repeat interval", fmt.Sprintf("%ds", p.Policy.RepeatIntervalSeconds)},
		}
		table.Print([]string{"SETTING", "VALUE"}, rows)
		if len(p.Channels) > 0 {
			fmt.Printf("\n%s\n\n", table.Bold("channels"))
			chRows := make([][]string, 0, len(p.Channels))
			for _, ch := range p.Channels {
				chRows = append(chRows, []string{
					truncID(ch.ID, 8),
					ch.ChannelType,
					table.Bool(ch.Active),
					ch.ConfigJSON,
				})
			}
			table.Print([]string{"ID", "TYPE", "ACTIVE", "CONFIG"}, chRows)
		}
		fmt.Println()
	},
}

var (
	policyCD       int
	policyRecovery bool
	policySilence  bool
	policyRepeat   int
)

var notifySetPolicyCmd = &cobra.Command{
	Use:   "set-policy <check-id>",
	Short: "Create or update the notification policy for a check",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		body := map[string]interface{}{
			"cooldown_seconds":        policyCD,
			"notify_on_recovery":      policyRecovery,
			"silenced":                policySilence,
			"repeat_interval_seconds": policyRepeat,
		}
		var p api.NotifyPolicy
		if err := c.Post("/v1/checks/"+args[0]+"/policy", body, &p); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(p)
			return
		}
		fmt.Printf("policy set for check %s\n", args[0])
	},
}

var (
	channelType   string
	channelConfig []string
)

var notifyAddChannelCmd = &cobra.Command{
	Use:   "add-channel <check-id>",
	Short: "Add a notification channel to a check policy",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		body := map[string]interface{}{
			"channel_type": channelType,
			"config":       parseKV(channelConfig),
		}
		var ch api.NotifyChannel
		if err := c.Post("/v1/checks/"+args[0]+"/policy/channels", body, &ch); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(ch)
			return
		}
		fmt.Printf("added %s channel %s\n", ch.ChannelType, ch.ID)
	},
}

var notifyRmChannelCmd = &cobra.Command{
	Use:   "rm-channel <check-id> <channel-id>",
	Short: "Remove a notification channel",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		if err := c.Delete("/v1/checks/" + args[0] + "/policy/channels/" + args[1]); err != nil {
			die(err)
		}
		fmt.Println("channel removed")
	},
}

var notifyLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show recent notification delivery logs",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var logs []api.NotifyLog
		if err := c.Get("/v1/notify/logs", &logs); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(logs)
			return
		}
		rows := make([][]string, 0, len(logs))
		for _, l := range logs {
			statusStr := l.Status
			switch l.Status {
			case "sent":
				statusStr = table.Green(l.Status)
			case "failed":
				statusStr = table.Red(l.Status)
			case "retrying":
				statusStr = table.Yellow(l.Status)
			}
			errMsg := l.LastError
			if len(errMsg) > 50 {
				errMsg = errMsg[:50] + "…"
			}
			rows = append(rows, []string{
				truncID(l.ChannelID, 8),
				statusStr,
				fmt.Sprintf("%d", l.Attempts),
				shortTime(l.SentAt),
				errMsg,
			})
		}
		table.Print([]string{"CHANNEL", "STATUS", "ATTEMPTS", "SENT AT", "ERROR"}, rows)
		fmt.Println()
	},
}

var notifyChannelsAllCmd = &cobra.Command{
	Use:   "channels",
	Short: "List all notification channels across all checks",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var channels []api.NotifyChannel
		if err := c.Get("/v1/notify/channels", &channels); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(channels)
			return
		}
		rows := make([][]string, 0, len(channels))
		for _, ch := range channels {
			rows = append(rows, []string{
				truncID(ch.ID, 8),
				truncID(ch.PolicyID, 8),
				ch.ChannelType,
				table.Bool(ch.Active),
				ch.ConfigJSON,
			})
		}
		table.Print([]string{"ID", "POLICY", "TYPE", "ACTIVE", "CONFIG"}, rows)
		fmt.Println()
	},
}

func init() {
	notifySetPolicyCmd.Flags().IntVar(&policyCD, "cooldown", 300, "Cooldown between alerts in seconds")
	notifySetPolicyCmd.Flags().BoolVar(&policyRecovery, "recovery", true, "Notify on recovery")
	notifySetPolicyCmd.Flags().BoolVar(&policySilence, "silence", false, "Silence all alerts")
	notifySetPolicyCmd.Flags().IntVar(&policyRepeat, "repeat", 0, "Repeat alert interval in seconds (0 = no repeat)")

	notifyAddChannelCmd.Flags().StringVar(&channelType, "type", "", "Channel type: slack, discord, webhook, nats (required)")
	notifyAddChannelCmd.Flags().StringArrayVar(&channelConfig, "config", nil, "Config as key=value (repeatable, e.g. --config url=https://...)")
	_ = notifyAddChannelCmd.MarkFlagRequired("type")

	notifyCmd.AddCommand(
		notifyPolicyCmd,
		notifySetPolicyCmd,
		notifyAddChannelCmd,
		notifyRmChannelCmd,
		notifyLogsCmd,
		notifyChannelsAllCmd,
	)
}
