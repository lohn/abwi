package cli

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"

	"github.com/lohn/abwi/internal/ado"
)

var updateCmd = &cli.Command{
	Name:      "update",
	Usage:     "Update fields of a work item",
	ArgsUsage: "<id>",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "title", Aliases: []string{"t"}, Usage: "new title"},
		&cli.StringFlag{Name: "state", Aliases: []string{"s"}, Usage: "new state (e.g. Active, Done)"},
		&cli.StringFlag{Name: "description", Aliases: []string{"d"}, Usage: "System.Description (@file and @- are supported)"},
		&cli.StringSliceFlag{Name: "field", Aliases: []string{"f"}, Usage: "field as <refname>=<value> (@file and @- are supported)"},
	},
	Action: runUpdate,
}

func runUpdate(ctx context.Context, cmd *cli.Command) error {
	id, err := argID(cmd, 0)
	if err != nil {
		return err
	}
	cfg, err := resolveConfig(cmd)
	if err != nil {
		return err
	}
	fields, err := collectFields(cmd, cfg.Aliases)
	if err != nil {
		return err
	}
	if v := cmd.String("state"); v != "" {
		fields["System.State"] = v
	}
	if len(fields) == 0 {
		return errors.New("nothing to update: pass --title, --state, --description, or --field")
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
	wi, err := client.Update(ctx, id, ops)
	if err != nil {
		return err
	}
	printWorkItemLine(client, wi)
	return nil
}
