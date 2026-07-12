package cli

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/lohn/abwi/internal/ado"
)

var showCmd = &cli.Command{
	Name:      "show",
	Usage:     "Show a work item",
	ArgsUsage: "<id>",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "json", Usage: "output raw JSON"},
	},
	Action: runShow,
}

func runShow(ctx context.Context, cmd *cli.Command) error {
	id, err := argID(cmd, 0)
	if err != nil {
		return err
	}
	cfg, err := resolveConfig(cmd)
	if err != nil {
		return err
	}
	client, err := ado.NewClient(ctx, cfg)
	if err != nil {
		return err
	}
	wi, err := client.Get(ctx, id, true)
	if err != nil {
		return err
	}
	if cmd.Bool("json") {
		return printJSON(wi)
	}
	htmlFields, err := client.HTMLFields(ctx)
	if err != nil {
		return err
	}
	fmt.Print(renderWorkItem(wi, htmlFields))
	return nil
}
