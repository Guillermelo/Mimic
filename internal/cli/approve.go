package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"mimic/internal/plan"
)

func runApprove(ctx context.Context, args []string, stdout io.Writer) error {
	_ = ctx
	fs := flag.NewFlagSet("approve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	planPath := fs.String("plan", "", "plan file path")
	out := fs.String("out", "", "approved plan output file path")
	collections := fs.String("collections", "", "comma-separated collections to approve")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireFlagValue("plan", *planPath); err != nil {
		return err
	}
	if err := requireFlagValue("out", *out); err != nil {
		return err
	}
	raw, err := plan.ReadFile(*planPath)
	if err != nil {
		return err
	}
	rawChecksum, err := plan.ChecksumFile(*planPath)
	if err != nil {
		return err
	}
	approved, skipped := splitApprovedOperations(raw.Operations, *collections)
	artifact := plan.ApprovedPlan{
		Kind:                 "mimic.approvedPlan",
		Source:               raw.Source,
		Target:               raw.Target,
		ApprovedAt:           time.Now().UTC(),
		Operator:             os.Getenv("USERNAME"),
		OriginalPlanChecksum: rawChecksum,
		ConfigChecksum:       raw.ConfigChecksum,
		ApprovedOperations:   approved,
		SkippedOperations:    skipped,
	}
	if err := plan.WriteApprovedFile(*out, artifact); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "approved plan written: %s\n", *out)
	fmt.Fprintf(stdout, "approved operations: %d\n", len(approved))
	fmt.Fprintf(stdout, "skipped operations: %d\n", len(skipped))
	return nil
}

func splitApprovedOperations(ops []plan.Operation, collections string) ([]plan.Operation, []plan.Operation) {
	if strings.TrimSpace(collections) == "" {
		return ops, nil
	}
	allowed := map[string]struct{}{}
	for _, collection := range strings.Split(collections, ",") {
		allowed[strings.TrimSpace(collection)] = struct{}{}
	}
	var approved []plan.Operation
	var skipped []plan.Operation
	for _, op := range ops {
		if _, ok := allowed[op.Collection]; ok {
			approved = append(approved, op)
		} else {
			skipped = append(skipped, op)
		}
	}
	return approved, skipped
}
