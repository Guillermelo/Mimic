package plan

func New(source string, target string) Plan {
	return Plan{
		Source:     source,
		Target:     target,
		Operations: []Operation{},
	}
}
