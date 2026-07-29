package plan

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadFile(path string) (Plan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Plan{}, fmt.Errorf("read plan file: %w", err)
	}

	var p Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return Plan{}, fmt.Errorf("parse plan file: %w", err)
	}
	return p, nil
}

func ReadApprovedFile(path string) (ApprovedPlan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ApprovedPlan{}, fmt.Errorf("read approved plan file: %w", err)
	}

	var p ApprovedPlan
	if err := json.Unmarshal(data, &p); err != nil {
		return ApprovedPlan{}, fmt.Errorf("parse approved plan file: %w", err)
	}
	return p, nil
}
