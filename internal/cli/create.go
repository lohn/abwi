package cli

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/lohn/abwi/internal/ado"
)

var createCmd = &cli.Command{
	Name:  "create",
	Usage: "Create a work item",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "type", Aliases: []string{"T"}, Usage: "work item type (falls back to default-type in config)"},
		&cli.StringFlag{Name: "title", Aliases: []string{"t"}, Required: true, Usage: "work item title"},
		&cli.StringFlag{Name: "description", Aliases: []string{"d"}, Usage: "System.Description (@file and @- are supported)"},
		&cli.StringSliceFlag{Name: "field", Aliases: []string{"f"}, Usage: "field as <refname>=<value> (@file and @- are supported)"},
		&cli.IntFlag{Name: "parent", Usage: "parent work item ID to link"},
	},
	Action: runCreate,
}

// collectFields merges --field, --title, and --description into one field map.
func collectFields(cmd *cli.Command, aliases map[string]string) (map[string]string, error) {
	fields, err := parseFields(cmd.StringSlice("field"), aliases, os.Stdin)
	if err != nil {
		return nil, err
	}
	if v := cmd.String("title"); v != "" {
		fields["System.Title"] = v
	}
	if v := cmd.String("description"); v != "" {
		d, err := expandArg(v, os.Stdin)
		if err != nil {
			return nil, err
		}
		fields["System.Description"] = d
	}
	return fields, nil
}

func runCreate(ctx context.Context, cmd *cli.Command) error {
	cfg, err := resolveConfig(cmd)
	if err != nil {
		return err
	}
	workItemType := cmd.String("type")
	if workItemType == "" {
		workItemType = cfg.DefaultType
	}
	if workItemType == "" {
		return errors.New("--type is required (or set default-type in config)")
	}
	fields, err := collectFields(cmd, cfg.Aliases)
	if err != nil {
		return err
	}
	client, err := ado.NewClient(ctx, cfg)
	if err != nil {
		return err
	}
	htmlFields, err := client.HTMLFields(ctx)
	if err != nil {
		return err
	}
	ops, err := ado.BuildFieldPatch(fields, htmlFields, cfg.Format)
	if err != nil {
		return err
	}
	if parent := int(cmd.Int("parent")); parent != 0 {
		parentRef, _ := ado.LinkTypeRef("parent")
		ops = append(ops, ado.RelationPatch(client.Org, parentRef, parent))
	}
	wi, err := client.Create(ctx, workItemType, ops)
	if err != nil {
		return err
	}
	printWorkItemLine(client, wi)
	return nil
}
