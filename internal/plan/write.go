package plan

import (
	"encoding/json"
	"fmt"
	"os"
)

func WriteFile(path string, p Plan) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("encode plan file: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write plan file: %w", err)
	}
	return nil
}
