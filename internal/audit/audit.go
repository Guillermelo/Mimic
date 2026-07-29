package audit

import "time"

type Run struct {
	StartedAt  time.Time
	FinishedAt time.Time
	User       string
	Source     string
	Target     string
	PlanPath   string
	Inserts    int
	Updates    int
	Deletes    int
	Indexes    int
	Errors     []string
}
