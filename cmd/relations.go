package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/api"
	"github.com/wanepia/wnp/internal/table"
)

var relationsCmd = &cobra.Command{
	Use:     "relations",
	Aliases: []string{"rel"},
	Short:   "Manage entity dependency relations",
}

var relationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all entity relations",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var list []api.EntityRelation
		if err := c.Get("/v1/relations", &list); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(list)
			return
		}
		rows := make([][]string, 0, len(list))
		for _, r := range list {
			rows = append(rows, []string{
				truncID(r.ID, 8),
				truncID(r.FromEntityID, 8),
				r.RelationType,
				truncID(r.ToEntityID, 8),
			})
		}
		table.Print([]string{"ID", "FROM", "TYPE", "TO"}, rows)
		fmt.Println()
	},
}

var (
	relFrom string
	relTo   string
	relType string
)

var relationsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a relation between entities",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		body := map[string]interface{}{
			"from_entity_id": relFrom,
			"to_entity_id":   relTo,
			"relation_type":  relType,
		}
		var r api.EntityRelation
		if err := c.Post("/v1/relations", body, &r); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(r)
			return
		}
		fmt.Printf("added relation %s: %s %s %s\n", truncID(r.ID, 8), truncID(r.FromEntityID, 8), table.Bold(r.RelationType), truncID(r.ToEntityID, 8))
	},
}

var relationsRmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Remove a relation",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		if err := c.Delete("/v1/relations/" + args[0]); err != nil {
			die(err)
		}
		fmt.Println("relation removed")
	},
}

func init() {
	relationsAddCmd.Flags().StringVar(&relFrom, "from", "", "Source entity ID (required)")
	relationsAddCmd.Flags().StringVar(&relTo, "to", "", "Target entity ID (required)")
	relationsAddCmd.Flags().StringVar(&relType, "type", "depends_on", "Relation type: depends_on, parent_of, calls, related_to")
	_ = relationsAddCmd.MarkFlagRequired("from")
	_ = relationsAddCmd.MarkFlagRequired("to")

	relationsCmd.AddCommand(relationsListCmd, relationsAddCmd, relationsRmCmd)
}
