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

	if err := mimicmongo.ValidateConnections(ctx, resolved.Source.URI, resolved.Target.URI, cfg.CollectionNames()); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "configuration is valid: %s\n", *configPath)
	return nil
}
