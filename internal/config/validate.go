package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

var validCollectionModes = map[string]struct{}{
	"upsert": {},
}

var validArrayStrategies = map[string]struct{}{
	"preserveOrder": {},
	"sort":          {},
	"set":           {},
	"replace":       {},
	"mergeByKey":    {},
}

func Validate(cfg Config) error {
	if strings.TrimSpace(cfg.Source.URIEnv) == "" {
		return errors.New("source.uriEnv is required")
	}
	if strings.TrimSpace(cfg.Target.URIEnv) == "" {
		return errors.New("target.uriEnv is required")
	}
	if len(cfg.Collections) == 0 {
		return errors.New("at least one collection must be configured")
	}

	for name, rule := range cfg.Collections {
		if strings.TrimSpace(name) == "" {
			return errors.New("collection names cannot be empty")
		}
		if len(rule.Key) == 0 {
			return fmt.Errorf("collection %q must define a stable key", name)
		}
		for _, key := range rule.Key {
			if strings.TrimSpace(key) == "" {
				return fmt.Errorf("collection %q has an empty stable key field", name)
			}
		}
		if _, ok := validCollectionModes[rule.Mode]; !ok {
			return fmt.Errorf("collection %q has unsupported mode %q", name, rule.Mode)
		}
		for field, arrayRule := range rule.Arrays {
			if _, ok := validArrayStrategies[arrayRule.Strategy]; !ok {
				return fmt.Errorf("collection %q array %q has unsupported strategy %q", name, field, arrayRule.Strategy)
			}
			if arrayRule.Strategy == "mergeByKey" && len(arrayRule.Key) == 0 {
				return fmt.Errorf("collection %q array %q must define key for mergeByKey", name, field)
			}
		}
		for field, ref := range rule.References {
			if strings.TrimSpace(field) == "" {
				return fmt.Errorf("collection %q has an empty reference field", name)
			}
			if strings.TrimSpace(ref.Collection) == "" {
				return fmt.Errorf("collection %q reference %q must define collection", name, field)
			}
			if len(ref.Key) == 0 {
				return fmt.Errorf("collection %q reference %q must define key", name, field)
			}
		}
	}

	for collection, indexes := range cfg.Indexes {
		if _, ok := cfg.Collections[collection]; !ok {
			return fmt.Errorf("indexes configured for non-allowed collection %q", collection)
		}
		for i, index := range indexes {
			if len(index.Keys) == 0 {
				return fmt.Errorf("index %d for collection %q must define keys", i, collection)
			}
			for field, direction := range index.Keys {
				if strings.TrimSpace(field) == "" {
					return fmt.Errorf("index %d for collection %q has an empty key", i, collection)
				}
				if direction != 1 && direction != -1 {
					return fmt.Errorf("index %d for collection %q key %q must be 1 or -1", i, collection, field)
				}
			}
			if strings.TrimSpace(index.Options.Name) == "" {
				return fmt.Errorf("index %d for collection %q must define options.name", i, collection)
			}
		}
	}

	return nil
}

func Resolve(cfg Config) (ResolvedConfig, error) {
	sourceURI, sourceOK := os.LookupEnv(cfg.Source.URIEnv)
	if !sourceOK || strings.TrimSpace(sourceURI) == "" {
		return ResolvedConfig{}, fmt.Errorf("environment variable %s is required", cfg.Source.URIEnv)
	}

	targetURI, targetOK := os.LookupEnv(cfg.Target.URIEnv)
	if !targetOK || strings.TrimSpace(targetURI) == "" {
		return ResolvedConfig{}, fmt.Errorf("environment variable %s is required", cfg.Target.URIEnv)
	}

	sourceDB, err := databaseName(sourceURI)
	if err != nil {
		return ResolvedConfig{}, fmt.Errorf("source database name: %w", err)
	}
	targetDB, err := databaseName(targetURI)
	if err != nil {
		return ResolvedConfig{}, fmt.Errorf("target database name: %w", err)
	}

	if sourceURI == targetURI || sourceDB == targetDB && strings.TrimSpace(sourceURI) == strings.TrimSpace(targetURI) {
		return ResolvedConfig{}, errors.New("source and target MongoDB connections must be different")
	}

	return ResolvedConfig{
		Config: cfg,
		Source: ResolvedEndpoint{URI: sourceURI, Database: sourceDB, Label: endpointLabel(cfg.Source.Label, sourceDB)},
		Target: ResolvedEndpoint{URI: targetURI, Database: targetDB, Label: endpointLabel(cfg.Target.Label, targetDB)},
	}, nil
}

func endpointLabel(label string, database string) string {
	if strings.TrimSpace(label) != "" {
		return strings.TrimSpace(label)
	}
	return database
}

func databaseName(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	name := strings.Trim(parsed.Path, "/")
	if name == "" {
		return "", fmt.Errorf("URI must include a database name")
	}
	if strings.Contains(name, "/") {
		return "", fmt.Errorf("URI must include exactly one database name")
	}
	return name, nil
}
