package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func checkCollectionAccess(ctx context.Context, db *mongo.Database, collectionNames []string) error {
	_, err := db.ListCollectionNames(ctx, bson.D{})
	return err
}
