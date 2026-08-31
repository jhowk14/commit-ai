package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jhowk14/commit-ai/v2/internal/app"
)

var version = "2.0.2"

func main() {
	application := app.New(version, os.Stdin, os.Stdout, os.Stderr)
	if err := application.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "❌", err)
		os.Exit(1)
	}
}
