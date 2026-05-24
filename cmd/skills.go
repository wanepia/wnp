package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/api"
	"github.com/wanepia/wnp/internal/table"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage MCP skill manifests",
}

var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List skills",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var skills []api.Skill
		if err := c.Get("/v1/skills", &skills); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(skills)
			return
		}
		rows := make([][]string, 0, len(skills))
		for _, s := range skills {
			rows = append(rows, []string{
				s.Slug,
				s.Name,
				s.Version,
				fmt.Sprintf("%d", len(s.Tools)),
				table.Bool(s.Enabled),
			})
		}
		table.Print([]string{"SLUG", "NAME", "VERSION", "TOOLS", "ENABLED"}, rows)
		fmt.Println()
	},
}

var skillsGetCmd = &cobra.Command{
	Use:   "get <slug>",
	Short: "Get a skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var s api.Skill
		if err := c.Get("/v1/skills/"+args[0], &s); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(s)
			return
		}
		fmt.Printf("\n%s  %s  %s\n", table.Bold(s.Name), table.Gray("v"+s.Version), table.Bool(s.Enabled))
		if s.Description != "" {
			fmt.Printf("%s\n", table.Dim(s.Description))
		}
		fmt.Println()
		if len(s.Tools) > 0 {
			rows := make([][]string, 0, len(s.Tools))
			for _, t := range s.Tools {
				rows = append(rows, []string{t.Name, t.Description})
			}
			table.Print([]string{"TOOL", "DESCRIPTION"}, rows)
		}
		fmt.Println()
	},
}

var skillsEnableCmd = &cobra.Command{
	Use:   "enable <slug>",
	Short: "Enable a skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setSkillEnabled(args[0], true)
	},
}

var skillsDisableCmd = &cobra.Command{
	Use:   "disable <slug>",
	Short: "Disable a skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setSkillEnabled(args[0], false)
	},
}

func setSkillEnabled(slug string, enabled bool) {
	c, err := newClient()
	if err != nil {
		die(err)
	}
	var s api.Skill
	if err := c.Put("/v1/skills/"+slug, map[string]interface{}{"enabled": enabled}, &s); err != nil {
		die(err)
	}
	state := "enabled"
	if !enabled {
		state = "disabled"
	}
	fmt.Printf("skill %s %s\n", s.Slug, state)
}

var skillsDeleteCmd = &cobra.Command{
	Use:   "delete <slug>",
	Short: "Delete a skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		if err := c.Delete("/v1/skills/" + args[0]); err != nil {
			die(err)
		}
		fmt.Println("deleted")
	},
}

func init() {
	skillsCmd.AddCommand(skillsListCmd, skillsGetCmd, skillsEnableCmd, skillsDisableCmd, skillsDeleteCmd)
}
