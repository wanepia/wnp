package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/api"
	"github.com/wanepia/wnp/internal/table"
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Manage team members and invitations",
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List team members and pending invitations",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var resp api.TeamResponse
		if err := c.Get("/v1/settings/users", &resp); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(resp)
			return
		}
		if len(resp.Users) > 0 {
			fmt.Printf("\n%s\n\n", table.Bold("members"))
			rows := make([][]string, 0, len(resp.Users))
			for _, u := range resp.Users {
				verified := table.Bool(u.EmailVerified)
				rows = append(rows, []string{u.Name, u.Email, u.Role, verified})
			}
			table.Print([]string{"NAME", "EMAIL", "ROLE", "VERIFIED"}, rows)
		}
		if len(resp.Invitations) > 0 {
			fmt.Printf("\n%s\n\n", table.Bold("pending invitations"))
			rows := make([][]string, 0, len(resp.Invitations))
			for _, inv := range resp.Invitations {
				rows = append(rows, []string{inv.Email, inv.Role, shortTime(inv.ExpiresAt)})
			}
			table.Print([]string{"EMAIL", "ROLE", "EXPIRES"}, rows)
		}
		fmt.Println()
	},
}

var (
	inviteRole string
)

var teamInviteCmd = &cobra.Command{
	Use:   "invite <email>",
	Short: "Invite a user to the team",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		body := map[string]interface{}{
			"email": args[0],
			"role":  inviteRole,
		}
		var result map[string]string
		if err := c.Post("/v1/settings/users/invite", body, &result); err != nil {
			die(err)
		}
		fmt.Printf("invitation sent to %s as %s\n", args[0], inviteRole)
	},
}

var teamRemoveCmd = &cobra.Command{
	Use:   "remove <user-id>",
	Short: "Remove a team member",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		if err := c.Delete("/v1/settings/users/" + args[0]); err != nil {
			die(err)
		}
		fmt.Println("member removed")
	},
}

func init() {
	teamInviteCmd.Flags().StringVar(&inviteRole, "role", "member", "Role: admin or member")
	teamCmd.AddCommand(teamListCmd, teamInviteCmd, teamRemoveCmd)
}
