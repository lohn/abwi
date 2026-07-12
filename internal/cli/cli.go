// Package cli implements the abwi command tree.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/lohn/abwi/internal/config"
)

// Run executes the abwi CLI with the given arguments (including argv[0]).
func Run(ctx context.Context, args []string) error {
	root := &cli.Command{
		Name:  "abwi",
		Usage: "Read and write Azure Boards work items with first-class Markdown",
		Flags: globalFlags(),
		Commands: []*cli.Command{
			showCmd,
			configCmd,
		},
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

// resolveConfig loads config files and applies global flag overrides. It does
// not require org/project to be set; ado.NewClient enforces that.
func resolveConfig(cmd *cli.Command) (*config.Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(cwd)
	if err != nil {
		return nil, err
	}
	override := func(dst *string, flag, key string) {
		if v := cmd.String(flag); v != "" {
			*dst, cfg.Origins[key] = v, "flag"
		}
	}
	override(&cfg.Org, "org", "org")
	override(&cfg.Project, "project", "project")
	override(&cfg.Format, "format", "format")
	override(&cfg.Auth, "auth", "auth")
	return cfg, cfg.Validate()
}

// argID parses positional argument i as a work item ID.
func argID(cmd *cli.Command, i int) (int, error) {
	s := cmd.Args().Get(i)
	if s == "" {
		return 0, fmt.Errorf("missing work item ID argument")
	}
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid work item ID %q", s)
	}
	return id, nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
