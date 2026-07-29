package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func checkCollectionAccess(ctx context.Context, db *mongo.Database, collectionNames []string) error {
	_, err := db.ListCollectionNames(ctx, bson.D{})
	return err
}

type IndexSpec struct {
	Name   string
	Keys   map[string]int
	Unique bool
}

func ListIndexSpecs(ctx context.Context, db *mongo.Database, collection string) ([]IndexSpec, error) {
	cursor, err := db.Collection(collection).Indexes().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list indexes: %w", err)
	}
	defer cursor.Close(ctx)

	var specs []IndexSpec
	for cursor.Next(ctx) {
		var raw struct {
			Name   string `bson:"name"`
			Key    bson.D `bson:"key"`
			Unique bool   `bson:"unique"`
		}
		if err := cursor.Decode(&raw); err != nil {
			return nil, fmt.Errorf("decode index: %w", err)
		}
		keys := map[string]int{}
		for _, element := range raw.Key {
			switch value := element.Value.(type) {
			case int32:
				keys[element.Key] = int(value)
			case int64:
				keys[element.Key] = int(value)
			case int:
				keys[element.Key] = value
			}
		}
		specs = append(specs, IndexSpec{Name: raw.Name, Keys: keys, Unique: raw.Unique})
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate indexes: %w", err)
	}
	return specs, nil
}
