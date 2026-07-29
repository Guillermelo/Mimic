package apply

import (
	"context"

	"mimic/internal/plan"
)

func Plan(ctx context.Context, p plan.Plan) error {
	return ValidatePlan(p)
}
