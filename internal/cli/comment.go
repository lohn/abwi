package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/lohn/abwi/internal/ado"
)

var commentCmd = &cli.Command{
	Name:  "comment",
	Usage: "Read and write work item comments",
	Commands: []*cli.Command{
		{
			Name:      "add",
			Usage:     "Add a comment (Markdown)",
			ArgsUsage: "<id> <text|@file|@->",
			Action:    runCommentAdd,
		},
		{
			Name:      "list",
			Usage:     "List comments",
			ArgsUsage: "<id>",
			Flags:     []cli.Flag{&cli.BoolFlag{Name: "json", Usage: "output raw JSON"}},
			Action:    runCommentList,
		},
	},
}

func runCommentAdd(ctx context.Context, cmd *cli.Command) error {
	id, err := argID(cmd, 0)
	if err != nil {
		return err
	}
	raw := cmd.Args().Get(1)
	if raw == "" {
		return errors.New("missing comment text argument")
	}
	text, err := expandArg(raw, os.Stdin)
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
	comment, err := client.AddComment(ctx, id, text)
	if err != nil {
		return err
	}
	fmt.Printf("comment %d added to #%d\n", comment.ID, id)
	return nil
}

func runCommentList(ctx context.Context, cmd *cli.Command) error {
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
	comments, err := client.ListComments(ctx, id)
	if err != nil {
		return err
	}
	if cmd.Bool("json") {
		return printJSON(comments)
	}
	if len(comments) == 0 {
		fmt.Println("no comments")
		return nil
	}
	fmt.Print(renderComments(comments))
	return nil
}
