# Mimic Progress

This document summarizes the current implementation status of Mimic and how the app works today.

## Current State

Mimic is now a Go CLI named `mimic`. The project is aligned with the README direction at the package and command level, but it is not production-ready yet.

The implemented code focuses on the safe review-first workflow:

```txt
validate -> diff -> plan -> review -> approve -> backup -> apply
```

The app still fails closed for real target writes. That means it can inspect MongoDB, build plans, review and approve them, and verify backup metadata, but the actual MongoDB write execution is intentionally blocked until transaction-backed apply behavior is completed.

## How The App Works Today

The CLI entrypoint is:

```bash
go run ./cmd/mimic --help
```

The expected config file is YAML:

```bash
mimic.yml
```

JSON config files are rejected because JSON is reserved for generated artifacts such as plans and approved plans.

The source and target MongoDB URIs are read from environment variables configured in the YAML file:

```yaml
source:
  uriEnv: STAGING_MONGO_URI

target:
  uriEnv: PRODUCTION_MONGO_URI
```

The tool only considers collections explicitly listed under `collections`. It does not scan or promote every MongoDB collection.

## Functional Features

- `mimic validate --config=mimic.yml`
  - Loads YAML config files.
  - Rejects non-`.yml` and non-`.yaml` config files.
  - Requires source and target URI environment variables.
  - Requires at least one configured collection.
  - Requires every collection to define a stable key.
  - Rejects unsupported collection modes.
  - Rejects unsupported array strategies.
  - Rejects indexes configured for collections that are not allowlisted.
  - Connects to source and target MongoDB.
  - Verifies basic collection access.
  - Checks duplicate stable keys in source and target before planning or applying.

- `mimic diff --config=mimic.yml`
  - Reads source and target documents from configured collections.
  - Ignores configured fields such as `_id`, `createdAt`, and `updatedAt`.
  - Uses configured stable keys instead of `_id`.
  - Detects inserts and updates.
  - Detects missing configured indexes in the target.
  - Prints a human-readable diff.
  - Does not modify MongoDB.

- `mimic plan --config=mimic.yml --out=plans/plan.json`
  - Builds a raw JSON plan from the current source and target diff.
  - Writes deterministic plan operations.
  - Includes source label, target label, creation time, and config checksum.
  - Does not modify MongoDB.

- `mimic review --plan=plans/plan.json`
  - Reads a raw plan file.
  - Prints operation counts by collection.
  - Shows basic risk information.
  - Does not modify MongoDB.

- `mimic approve --plan=plans/plan.json --out=plans/plan.approved.json`
  - Creates an approved plan artifact.
  - Stores the original plan checksum.
  - Stores the config checksum from the raw plan.
  - Stores approved and skipped operations.
  - Supports approving only selected collections with `--collections=a,b,c`.
  - Writes an approved plan checksum.

- `mimic backup --config=mimic.yml --plan=plans/plan.approved.json --out=backups/run`
  - Requires an approved plan.
  - Verifies the approved plan config checksum against the current config.
  - Creates required backup directories:

```txt
backups/run/source
backups/run/target
```

  - Writes `backups/run/metadata.json`.
  - Stores source and target labels, database names, included collections, config checksum, and approved plan checksum.

- `mimic apply --plan=plans/plan.approved.json --backup=backups/run --confirm=<target>`
  - Refuses raw plans.
  - Requires an approved plan.
  - Verifies the approved plan checksum.
  - Requires backup metadata.
  - Verifies the backup metadata matches the approved plan checksum.
  - Requires the confirmation value to exactly match the approved plan target label.
  - Currently blocks real MongoDB writes until transaction-backed execution is implemented.

- `mimic export-script --plan=... --format=mongodb-js`
  - Command exists.
  - The exporter package exists.
  - Actual script generation is not implemented yet.

## Safety Features Already Enforced

- Dry-run behavior is preserved for `validate`, `diff`, `plan`, `review`, `approve`, and `backup`.
- No command modifies target data except the `apply` command path.
- The current `apply` command refuses to perform real writes until safe transaction execution exists.
- `_id` can be ignored through default config rules and is not required as a stable key.
- Stable keys are mandatory for all configured collections.
- Duplicate stable keys fail validation before planning.
- Apply does not recalculate the diff.
- Apply reads only the approved plan artifact.
- Apply verifies approved plan checksum before continuing.
- Apply requires backup metadata that matches the approved plan checksum.
- Index creation is detected in plans, but real index application is blocked for now.

## Known Gaps To Fix

- Real MongoDB write execution is not implemented.
- `apply` needs transaction-backed `insertOne` and `updateOne` execution.
- `apply` needs audit log writing for every attempted operation.
- `apply` needs before/after counts and operation result tracking.
- `apply` needs a tested rollback or restore strategy before enabling writes.
- `backup` currently creates local metadata and directories only. It does not run `mongodump` or Atlas snapshots yet.
- Backup artifact verification should be stricter once real dumps exist.
- `export-script` does not generate MongoDB JavaScript yet.
- Delete operations are not implemented.
- Drop index operations are not implemented.
- Reference mapping rules are modeled in config validation but not applied in diff generation.
- `mergeByKey` array strategy is accepted but currently behaves like order-preserving normalization.
- Unique index duplicate checks before index creation still need dedicated implementation.
- There are no integration tests against disposable MongoDB containers yet.
- Unit tests exist only for a small part of config validation.
- Docs and examples still need a final pass to remove old `mongo-promote` naming.

## Verification

The current codebase was verified with:

```bash
go test ./...
```

Result:

```txt
PASS
```

Packages without tests compiled successfully. Existing config tests passed.

## Recommended Next Work

1. Rename the example config from `examples/mongo-promote.yml` to `examples/mimic.yml`.
2. Update `docs/ARCHITECTURE.md` and `docs/TESTING.md` to use `mimic` naming.
3. Add focused tests for config file extension validation, index validation, array normalization, diff generation, plan approval, and checksum validation.
4. Implement `export-script --format=mongodb-js`.
5. Implement real backup through `mongodump` for local MongoDB.
6. Implement transaction-backed apply for `insertOne` and `updateOne`.
7. Add integration tests using the local Docker MongoDB setup.
