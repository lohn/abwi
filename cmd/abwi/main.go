package main

import (
	"context"
	"fmt"
	"os"

	"github.com/lohn/abwi/internal/cli"
)

func main() {
	if err := cli.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "abwi:", err)
		os.Exit(1)
	}
}
