package apply

import (
	"context"
	"errors"
	"fmt"

	"mimic/internal/plan"
)

func Plan(ctx context.Context, p plan.Plan) error {
	_ = ctx
	return ValidatePlan(p)
}

func ApprovedPlan(ctx context.Context, p plan.ApprovedPlan) error {
	_ = ctx
	if p.Kind != "mimic.approvedPlan" {
		return errors.New("approved plan kind is required")
	}
	if p.ApprovedChecksum == "" {
		return errors.New("approved plan checksum is required")
	}
	for _, op := range p.ApprovedOperations {
		switch op.Type {
		case "insertOne", "updateOne":
			continue
		case "createIndex":
			return fmt.Errorf("createIndex requires a separate explicit maintenance apply phase")
		default:
			return fmt.Errorf("operation type %q is not supported by apply", op.Type)
		}
	}
	return errors.New("MongoDB write apply is intentionally blocked until transaction execution is configured")
}
