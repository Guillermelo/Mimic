package diff

import (
	"fmt"
	"sort"

	"mimic/internal/config"
)

func NormalizeDocument(document map[string]any, ignoredFields []string, arrayRules map[string]config.ArrayRule) (map[string]any, error) {
	ignored := make(map[string]struct{}, len(ignoredFields))
	for _, field := range ignoredFields {
		ignored[field] = struct{}{}
	}

	normalized := make(map[string]any, len(document))
	for key, value := range document {
		if _, skip := ignored[key]; skip {
			continue
		}
		arrayRule, hasArrayRule := arrayRules[key]
		if hasArrayRule {
			normalizedValue, err := normalizeArray(value, arrayRule)
			if err != nil {
				return nil, fmt.Errorf("field %q: %w", key, err)
			}
			normalized[key] = normalizedValue
			continue
		}
		normalized[key] = normalizeValue(value)
	}
	return normalized, nil
}

func normalizeValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, nested := range typed {
			out[key] = normalizeValue(nested)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, nested := range typed {
			out[i] = normalizeValue(nested)
		}
		return out
	default:
		return value
	}
}

func normalizeArray(value any, rule config.ArrayRule) (any, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("configured array rule found non-array value")
	}
	normalized := make([]any, 0, len(items))
	for _, item := range items {
		normalized = append(normalized, normalizeValue(item))
	}
	switch rule.Strategy {
	case "preserveOrder", "replace", "mergeByKey":
		return normalized, nil
	case "sort":
		sort.SliceStable(normalized, func(i, j int) bool {
			return canonical(normalized[i]) < canonical(normalized[j])
		})
		return normalized, nil
	case "set":
		sort.SliceStable(normalized, func(i, j int) bool {
			return canonical(normalized[i]) < canonical(normalized[j])
		})
		deduped := make([]any, 0, len(normalized))
		var last string
		for i, item := range normalized {
			current := canonical(item)
			if i == 0 || current != last {
				deduped = append(deduped, item)
			}
			last = current
		}
		return deduped, nil
	default:
		return nil, fmt.Errorf("unsupported strategy %q", rule.Strategy)
	}
}
