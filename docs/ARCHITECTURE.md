# Architecture

## Goal

Mimic is a Go CLI for controlled MongoDB data promotion between two databases. Its main workflow is:

```txt
source MongoDB -> diff -> plan -> reviewed apply -> audit
```

The CLI must promote only explicitly configured configuration data. It must not make one database equal to another.

## Proposed Package Layout

```txt
cmd/
  mongo-promote/
    main.go
internal/
  cli/
    root.go
    validate.go
    diff.go
    plan.go
    apply.go
    export.go
  config/
    config.go
    load.go
    validate.go
  mongo/
    client.go
    collections.go
    indexes.go
  diff/
    normalize.go
    documents.go
    collections.go
    indexes.go
  plan/
    operation.go
    build.go
    read.go
    write.go
  apply/
    apply.go
    validators.go
  audit/
    audit.go
  exporters/
    mongodb_js.go
examples/
  mongo-promote.yml
testdata/
  mongo/
    source-init/
    target-init/
```

## Boundaries

- `cmd/mongo-promote` only wires the application and exits with process codes.
- `internal/cli` owns command parsing and output formatting.
- `internal/config` owns config loading, defaulting, and validation.
- `internal/mongo` owns MongoDB connectivity and low-level collection/index access.
- `internal/diff` owns normalized comparison logic.
- `internal/plan` owns serializable operation models and plan file I/O.
- `internal/apply` owns applying a previously generated plan.
- `internal/audit` owns execution records.
- `internal/exporters` owns optional script generation.

## First Milestone

The first implementation milestone should be intentionally small:

1. Create a CLI shell with `validate`.
2. Load YAML config.
3. Resolve source and target MongoDB URIs from environment variables.
4. Validate that source and target are not the same connection string.
5. Validate collection rules and stable keys without touching data.
6. Add unit tests for config validation.

After that, add MongoDB connectivity and integration tests against the local Docker environment.
