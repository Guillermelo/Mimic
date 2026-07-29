package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"

	"mimic/internal/config"
	mimicmongo "mimic/internal/mongo"
	"mimic/internal/plan"
)

func runBackup(ctx context.Context, args []string, stdout io.Writer) error {
	_ = ctx
	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	configPath := fs.String("config", "", "configuration file path")
	planPath := fs.String("plan", "", "approved plan file path")
	out := fs.String("out", "", "backup output directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireFlagValue("config", *configPath); err != nil {
		return err
	}
	if err := requireFlagValue("plan", *planPath); err != nil {
		return err
	}
	if err := requireFlagValue("out", *out); err != nil {
		return err
	}
	cfg, err := config.LoadFile(*configPath)
	if err != nil {
		return err
	}
	resolved, err := config.Resolve(cfg)
	if err != nil {
		return err
	}
	approved, err := plan.ReadApprovedFile(*planPath)
	if err != nil {
		return err
	}
	if approved.Kind != "mimic.approvedPlan" {
		return fmt.Errorf("backup requires an approved plan")
	}
	configChecksum, err := plan.ChecksumFile(*configPath)
	if err != nil {
		return err
	}
	if approved.ConfigChecksum != "" && approved.ConfigChecksum != configChecksum {
		return fmt.Errorf("approved plan config checksum does not match current config")
	}
	metadata := mimicmongo.BackupMetadata{
		SourceLabel:          resolved.Source.Label,
		TargetLabel:          resolved.Target.Label,
		SourceDatabase:       resolved.Source.Database,
		TargetDatabase:       resolved.Target.Database,
		IncludedCollections:  cfg.CollectionNames(),
		CreatedAt:            time.Now().UTC(),
		ToolVersion:          "mimic-dev",
		ConfigChecksum:       configChecksum,
		ApprovedPlanChecksum: approved.ApprovedChecksum,
	}
	if err := mimicmongo.WriteBackupMetadata(*out, metadata); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "backup metadata written: %s\n", *out)
	fmt.Fprintln(stdout, "source and target backup directories created.")
	return nil
}
