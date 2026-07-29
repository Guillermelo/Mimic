package cli

import (
	"context"
	"fmt"
	"sort"

	"mimic/internal/config"
	"mimic/internal/diff"
	mimicmongo "mimic/internal/mongo"
	"mimic/internal/plan"
)

type builtPlan struct {
	Plan          plan.Plan
	DocumentDiffs []diff.DocumentChange
	IndexDiffs    []diff.IndexChange
}

func buildPlan(ctx context.Context, configPath string) (builtPlan, error) {
	cfg, err := config.LoadFile(configPath)
	if err != nil {
		return builtPlan{}, err
	}
	resolved, err := config.Resolve(cfg)
	if err != nil {
		return builtPlan{}, err
	}
	sourceClient, err := mimicmongo.Connect(ctx, resolved.Source.URI)
	if err != nil {
		return builtPlan{}, fmt.Errorf("source connection failed: %w", err)
	}
	defer sourceClient.Disconnect(context.Background())
	targetClient, err := mimicmongo.Connect(ctx, resolved.Target.URI)
	if err != nil {
		return builtPlan{}, fmt.Errorf("target connection failed: %w", err)
	}
	defer targetClient.Disconnect(context.Background())

	sourceDB := sourceClient.Database(resolved.Source.Database)
	targetDB := targetClient.Database(resolved.Target.Database)
	p := plan.New(resolved.Source.Label, resolved.Target.Label, "")
	if checksum, err := plan.ChecksumFile(configPath); err == nil {
		p.ConfigChecksum = checksum
	}

	var documentDiffs []diff.DocumentChange
	names := cfg.CollectionNames()
	for _, collection := range names {
		rule := cfg.Collections[collection]
		if err := mimicmongo.CheckDuplicateStableKeys(ctx, sourceDB, collection, rule.Key); err != nil {
			return builtPlan{}, fmt.Errorf("source collection %q: %w", collection, err)
		}
		if err := mimicmongo.CheckDuplicateStableKeys(ctx, targetDB, collection, rule.Key); err != nil {
			return builtPlan{}, fmt.Errorf("target collection %q: %w", collection, err)
		}
		sourceDocs, err := mimicmongo.FetchDocuments(ctx, sourceDB, collection)
		if err != nil {
			return builtPlan{}, fmt.Errorf("source collection %q: %w", collection, err)
		}
		targetDocs, err := mimicmongo.FetchDocuments(ctx, targetDB, collection)
		if err != nil {
			return builtPlan{}, fmt.Errorf("target collection %q: %w", collection, err)
		}
		ops, changes, err := diff.BuildCollectionOperations(collection, rule, cfg.Defaults, sourceDocs, targetDocs)
		if err != nil {
			return builtPlan{}, fmt.Errorf("diff collection %q: %w", collection, err)
		}
		p.Operations = append(p.Operations, ops...)
		documentDiffs = append(documentDiffs, changes...)
	}

	existingIndexes := map[string][]mimicmongo.IndexSpec{}
	for collection := range cfg.Indexes {
		indexes, err := mimicmongo.ListIndexSpecs(ctx, targetDB, collection)
		if err != nil {
			return builtPlan{}, fmt.Errorf("target collection %q indexes: %w", collection, err)
		}
		existingIndexes[collection] = indexes
	}
	indexOps, indexDiffs := diff.BuildIndexOperations(cfg.Indexes, existingIndexes)
	p.Operations = append(p.Operations, indexOps...)
	sort.SliceStable(p.Operations, func(i, j int) bool {
		if p.Operations[i].Collection == p.Operations[j].Collection {
			return p.Operations[i].Type < p.Operations[j].Type
		}
		return p.Operations[i].Collection < p.Operations[j].Collection
	})
	return builtPlan{Plan: p, DocumentDiffs: documentDiffs, IndexDiffs: indexDiffs}, nil
}

func summarizeOperations(ops []plan.Operation) map[string]map[string]int {
	summary := map[string]map[string]int{}
	for _, op := range ops {
		if _, ok := summary[op.Collection]; !ok {
			summary[op.Collection] = map[string]int{}
		}
		summary[op.Collection][op.Type]++
	}
	return summary
}
