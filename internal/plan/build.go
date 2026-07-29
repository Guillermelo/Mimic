package plan

import "time"

func New(source string, target string, configChecksum string) Plan {
	return Plan{
		Kind:           "mimic.plan",
		Source:         source,
		Target:         target,
		ConfigChecksum: configChecksum,
		CreatedAt:      time.Now().UTC(),
		Operations:     []Operation{},
	}
}
