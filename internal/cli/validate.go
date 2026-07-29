package cli

import (
	"context"
	"fmt"
	"io"

	"mimic/internal/config"
	mimicmongo "mimic/internal/mongo"
)

func runValidate(ctx context.Context, args []string, stdout io.Writer) error {
	fs, configPath := configFlagSet("validate")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.LoadFile(*configPath)
	if err != nil {
		return err
	}

	resolved, err := config.Resolve(cfg)
	if err != nil {
		return err
	}

	collections := make(map[string]mimicmongo.CollectionRule, len(cfg.Collections))
	for name, rule := range cfg.Collections {
		collections[name] = mimicmongo.CollectionRule{Key: rule.Key}
	}
	source := mimicmongo.Endpoint{URI: resolved.Source.URI, Database: resolved.Source.Database}
	target := mimicmongo.Endpoint{URI: resolved.Target.URI, Database: resolved.Target.Database}
	if err := mimicmongo.ValidateConnections(ctx, source, target, collections); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "configuration is valid: %s\n", *configPath)
	return nil
}
