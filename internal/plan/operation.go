package plan

import "time"

type Plan struct {
	Kind           string      `json:"kind"`
	Source         string      `json:"source"`
	Target         string      `json:"target"`
	ConfigChecksum string      `json:"configChecksum,omitempty"`
	CreatedAt      time.Time   `json:"createdAt"`
	Operations     []Operation `json:"operations"`
}

type Operation struct {
	Type       string         `json:"type"`
	Collection string         `json:"collection"`
	IndexName  string         `json:"indexName,omitempty"`
	Filter     map[string]any `json:"filter,omitempty"`
	Document   map[string]any `json:"document,omitempty"`
	Update     map[string]any `json:"update,omitempty"`
	Keys       map[string]int `json:"keys,omitempty"`
	Options    map[string]any `json:"options,omitempty"`
}

type ApprovedPlan struct {
	Kind                 string      `json:"kind"`
	Source               string      `json:"source"`
	Target               string      `json:"target"`
	ApprovedAt           time.Time   `json:"approvedAt"`
	Operator             string      `json:"operator,omitempty"`
	OriginalPlanChecksum string      `json:"originalPlanChecksum"`
	ConfigChecksum       string      `json:"configChecksum"`
	ApprovedChecksum     string      `json:"approvedChecksum,omitempty"`
	ApprovedOperations   []Operation `json:"approvedOperations"`
	SkippedOperations    []Operation `json:"skippedOperations"`
}
