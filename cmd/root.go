package cmd

import (
	"os"

	"github.com/urfave/cli/v2"
)

// Version is set at build time via -ldflags (e.g. goreleaser sets it from the Git tag).
var Version = "dev"

// Execute runs the CLI application
func Execute() error {
	app := &cli.App{
		Name:        "csvlinter",
		Usage:       "A modern, streaming-first CSV validator with JSON Schema support",
		Description: "Validates structure, content, and encoding of CSV files — built for CI, CLI, and editor integration",
		Version:     Version,
		Commands: []*cli.Command{
			validateCommand,
		},
	}

	return app.Run(os.Args)
}
