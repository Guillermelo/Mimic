package apply

import (
	"errors"

	"mimic/internal/plan"
)

func ValidatePlan(p plan.Plan) error {
	if p.Target == "" {
		return errors.New("plan target is required")
	}
	for i, op := range p.Operations {
		if op.Type == "" {
			return errors.New("operation type is required")
		}
		if op.Collection == "" {
			return errors.New("operation collection is required")
		}
		_ = i
	}
	return nil
}
