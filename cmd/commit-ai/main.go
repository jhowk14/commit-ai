package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/jhowk14/commit-ai/v2/internal/app"
)

var version = "2.0.3"

func main() {
	application := app.New(version, os.Stdin, os.Stdout, os.Stderr)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := application.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "❌", err)
		os.Exit(1)
	}
}
