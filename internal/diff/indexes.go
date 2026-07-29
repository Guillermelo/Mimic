package diff

import (
	"reflect"
	"sort"

	"mimic/internal/config"
	"mimic/internal/mongo"
	"mimic/internal/plan"
)

type IndexChange struct {
	Collection string
	Name       string
	Type       string
}

func BuildIndexOperations(expected map[string][]config.IndexRule, existing map[string][]mongo.IndexSpec) ([]plan.Operation, []IndexChange) {
	var collections []string
	for collection := range expected {
		collections = append(collections, collection)
	}
	sort.Strings(collections)

	var ops []plan.Operation
	var changes []IndexChange
	for _, collection := range collections {
		existingByName := map[string]mongo.IndexSpec{}
		for _, index := range existing[collection] {
			existingByName[index.Name] = index
		}
		for _, expectedIndex := range expected[collection] {
			current, ok := existingByName[expectedIndex.Options.Name]
			if ok && reflect.DeepEqual(current.Keys, expectedIndex.Keys) && current.Unique == expectedIndex.Options.Unique {
				continue
			}
			operation := plan.Operation{
				Type:       "createIndex",
				Collection: collection,
				IndexName:  expectedIndex.Options.Name,
				Keys:       expectedIndex.Keys,
				Options: map[string]any{
					"name":   expectedIndex.Options.Name,
					"unique": expectedIndex.Options.Unique,
				},
			}
			ops = append(ops, operation)
			changes = append(changes, IndexChange{Collection: collection, Name: expectedIndex.Options.Name, Type: "createIndex"})
		}
	}
	return ops, changes
}
