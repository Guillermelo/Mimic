package cli

import (
	"context"
	"flag"
	"io"
)

func runPlan(ctx context.Context, args []string, stdout io.Writer) error {
	fs, _ := configFlagSet("plan")
	out := fs.String("out", "", "plan output file path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireFlagValue("out", *out); err != nil {
		return err
	}
	return notImplemented("plan")
}

func planFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}
