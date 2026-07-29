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
