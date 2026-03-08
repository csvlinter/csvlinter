package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/csvlinter/csvlinter/pkg/csvlinter"
	"github.com/csvlinter/csvlinter/internal/schema"

	"github.com/urfave/cli/v2"
)

var validateCommand = &cli.Command{
	Name:      "validate",
	Usage:     "Validate a CSV file or STDIN against structure and optional schema",
	ArgsUsage: "<csv-file or - for STDIN>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "schema",
			Aliases: []string{"s"},
			Usage:   "Path to JSON Schema file. If not set, will look for <csv>.schema.json or csvlinter.schema.json in the same or parent directories (see docs)",
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
		&cli.Int64Flag{
			Name:   "max-size",
			Value:  10 * 1024 * 1024, // 10MB default
			Usage:  "Maximum input size in bytes when reading from STDIN",
			Hidden: true,
		},
		&cli.StringFlag{
			Name:  "filename",
			Usage: "Logical filename to use for schema resolution and reporting when reading from STDIN",
		},
		&cli.BoolFlag{
			Name:  "infer-schema",
			Usage: "Infer JSON Schema from CSV data when no schema file is provided; validate against inferred schema",
		},
		&cli.StringFlag{
			Name:  "infer-schema-output",
			Usage: "When using --infer-schema, write the inferred schema to this path",
		},
	},
	Action: validateAction,
}

func exitError(c *cli.Context, format, msg string) error {
	if format == "json" {
		fmt.Fprintf(c.App.Writer, `{"errors":[{"line_number":1,"message":%q}]}`+"\n", msg)
		return cli.Exit("", 1)
	}
	return cli.Exit(msg, 1)
}

func validateAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return exitError(c, c.String("format"), "Error: CSV file path or - for STDIN is required")
	}

	csvPath := c.Args().Get(0)
	format := c.String("format")
	delimiter := c.String("delimiter")
	maxSize := c.Int64("max-size")
	filename := c.String("filename")

	var input io.Reader
	var name string

	if csvPath == "-" {
		input = io.LimitReader(os.Stdin, maxSize)
		if filename != "" {
			name = filename
		} else {
			name = "STDIN"
		}
	} else {
		file, err := os.Open(csvPath)
		if err != nil {
			return exitError(c, format, fmt.Sprintf("Error: Cannot open file '%s': %v", csvPath, err))
		}
		defer file.Close()
		input = file
		name = csvPath
	}

	schemaPath := c.String("schema")
	if schemaPath == "" {
		if csvPath == "-" && filename != "" {
			schemaPath = schema.ResolveSchema(filename)
		} else if csvPath != "-" {
			schemaPath = schema.ResolveSchema(csvPath)
		}
	}
	if schemaPath != "" {
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			return exitError(c, format, fmt.Sprintf("Error: Schema file '%s' does not exist", schemaPath))
		}
	}

	if format != "pretty" && format != "json" {
		return cli.Exit("Error: Format must be 'pretty' or 'json'", 1)
	}

	opts := csvlinter.Options{
		Delimiter:          delimiter,
		FailFast:           c.Bool("fail-fast"),
		Format:             format,
		Output:             c.String("output"),
		Filename:           name,
		SchemaPath:         schemaPath,
		InferSchema:        c.Bool("infer-schema"),
		InferSchemaOutput:  c.String("infer-schema-output"),
	}
	results, err := csvlinter.LintAdvanced(input, opts, c.App.Writer)
	if err != nil {
		return exitError(c, format, err.Error())
	}
	if results != nil && !results.Valid {
		if format == "json" {
			return cli.Exit("", 1)
		}
		return cli.Exit("validation failed", 1)
	}
	return nil
}
