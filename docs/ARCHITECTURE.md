# Mimic Architecture

## Goal

Mimic is a Go CLI for controlled MongoDB data promotion between two databases. Its main workflow is:

```txt
source MongoDB -> validate -> diff -> plan -> review -> approve -> backup -> apply -> audit
```

The CLI must promote only explicitly configured configuration data. It must not make one database equal to another.

## Package Layout

```txt
cmd/
  mimic/
    main.go
internal/
  cli/
    root.go
    validate.go
    diff.go
    plan.go
    review.go
    approve.go
    backup.go
    apply.go
    export.go
  config/
    config.go
    load.go
    validate.go
  mongo/
    connect.go
    collections.go
    indexes.go
    backup.go
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
  mimic.yml
testdata/
  mongo/
    source-init/
    target-init/
```

## Boundaries

- `cmd/mimic` only wires the application and exits with process codes.
- `internal/cli` owns command parsing and output formatting.
- `internal/config` owns config loading, defaulting, and validation.
- `internal/mongo` owns MongoDB connectivity, collection/index access, and backup metadata helpers.
- `internal/diff` owns normalized comparison logic.
- `internal/plan` owns serializable operation models, approval artifacts, checksums, and plan file I/O.
- `internal/apply` owns applying a previously approved plan after backup verification.
- `internal/audit` owns execution records.
- `internal/exporters` owns optional script generation.

## Current Safety Boundary

Mimic is review-first and dry-run by default. Commands may inspect MongoDB, build plan artifacts, approve selected operations, and verify backup metadata, but real target writes belong only to the `apply` path.

`apply` must:

1. Refuse raw plans.
2. Read the exact approved plan provided by the operator.
3. Verify the approved plan checksum.
4. Verify backup metadata against the approved plan checksum.
5. Require explicit target confirmation.
6. Avoid recalculating the diff during apply.
7. Fail closed when transaction-backed writes or rollback safety are unavailable.

## Implementation Focus

The next implementation work should stay aligned with the README safety model:

1. Add focused tests for diff generation, plan approval, checksum validation, and array normalization.
2. Implement `export-script --format=mongodb-js`.
3. Implement real backup support through `mongodump` for local MongoDB.
4. Implement transaction-backed `insertOne` and `updateOne` apply behavior.
5. Add integration tests against the disposable Docker MongoDB environment.
