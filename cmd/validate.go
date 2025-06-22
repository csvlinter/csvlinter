package cmd

import (
	"fmt"
	"os"

	"csvlinter/internal/reporter"
	"csvlinter/internal/schema"
	"csvlinter/internal/validator"

	"github.com/urfave/cli/v2"
)

var validateCommand = &cli.Command{
	Name:      "validate",
	Usage:     "Validate a CSV file against structure and optional schema",
	ArgsUsage: "<csv-file>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "schema",
			Aliases: []string{"s"},
			Usage:   "Path to JSON Schema file",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "Output file for structured validation results",
		},
		&cli.StringFlag{
			Name:    "format",
			Aliases: []string{"f"},
			Value:   "pretty",
			Usage:   "Output format (pretty, json)",
		},
		&cli.StringFlag{
			Name:    "delimiter",
			Aliases: []string{"d"},
			Value:   ",",
			Usage:   "Delimiter character (defaults to comma)",
		},
		&cli.BoolFlag{
			Name:    "fail-fast",
			Aliases: []string{"ff"},
			Usage:   "Stop after first error",
		},
	},
	Action: validateAction,
}

func validateAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit("Error: CSV file path is required", 1)
	}

	csvPath := c.Args().Get(0)
	schemaPath := c.String("schema")
	outputPath := c.String("output")
	format := c.String("format")
	delimiter := c.String("delimiter")
	failFast := c.Bool("fail-fast")

	// Validate input file exists
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		return cli.Exit(fmt.Sprintf("Error: CSV file '%s' does not exist", csvPath), 1)
	}

	// Validate schema file if provided
	var schemaValidator *schema.Validator
	if schemaPath != "" {
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			return cli.Exit(fmt.Sprintf("Error: Schema file '%s' does not exist", schemaPath), 1)
		}

		var err error
		schemaValidator, err = schema.NewValidator(schemaPath)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Error loading schema: %v", err), 1)
		}
	}

	// Validate format
	if format != "pretty" && format != "json" {
		return cli.Exit("Error: Format must be 'pretty' or 'json'", 1)
	}

	// Create validator
	v := validator.New(csvPath, delimiter, schemaValidator, failFast)

	// Run validation
	results, err := v.Validate()
	if err != nil {
		// This will now only catch operational errors like file not found during parsing
		return cli.Exit(fmt.Sprintf("Error during validation: %v", err), 1)
	}

	// Create reporter
	r := reporter.New(format, outputPath)

	// Output results
	if err := r.Report(results, c.App.Writer); err != nil {
		return cli.Exit(fmt.Sprintf("Error writing output: %v", err), 1)
	}

	return nil
}
