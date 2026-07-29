package cli

import (
	"context"
	"flag"
	"io"
)

func runExportScript(ctx context.Context, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("export-script", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	planPath := fs.String("plan", "", "plan file path")
	format := fs.String("format", "mongodb-js", "export format")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireFlagValue("plan", *planPath); err != nil {
		return err
	}
	if err := requireFlagValue("format", *format); err != nil {
		return err
	}
	return notImplemented("export-script")
}
