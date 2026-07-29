package cli

import (
	"context"
	"flag"
	"io"
)

func runApply(ctx context.Context, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	planPath := fs.String("plan", "", "plan file path")
	confirm := fs.String("confirm", "", "confirmation target name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireFlagValue("plan", *planPath); err != nil {
		return err
	}
	if err := requireFlagValue("confirm", *confirm); err != nil {
		return err
	}
	return notImplemented("apply")
}
