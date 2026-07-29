package diff

func NormalizeDocument(document map[string]any, ignoredFields []string) map[string]any {
	ignored := make(map[string]struct{}, len(ignoredFields))
	for _, field := range ignoredFields {
		ignored[field] = struct{}{}
	}

	normalized := make(map[string]any, len(document))
	for key, value := range document {
		if _, skip := ignored[key]; skip {
			continue
		}
		normalized[key] = value
	}
	return normalized
}
