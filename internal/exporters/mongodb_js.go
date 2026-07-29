package exporters

import (
	"errors"

	"mimic/internal/plan"
)

func MongoDBJS(p plan.Plan) (string, error) {
	if len(p.Operations) == 0 {
		return "", errors.New("cannot export an empty plan")
	}
	return "", errors.New("mongodb-js export is not implemented yet")
}
