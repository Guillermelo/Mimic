package diff

type DocumentChange struct {
	Collection string
	Key        map[string]any
	Type       string
}
