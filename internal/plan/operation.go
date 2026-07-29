package plan

import "time"

type Plan struct {
	Source     string      `json:"source"`
	Target     string      `json:"target"`
	CreatedAt  time.Time   `json:"createdAt"`
	Operations []Operation `json:"operations"`
}

type Operation struct {
	Type       string         `json:"type"`
	Collection string         `json:"collection"`
	Filter     map[string]any `json:"filter,omitempty"`
	Document   map[string]any `json:"document,omitempty"`
	Update     map[string]any `json:"update,omitempty"`
	Options    map[string]any `json:"options,omitempty"`
}
