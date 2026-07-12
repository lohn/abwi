package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/lohn/abwi/internal/ado"
)

var listCmd = &cli.Command{
	Name:  "list",
	Usage: "List work items (defaults to yours, most recently changed first)",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "wiql", Usage: "full WIQL query (@file and @- are supported); overrides other filters"},
		&cli.StringFlag{Name: "type", Aliases: []string{"T"}, Usage: "filter by work item type"},
		&cli.StringFlag{Name: "state", Aliases: []string{"s"}, Usage: "filter by state"},
		&cli.StringFlag{Name: "assignee", Usage: "filter by assignee display name or email"},
		&cli.BoolFlag{Name: "all", Usage: "drop the assigned-to-me filter"},
		&cli.IntFlag{Name: "limit", Value: 50, Usage: "maximum number of items"},
		&cli.BoolFlag{Name: "json", Usage: "output raw JSON"},
	},
	Action: runList,
}

func runList(ctx context.Context, cmd *cli.Command) error {
	cfg, err := resolveConfig(cmd)
	if err != nil {
		return err
	}
	client, err := ado.NewClient(ctx, cfg)
	if err != nil {
		return err
	}
	query := ado.BuildWIQL(cmd.String("type"), cmd.String("state"), cmd.String("assignee"), cmd.Bool("all"))
	if w := cmd.String("wiql"); w != "" {
		query, err = expandArg(w, os.Stdin)
		if err != nil {
			return err
		}
	}
	items, err := client.Query(ctx, query, int(cmd.Int("limit")))
	if err != nil {
		return err
	}
	if cmd.Bool("json") {
		return printJSON(items)
	}
	if len(items) == 0 {
		fmt.Println("no work items found")
		return nil
	}
	fmt.Print(renderTable(items))
	return nil
}
