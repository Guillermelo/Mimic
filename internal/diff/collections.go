package diff

type CollectionDiff struct {
	Collection string
	Changes    []DocumentChange
}
