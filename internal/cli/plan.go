package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"mimic/internal/plan"
)

func runPlan(ctx context.Context, args []string, stdout io.Writer) error {
	fs, configPath := configFlagSet("plan")
	out := fs.String("out", "", "plan output file path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireFlagValue("out", *out); err != nil {
		return err
	}
	built, err := buildPlan(ctx, *configPath)
	if err != nil {
		return err
	}
	if err := plan.WriteFile(*out, built.Plan); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "plan written: %s\n", *out)
	fmt.Fprintf(stdout, "operations: %d\n", len(built.Plan.Operations))
	fmt.Fprintln(stdout, "No changes have been applied.")
	return nil
}

func planFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}
