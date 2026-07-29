package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

type commandFunc func(context.Context, []string, io.Writer) error

var commands = map[string]commandFunc{
	"validate":      runValidate,
	"diff":          runDiff,
	"plan":          runPlan,
	"review":        runReview,
	"approve":       runApprove,
	"backup":        runBackup,
	"apply":         runApply,
	"export-script": runExportScript,
}

func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	switch args[0] {
	case "-h", "--help", "help":
		printUsage(stdout)
		return nil
	}

	run, ok := commands[args[0]]
	if !ok {
		return fmt.Errorf("unknown command %q", args[0])
	}

	return run(ctx, args[1:], stdout)
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "mimic compares configured MongoDB collections and builds safe promotion plans.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  mimic validate --config=mimic.yml")
	fmt.Fprintln(w, "  mimic diff --config=mimic.yml")
	fmt.Fprintln(w, "  mimic plan --config=mimic.yml --out=plans/plan.json")
	fmt.Fprintln(w, "  mimic review --plan=plans/plan.json")
	fmt.Fprintln(w, "  mimic approve --plan=plans/plan.json --out=plans/plan.approved.json")
	fmt.Fprintln(w, "  mimic backup --config=mimic.yml --plan=plans/plan.approved.json --out=backups/run")
	fmt.Fprintln(w, "  mimic apply --plan=plans/plan.approved.json --backup=backups/run --confirm=production")
	fmt.Fprintln(w, "  mimic export-script --plan=plans/plan.approved.json --format=mongodb-js")
}

func configFlagSet(name string) (*flag.FlagSet, *string) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	configPath := fs.String("config", "mimic.yml", "configuration file path")
	return fs, configPath
}

func requireFlagValue(flagName string, value string) error {
	if value == "" {
		return fmt.Errorf("--%s is required", flagName)
	}
	return nil
}

func notImplemented(command string) error {
	return errors.New(command + " is scaffolded but not implemented yet")
}
