package config

import (
	"errors"
	"fmt"
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

	if sourceURI == targetURI {
		return ResolvedConfig{}, errors.New("source and target MongoDB URIs must be different")
	}

	return ResolvedConfig{
		Config: cfg,
		Source: ResolvedEndpoint{URI: sourceURI},
		Target: ResolvedEndpoint{URI: targetURI},
	}, nil
}
