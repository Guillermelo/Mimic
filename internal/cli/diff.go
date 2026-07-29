package cli

import (
	"context"
	"fmt"
	"io"
)

func runDiff(ctx context.Context, args []string, stdout io.Writer) error {
	fs, configPath := configFlagSet("diff")
	if err := fs.Parse(args); err != nil {
		return err
	}
	built, err := buildPlan(ctx, *configPath)
	if err != nil {
		return err
	}
	if len(built.Plan.Operations) == 0 {
		fmt.Fprintln(stdout, "No differences found.")
		return nil
	}
	for _, change := range built.DocumentDiffs {
		switch change.Type {
		case "insert":
			fmt.Fprintf(stdout, "%s\n  + insert %v\n", change.Collection, change.Key)
		case "update":
			fmt.Fprintf(stdout, "%s\n  ~ update %v\n", change.Collection, change.Key)
		}
	}
	for _, change := range built.IndexDiffs {
		fmt.Fprintf(stdout, "indexes\n  + %s %s.%s\n", change.Type, change.Collection, change.Name)
	}
	fmt.Fprintln(stdout, "No changes have been applied.")
	return nil
}
