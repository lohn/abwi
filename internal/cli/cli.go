// Package cli implements the abwi command tree.
package cli

import (
	"context"

	"github.com/urfave/cli/v3"
)

// Run executes the abwi CLI with the given arguments (including argv[0]).
func Run(ctx context.Context, args []string) error {
	root := &cli.Command{
		Name:  "abwi",
		Usage: "Read and write Azure Boards work items with first-class Markdown",
		Flags: globalFlags(),
	}
	return root.Run(ctx, args)
}

// globalFlags are accepted before any subcommand and override config values.
func globalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "org", Usage: "organization URL (https://dev.azure.com/<org>)"},
		&cli.StringFlag{Name: "project", Usage: "project name"},
		&cli.StringFlag{Name: "format", Usage: "large text format: markdown or html"},
		&cli.StringFlag{Name: "auth", Usage: "authentication: entra or pat"},
	}
}
