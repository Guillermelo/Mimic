package mongo

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func ValidateConnections(ctx context.Context, sourceURI string, targetURI string, collectionNames []string) error {
	sourceClient, err := Connect(ctx, sourceURI)
	if err != nil {
		return fmt.Errorf("source connection failed: %w", err)
	}
	defer sourceClient.Disconnect(context.Background())

	targetClient, err := Connect(ctx, targetURI)
	if err != nil {
		return fmt.Errorf("target connection failed: %w", err)
	}
	defer targetClient.Disconnect(context.Background())

	sourceDB, err := databaseName(sourceURI)
	if err != nil {
		return fmt.Errorf("source database name: %w", err)
	}
	targetDB, err := databaseName(targetURI)
	if err != nil {
		return fmt.Errorf("target database name: %w", err)
	}

	if err := checkCollectionAccess(ctx, sourceClient.Database(sourceDB), collectionNames); err != nil {
		return fmt.Errorf("source collection access failed: %w", err)
	}
	if err := checkCollectionAccess(ctx, targetClient.Database(targetDB), collectionNames); err != nil {
		return fmt.Errorf("target collection access failed: %w", err)
	}

	return nil
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
