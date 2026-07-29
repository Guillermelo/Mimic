package diff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"mimic/internal/config"
	"mimic/internal/mongo"
	"mimic/internal/plan"
)

type DocumentChange struct {
	Collection string
	Key        map[string]any
	Type       string
	Before     map[string]any
	After      map[string]any
}

func BuildCollectionOperations(collection string, rule config.CollectionRule, defaults config.Defaults, sourceDocs []map[string]any, targetDocs []map[string]any) ([]plan.Operation, []DocumentChange, error) {
	ignored := append([]string{}, defaults.IgnoreFields...)
	ignored = append(ignored, rule.IgnoreFields...)
	sourceByKey, err := indexDocuments(sourceDocs, rule.Key, ignored, rule.Arrays)
	if err != nil {
		return nil, nil, fmt.Errorf("source: %w", err)
	}
	targetByKey, err := indexDocuments(targetDocs, rule.Key, ignored, rule.Arrays)
	if err != nil {
		return nil, nil, fmt.Errorf("target: %w", err)
	}

	keys := make([]string, 0, len(sourceByKey))
	for key := range sourceByKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var ops []plan.Operation
	var changes []DocumentChange
	for _, key := range keys {
		source := sourceByKey[key]
		target, exists := targetByKey[key]
		filter := stableFilter(source.original, rule.Key)
		if !exists {
			ops = append(ops, plan.Operation{
				Type:       "insertOne",
				Collection: collection,
				Document:   source.normalized,
			})
			changes = append(changes, DocumentChange{Collection: collection, Key: filter, Type: "insert", After: source.normalized})
			continue
		}
		if reflect.DeepEqual(source.normalized, target.normalized) {
			continue
		}
		set, unset := fieldPatch(source.normalized, target.normalized)
		update := map[string]any{}
		if len(set) > 0 {
			update["$set"] = set
		}
		if len(unset) > 0 {
			update["$unset"] = unset
		}
		ops = append(ops, plan.Operation{
			Type:       "updateOne",
			Collection: collection,
			Filter:     filter,
			Update:     update,
			Options:    map[string]any{"upsert": true},
		})
		changes = append(changes, DocumentChange{Collection: collection, Key: filter, Type: "update", Before: target.normalized, After: source.normalized})
	}
	return ops, changes, nil
}

type indexedDocument struct {
	original   map[string]any
	normalized map[string]any
}

func indexDocuments(docs []map[string]any, keyFields []string, ignored []string, arrays map[string]config.ArrayRule) (map[string]indexedDocument, error) {
	indexed := map[string]indexedDocument{}
	for _, doc := range docs {
		key, err := mongo.StableKey(doc, keyFields)
		if err != nil {
			return nil, err
		}
		normalized, err := NormalizeDocument(doc, ignored, arrays)
		if err != nil {
			return nil, err
		}
		indexed[key] = indexedDocument{original: doc, normalized: normalized}
	}
	return indexed, nil
}

func stableFilter(doc map[string]any, fields []string) map[string]any {
	filter := make(map[string]any, len(fields))
	for _, field := range fields {
		filter[field] = doc[field]
	}
	return filter
}

func fieldPatch(source map[string]any, target map[string]any) (map[string]any, map[string]any) {
	set := map[string]any{}
	unset := map[string]any{}
	for key, sourceValue := range source {
		if !reflect.DeepEqual(sourceValue, target[key]) {
			set[key] = sourceValue
		}
	}
	for key := range target {
		if _, ok := source[key]; !ok {
			unset[key] = ""
		}
	}
	return set, unset
}

func canonical(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}
