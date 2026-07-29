package mongo

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Endpoint struct {
	URI      string
	Database string
}

type CollectionRule struct {
	Key []string
}

func ValidateConnections(ctx context.Context, source Endpoint, target Endpoint, collections map[string]CollectionRule) error {
	sourceClient, err := Connect(ctx, source.URI)
	if err != nil {
		return fmt.Errorf("source connection failed: %w", err)
	}
	defer sourceClient.Disconnect(context.Background())

	targetClient, err := Connect(ctx, target.URI)
	if err != nil {
		return fmt.Errorf("target connection failed: %w", err)
	}
	defer targetClient.Disconnect(context.Background())

	names := make([]string, 0, len(collections))
	for name := range collections {
		names = append(names, name)
	}
	sourceDB := sourceClient.Database(source.Database)
	targetDB := targetClient.Database(target.Database)
	if err := checkCollectionAccess(ctx, sourceDB, names); err != nil {
		return fmt.Errorf("source collection access failed: %w", err)
	}
	if err := checkCollectionAccess(ctx, targetDB, names); err != nil {
		return fmt.Errorf("target collection access failed: %w", err)
	}
	for name, rule := range collections {
		if err := CheckDuplicateStableKeys(ctx, sourceDB, name, rule.Key); err != nil {
			return fmt.Errorf("source collection %q: %w", name, err)
		}
		if err := CheckDuplicateStableKeys(ctx, targetDB, name, rule.Key); err != nil {
			return fmt.Errorf("target collection %q: %w", name, err)
		}
	}

	return nil
}

func FetchDocuments(ctx context.Context, db *mongo.Database, collection string) ([]map[string]any, error) {
	cursor, err := db.Collection(collection).Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []map[string]any
	for cursor.Next(ctx) {
		var doc map[string]any
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode document: %w", err)
		}
		docs = append(docs, doc)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate documents: %w", err)
	}
	return docs, nil
}

func CheckDuplicateStableKeys(ctx context.Context, db *mongo.Database, collection string, keyFields []string) error {
	docs, err := FetchDocuments(ctx, db, collection)
	if err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for _, doc := range docs {
		key, err := StableKey(doc, keyFields)
		if err != nil {
			return err
		}
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate stable key %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func StableKey(doc map[string]any, fields []string) (string, error) {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		value, ok := doc[field]
		if !ok {
			return "", fmt.Errorf("document missing stable key field %q", field)
		}
		parts = append(parts, fmt.Sprintf("%s=%v", field, value))
	}
	return strings.Join(parts, "|"), nil
}
