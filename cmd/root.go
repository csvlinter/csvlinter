package cmd

import (
	"os"

	"github.com/urfave/cli/v2"
)

// Execute runs the CLI application
func Execute() error {
	app := &cli.App{
		Name:        "csvlinter",
		Usage:       "A modern, streaming-first CSV validator with JSON Schema support",
		Description: "Validates structure, content, and encoding of CSV files — built for CI, CLI, and editor integration",
		Version:     "1.0.0",
		Commands: []*cli.Command{
			validateCommand,
		},
	}

	return app.Run(os.Args)
}
