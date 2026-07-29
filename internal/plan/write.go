package plan

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func WriteFile(path string, p Plan) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("encode plan file: %w", err)
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create plan directory: %w", err)
		}
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write plan file: %w", err)
	}
	return nil
}

func WriteApprovedFile(path string, p ApprovedPlan) error {
	copy := p
	copy.ApprovedChecksum = ""
	checksum, err := ChecksumValue(copy)
	if err != nil {
		return err
	}
	copy.ApprovedChecksum = checksum
	data, err := json.MarshalIndent(copy, "", "  ")
	if err != nil {
		return fmt.Errorf("encode approved plan file: %w", err)
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create approved plan directory: %w", err)
		}
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write approved plan file: %w", err)
	}
	return nil
}

func ChecksumFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read checksum input: %w", err)
	}
	return ChecksumBytes(data), nil
}

func ChecksumValue(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode checksum input: %w", err)
	}
	return ChecksumBytes(data), nil
}

func ChecksumBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", sum)
}
