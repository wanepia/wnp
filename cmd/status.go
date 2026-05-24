package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/api"
	"github.com/wanepia/wnp/internal/table"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show overall fleet status",
}

var statusOverviewCmd = &cobra.Command{
	Use:   "show",
	Short: "Fleet health summary",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var status api.StatusResponse
		if err := c.Get("/v1/status", &status); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(status)
			return
		}

		fmt.Printf("\n  %s  %s  %s\n\n",
			table.Green(fmt.Sprintf("%d up", status.Up)),
			table.Yellow(fmt.Sprintf("%d degraded", status.Degraded)),
			table.Red(fmt.Sprintf("%d down", status.Down)),
		)

		rows := make([][]string, 0, len(status.Entities))
		for _, e := range status.Entities {
			rows = append(rows, []string{
				e.Name,
				e.BlueprintSlug,
				e.Slug,
				table.StatusColor(e.CurrentStatus),
			})
		}
		table.Print([]string{"NAME", "BLUEPRINT", "SLUG", "STATUS"}, rows)
		fmt.Println()
	},
}

var statusTransitionsCmd = &cobra.Command{
	Use:   "transitions",
	Short: "Recent state-change events across all entities",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var list []api.StateTransitionWithEntity
		if err := c.Get("/v1/status/transitions", &list); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(list)
			return
		}
		rows := make([][]string, 0, len(list))
		for _, t := range list {
			arrow := table.Gray(t.FromState) + " → " + table.StatusColor(t.ToState)
			rows = append(rows, []string{
				t.EntityName,
				arrow,
				t.TriggerReason,
				shortTime(t.TransitionedAt),
			})
		}
		table.Print([]string{"ENTITY", "TRANSITION", "REASON", "WHEN"}, rows)
		fmt.Println()
	},
}

func init() {
	statusCmd.AddCommand(statusOverviewCmd, statusTransitionsCmd)
	statusCmd.Run = statusOverviewCmd.Run
}

func shortTime(s string) string {
	if len(s) >= 16 {
		return strings.Replace(s[:16], "T", " ", 1)
	}
	return s
}
