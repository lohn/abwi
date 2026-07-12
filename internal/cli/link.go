package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/urfave/cli/v3"

	"github.com/lohn/abwi/internal/ado"
)

var linkCmd = &cli.Command{
	Name:      "link",
	Usage:     "Link two work items",
	ArgsUsage: "<id> <target-id>",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "type", Value: "related", Usage: "parent, child, related, or a System.LinkTypes.* reference"},
	},
	Action: runLink,
}

var unlinkCmd = &cli.Command{
	Name:      "unlink",
	Usage:     "Remove a link between two work items",
	ArgsUsage: "<id> <target-id>",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "type", Usage: "restrict to one link type when multiple links exist"},
	},
	Action: runUnlink,
}

func runLink(ctx context.Context, cmd *cli.Command) error {
	id, err := argID(cmd, 0)
	if err != nil {
		return err
	}
	target, err := argID(cmd, 1)
	if err != nil {
		return err
	}
	typeRef, err := ado.LinkTypeRef(cmd.String("type"))
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
	ops := []webapi.JsonPatchOperation{ado.RelationPatch(client.Org, typeRef, target)}
	if _, err := client.Update(ctx, id, ops); err != nil {
		return err
	}
	fmt.Printf("linked #%d -[%s]-> #%d\n", id, ado.RelAlias(typeRef), target)
	return nil
}

func runUnlink(ctx context.Context, cmd *cli.Command) error {
	id, err := argID(cmd, 0)
	if err != nil {
		return err
	}
	target, err := argID(cmd, 1)
	if err != nil {
		return err
	}
	typeRef := ""
	if v := cmd.String("type"); v != "" {
		typeRef, err = ado.LinkTypeRef(v)
		if err != nil {
			return err
		}
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
	if wi.Relations == nil {
		return fmt.Errorf("#%d has no relations", id)
	}
	idx := ado.FindRelationIndexes(*wi.Relations, target, typeRef)
	if len(idx) == 0 {
		return fmt.Errorf("no relation to #%d found on #%d", target, id)
	}
	if len(idx) > 1 && typeRef == "" {
		var types []string
		for _, i := range idx {
			if rel := (*wi.Relations)[i].Rel; rel != nil {
				types = append(types, ado.RelAlias(*rel))
			}
		}
		return fmt.Errorf("multiple relations to #%d (%s): disambiguate with --type", target, strings.Join(types, ", "))
	}
	if _, err := client.Update(ctx, id, ado.RemoveRelationOps(idx)); err != nil {
		return err
	}
	fmt.Printf("unlinked #%d -x-> #%d\n", id, target)
	return nil
}
