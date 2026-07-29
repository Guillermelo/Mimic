package config

import "testing"

func TestValidateRequiresCollections(t *testing.T) {
	cfg := Config{
		Source: Endpoint{URIEnv: "SOURCE_URI"},
		Target: Endpoint{URIEnv: "TARGET_URI"},
	}

	if err := Validate(cfg); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateRequiresStableKey(t *testing.T) {
	cfg := Config{
		Source: Endpoint{URIEnv: "SOURCE_URI"},
		Target: Endpoint{URIEnv: "TARGET_URI"},
		Collections: map[string]CollectionRule{
			"settings": {Mode: "upsert"},
		},
	}

	if err := Validate(cfg); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateAcceptsUpsertCollection(t *testing.T) {
	cfg := Config{
		Source: Endpoint{URIEnv: "SOURCE_URI"},
		Target: Endpoint{URIEnv: "TARGET_URI"},
		Collections: map[string]CollectionRule{
			"settings": {
				Key:  []string{"key"},
				Mode: "upsert",
			},
		},
	}

	if err := Validate(cfg); err != nil {
		t.Fatalf("expected valid config: %v", err)
	}
}
