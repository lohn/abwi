package cli

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var configCmd = &cli.Command{
	Name:      "config",
	Usage:     "Show the resolved configuration",
	ArgsUsage: "[<key>]",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "json", Usage: "output as JSON"},
	},
	Action: runConfig,
}

func runConfig(ctx context.Context, cmd *cli.Command) error {
	cfg, err := resolveConfig(cmd)
	if err != nil {
		return err
	}
	if key := cmd.Args().First(); key != "" {
		for _, it := range cfg.Items() {
			if it.Key == key {
				fmt.Println(it.Value)
				return nil
			}
		}
		if ref, ok := cfg.Aliases[key]; ok {
			fmt.Println(ref)
			return nil
		}
		return fmt.Errorf("unknown config key %q", key)
	}
	if cmd.Bool("json") {
		return printJSON(struct {
			Org         string            `json:"org"`
			Project     string            `json:"project"`
			Format      string            `json:"format"`
			Auth        string            `json:"auth"`
			DefaultType string            `json:"default-type"`
			Aliases     map[string]string `json:"aliases"`
			Origins     map[string]string `json:"origins"`
		}{cfg.Org, cfg.Project, cfg.Format, cfg.Auth, cfg.DefaultType, cfg.Aliases, cfg.Origins})
	}
	fmt.Print(renderConfig(cfg))
	return nil
}
