package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/api"
	"github.com/wanepia/wnp/internal/table"
)

var blueprintsCmd = &cobra.Command{
	Use:     "blueprints",
	Aliases: []string{"bp"},
	Short:   "Manage blueprints",
}

var blueprintsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all blueprints",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var bps []api.Blueprint
		if err := c.Get("/v1/blueprints", &bps); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(bps)
			return
		}
		rows := make([][]string, 0, len(bps))
		for _, b := range bps {
			fieldNames := make([]string, 0, len(b.Fields))
			for _, f := range b.Fields {
				fieldNames = append(fieldNames, f.Name)
			}
			rows = append(rows, []string{
				b.Slug,
				b.Name,
				b.Description,
				strings.Join(fieldNames, ", "),
			})
		}
		table.Print([]string{"SLUG", "NAME", "DESCRIPTION", "FIELDS"}, rows)
		fmt.Println()
	},
}

var blueprintsGetCmd = &cobra.Command{
	Use:   "get <slug>",
	Short: "Get a blueprint",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var bp api.Blueprint
		if err := c.Get("/v1/blueprints/"+args[0], &bp); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(bp)
			return
		}
		fmt.Printf("\n%s  %s\n", table.Bold(bp.Name), table.Gray("("+bp.Slug+")"))
		if bp.Description != "" {
			fmt.Printf("%s\n", table.Dim(bp.Description))
		}
		fmt.Println()
		if len(bp.Fields) > 0 {
			rows := make([][]string, 0, len(bp.Fields))
			for _, f := range bp.Fields {
				req := ""
				if f.Required {
					req = table.Yellow("required")
				}
				rows = append(rows, []string{f.Name, f.FieldType, req, f.DefaultValue})
			}
			table.Print([]string{"FIELD", "TYPE", "", "DEFAULT"}, rows)
		}
		fmt.Println()
	},
}

var (
	bpSlug string
	bpName string
	bpDesc string
)

var blueprintsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new blueprint",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		body := map[string]interface{}{
			"slug":        bpSlug,
			"name":        bpName,
			"description": bpDesc,
			"fields":      []interface{}{},
		}
		var bp api.Blueprint
		if err := c.Post("/v1/blueprints", body, &bp); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(bp)
			return
		}
		fmt.Printf("created blueprint %s (%s)\n", table.Bold(bp.Name), bp.Slug)
	},
}

func init() {
	blueprintsCreateCmd.Flags().StringVar(&bpSlug, "slug", "", "Blueprint slug (required)")
	blueprintsCreateCmd.Flags().StringVar(&bpName, "name", "", "Blueprint name (required)")
	blueprintsCreateCmd.Flags().StringVar(&bpDesc, "desc", "", "Description")
	_ = blueprintsCreateCmd.MarkFlagRequired("slug")
	_ = blueprintsCreateCmd.MarkFlagRequired("name")

	blueprintsCmd.AddCommand(blueprintsListCmd, blueprintsGetCmd, blueprintsCreateCmd)
}
