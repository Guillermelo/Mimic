package mongo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type BackupMetadata struct {
	SourceLabel          string    `json:"sourceLabel"`
	TargetLabel          string    `json:"targetLabel"`
	SourceDatabase       string    `json:"sourceDatabase"`
	TargetDatabase       string    `json:"targetDatabase"`
	IncludedCollections  []string  `json:"includedCollections"`
	CreatedAt            time.Time `json:"createdAt"`
	ToolVersion          string    `json:"toolVersion"`
	ConfigChecksum       string    `json:"configChecksum"`
	ApprovedPlanChecksum string    `json:"approvedPlanChecksum"`
}

func WriteBackupMetadata(path string, metadata BackupMetadata) error {
	if err := os.MkdirAll(filepath.Join(path, "source"), 0o755); err != nil {
		return fmt.Errorf("create source backup directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(path, "target"), 0o755); err != nil {
		return fmt.Errorf("create target backup directory: %w", err)
	}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("encode backup metadata: %w", err)
	}
	if err := os.WriteFile(filepath.Join(path, "metadata.json"), data, 0o600); err != nil {
		return fmt.Errorf("write backup metadata: %w", err)
	}
	return nil
}

func ReadBackupMetadata(path string) (BackupMetadata, error) {
	data, err := os.ReadFile(filepath.Join(path, "metadata.json"))
	if err != nil {
		return BackupMetadata{}, fmt.Errorf("read backup metadata: %w", err)
	}
	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return BackupMetadata{}, fmt.Errorf("parse backup metadata: %w", err)
	}
	if _, err := os.Stat(filepath.Join(path, "source")); err != nil {
		return BackupMetadata{}, fmt.Errorf("source backup directory missing: %w", err)
	}
	if _, err := os.Stat(filepath.Join(path, "target")); err != nil {
		return BackupMetadata{}, fmt.Errorf("target backup directory missing: %w", err)
	}
	return metadata, nil
}
