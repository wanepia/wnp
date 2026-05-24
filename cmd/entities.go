package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wanepia/wnp/internal/api"
	"github.com/wanepia/wnp/internal/table"
)

var entitiesCmd = &cobra.Command{
	Use:     "entities",
	Aliases: []string{"ent", "e"},
	Short:   "Manage catalog entities",
}

var entitiesListCmd = &cobra.Command{
	Use:   "list <blueprint>",
	Short: "List entities for a blueprint",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var entities []api.Entity
		if err := c.Get("/v1/blueprints/"+args[0]+"/entities", &entities); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(entities)
			return
		}
		rows := make([][]string, 0, len(entities))
		for _, e := range entities {
			rows = append(rows, []string{
				e.Slug,
				e.Name,
				table.StatusColor(e.CurrentStatus),
				shortTime(e.CreatedAt),
			})
		}
		table.Print([]string{"SLUG", "NAME", "STATUS", "CREATED"}, rows)
		fmt.Println()
	},
}

var entitiesGetCmd = &cobra.Command{
	Use:   "get <blueprint> <slug>",
	Short: "Get an entity",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		var e api.Entity
		if err := c.Get("/v1/blueprints/"+args[0]+"/entities/"+args[1], &e); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(e)
			return
		}
		fmt.Printf("\n%s  %s  %s\n", table.Bold(e.Name), table.Gray("("+e.Slug+")"), table.StatusColor(e.CurrentStatus))
		fmt.Printf("%s\n\n", table.Dim("changed: "+shortTime(e.StatusChangedAt)))
		if len(e.Fields) > 0 {
			rows := make([][]string, 0, len(e.Fields))
			for _, f := range e.Fields {
				rows = append(rows, []string{f.FieldDefID, f.Value})
			}
			table.Print([]string{"FIELD", "VALUE"}, rows)
		}
		fmt.Println()
	},
}

var (
	entSlug   string
	entName   string
	entFields []string
)

var entitiesCreateCmd = &cobra.Command{
	Use:   "create <blueprint>",
	Short: "Create an entity",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		fields := parseKV(entFields)
		body := map[string]interface{}{
			"slug":   entSlug,
			"name":   entName,
			"fields": fields,
		}
		var e api.Entity
		if err := c.Post("/v1/blueprints/"+args[0]+"/entities", body, &e); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(e)
			return
		}
		fmt.Printf("created entity %s (%s)\n", table.Bold(e.Name), e.Slug)
	},
}

var entitiesUpdateCmd = &cobra.Command{
	Use:   "update <blueprint> <slug>",
	Short: "Update an entity",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		body := map[string]interface{}{}
		if entName != "" {
			body["name"] = entName
		}
		if len(entFields) > 0 {
			body["fields"] = parseKV(entFields)
		}
		if len(body) == 0 {
			die(fmt.Errorf("no fields specified; use --name or --field key=value"))
		}
		var e api.Entity
		if err := c.Put("/v1/blueprints/"+args[0]+"/entities/"+args[1], body, &e); err != nil {
			die(err)
		}
		if flagJSON {
			printJSON(e)
			return
		}
		fmt.Printf("updated entity %s\n", table.Bold(e.Name))
	},
}

var entitiesDeleteCmd = &cobra.Command{
	Use:   "delete <blueprint> <slug>",
	Short: "Delete an entity",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient()
		if err != nil {
			die(err)
		}
		if err := c.Delete("/v1/blueprints/" + args[0] + "/entities/" + args[1]); err != nil {
			die(err)
		}
		fmt.Println("deleted")
	},
}

func parseKV(pairs []string) map[string]string {
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

func init() {
	entitiesCreateCmd.Flags().StringVar(&entSlug, "slug", "", "Entity slug (required)")
	entitiesCreateCmd.Flags().StringVar(&entName, "name", "", "Entity name (required)")
	entitiesCreateCmd.Flags().StringArrayVar(&entFields, "field", nil, "Field value as key=value (repeatable)")
	_ = entitiesCreateCmd.MarkFlagRequired("slug")
	_ = entitiesCreateCmd.MarkFlagRequired("name")

	entitiesUpdateCmd.Flags().StringVar(&entName, "name", "", "New name")
	entitiesUpdateCmd.Flags().StringArrayVar(&entFields, "field", nil, "Field value as key=value (repeatable)")

	entitiesCmd.AddCommand(entitiesListCmd, entitiesGetCmd, entitiesCreateCmd, entitiesUpdateCmd, entitiesDeleteCmd)
}
