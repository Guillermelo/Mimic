package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"sort"

	"mimic/internal/plan"
)

func runReview(ctx context.Context, args []string, stdout io.Writer) error {
	_ = ctx
	fs := flag.NewFlagSet("review", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	planPath := fs.String("plan", "", "plan file path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireFlagValue("plan", *planPath); err != nil {
		return err
	}
	p, err := plan.ReadFile(*planPath)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Target: %s\n", p.Target)
	fmt.Fprintf(stdout, "Source: %s\n\n", p.Source)
	fmt.Fprintln(stdout, "Collections:")
	summary := summarizeOperations(p.Operations)
	collections := make([]string, 0, len(summary))
	for collection := range summary {
		collections = append(collections, collection)
	}
	sort.Strings(collections)
	for _, collection := range collections {
		counts := summary[collection]
		fmt.Fprintf(stdout, "  %s\n", collection)
		fmt.Fprintf(stdout, "    %d inserts\n", counts["insertOne"])
		fmt.Fprintf(stdout, "    %d updates\n", counts["updateOne"])
		fmt.Fprintf(stdout, "    %d deletes\n", counts["deleteOne"])
		fmt.Fprintf(stdout, "    %d indexes\n\n", counts["createIndex"])
	}
	fmt.Fprintln(stdout, "Risk checks:")
	fmt.Fprintln(stdout, "  deletes: disabled unless approved plan contains deleteOne")
	fmt.Fprintln(stdout, "  backup required: yes")
	fmt.Fprintln(stdout, "No changes have been applied.")
	return nil
}
