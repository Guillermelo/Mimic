# Mimic

## Overview

Mimic is a review-first Go CLI that compares two MongoDB databases, shows the exact changes it would make, and applies only the operations explicitly approved by the operator.

The main use case is:

```txt
staging -> production
```

The tool must also work with any pair of MongoDB databases:

```txt
dev -> staging
staging -> production
backup -> production
client-a -> client-b
```

The tool must not assume that the entire source database is the desired truth. MongoDB does not know which documents are configuration, test data, production data, or temporary state. For that reason, every comparison, approval, and write must be controlled by an explicit YAML configuration file and an approved plan.

The tool will be developed in **Go**.

## Problem

Teams often make configuration changes in staging and later need to apply those changes to production without overwriting real production data.

Examples of data that may usually be promoted:

```txt
settings
roles
permissions
vehicleTypes
documentTemplates
formFields
languages
paymentGateways
billingPlans
subscriptionPlans
contentPages
base categories
```

Examples of data that should usually not be promoted:

```txt
users
orders
payments
walletTransactions
cards
trips
logs
notifications
reviews
chats
sessions
tokens
```

## Non-Goal

This tool must not be a button for:

```txt
make production equal to staging
```

That would be dangerous because staging can contain test data, incomplete documents, different IDs, temporary changes, and experimental records.

The tool must only operate with:

- explicitly allowed collections;
- stable keys per collection;
- ignored fields;
- array comparison rules;
- pre-apply validations;
- human review before apply;
- an approved immutable plan before apply;
- backups before writing;
- dry-run by default;
- explicit confirmation for real writes.

## Core Concepts

### Source

The database that contains the desired configurable changes. Usually staging.

### Target

The database that will receive the approved changes. Usually production.

### Backup

A backup or snapshot taken before any write operation.

Both source and target must be backed up before applying a plan. The target backup is required for recovery. The source backup is required for auditability and reproducibility of the exact data that was promoted.

### Plan

A generated JSON file containing proposed operations.

A raw plan is not enough to modify the target. The operator must review and approve the plan first.

Examples:

```txt
insert document
update document
create index
drop index, only when explicitly allowed
delete document, only when explicitly allowed
```

### Approved Plan

An approved copy of a plan that contains only the operations accepted by the operator.

`apply` must only execute an approved plan. It must not apply a raw plan directly, and it must not recalculate the diff during apply.

### Stable Key

A logical key used to identify the same document in both databases.

The tool must not use `_id` by default because the same business concept can have different ObjectIds across environments.

Examples:

```txt
settings.key
roles.slug
languages.code
vehicleTypes.code
documentTemplates.slug
paymentGateways.provider
```

A stable key can also be composite:

```txt
country + city + serviceType
```

## YAML Configuration File

Mimic configuration must be written in YAML.

Supported config extensions:

```txt
.yml
.yaml
```

JSON config files are not supported. JSON is reserved for generated machine artifacts such as plans, approved plans, and audit records.

Example `mimic.yml`:

```yaml
source:
  uriEnv: STAGING_MONGO_URI

target:
  uriEnv: PRODUCTION_MONGO_URI

defaults:
  dryRun: true
  allowDeletes: false
  ignoreFields:
    - _id
    - __v
    - createdAt
    - updatedAt

collections:
  settings:
    key:
      - key
    mode: upsert
    ignoreFields:
      - lastModifiedBy

  roles:
    key:
      - slug
    mode: upsert

  vehicleTypes:
    key:
      - code
    mode: upsert
    arrays:
      permissions:
        strategy: sort
      zones:
        strategy: preserveOrder

  documentTemplates:
    key:
      - slug
    mode: upsert

indexes:
  settings:
    - keys:
        key: 1
      options:
        unique: true
        name: key_unique

  roles:
    - keys:
        slug: 1
      options:
        unique: true
        name: slug_unique
```

## Expected Commands

### Validate Configuration

```bash
mimic validate --config=mimic.yml
```

This command must verify:

- the config file extension is `.yml` or `.yaml`;
- the required environment variables exist;
- the tool can connect to both databases;
- configured collections exist or can be created;
- every configured collection has a stable key;
- no rule is ambiguous;
- source and target do not point to the same database;
- unique stable keys do not produce duplicates in either database.

### Generate Diff

```bash
mimic diff --config=mimic.yml
```

This command must show differences without touching the target database.

Example output:

```txt
settings
  + insert key="dynamic_pricing_enabled"
  ~ update key="commission_percentage"
    value: 10 -> 12

roles
  + insert slug="dispatcher"

indexes
  + createIndex settings.key_unique
```

### Generate Plan

```bash
mimic plan --config=mimic.yml --out=plans/2026-07-29-staging-to-prod.json
```

This command must create a JSON file with proposed operations.

Example:

```json
{
  "source": "staging",
  "target": "production",
  "createdAt": "2026-07-29T00:00:00.000Z",
  "operations": [
    {
      "type": "updateOne",
      "collection": "settings",
      "filter": {
        "key": "commission_percentage"
      },
      "update": {
        "$set": {
          "value": 12
        }
      },
      "options": {
        "upsert": true
      }
    }
  ]
}
```

### Review Plan

```bash
mimic review --plan=plans/2026-07-29-staging-to-prod.json
```

This command must show the proposed operations in a human-readable format before anything is applied.

Example output:

```txt
Target: production
Source: staging

Collections:
  settings
    1 insert
    1 update
    0 deletes

  roles
    1 insert
    0 updates
    0 deletes

Indexes:
  settings
    1 createIndex

Risk checks:
  deletes: disabled
  unique indexes: validated
  backup required: yes

No changes have been applied.
```

### Approve Plan

```bash
mimic approve \
  --plan=plans/2026-07-29-staging-to-prod.json \
  --out=plans/2026-07-29-staging-to-prod.approved.json
```

This command creates an approved plan artifact.

The approval step should support removing operations from the plan before approval. For example, an operator may approve updates in `settings` but skip a proposed index change.

The approved plan must include:

```txt
original plan checksum
YAML config checksum
approved plan checksum
approved operations
skipped operations
approval timestamp
operator identity, when available
```

### Backup Databases

```bash
mimic backup \
  --config=mimic.yml \
  --plan=plans/2026-07-29-staging-to-prod.approved.json \
  --out=backups/2026-07-29-staging-to-prod
```

This command must create backups for both databases before any write is applied.

Required backup artifacts:

```txt
backups/2026-07-29-staging-to-prod/source
backups/2026-07-29-staging-to-prod/target
backups/2026-07-29-staging-to-prod/metadata.json
```

The metadata file should include:

```txt
source connection label
target connection label
database names
included collections
backup timestamp
tool version
config checksum
approved plan checksum
```

The backup command should use MongoDB-native backup mechanisms where possible, such as `mongodump` for self-managed MongoDB or Atlas snapshots when running in Atlas-backed environments.

### Apply Plan

```bash
mimic apply \
  --plan=plans/2026-07-29-staging-to-prod.approved.json \
  --backup=backups/2026-07-29-staging-to-prod \
  --confirm=production
```

This command must apply only the operations already present in the approved plan.

For safety, `apply` must only accept an approved plan and must not recalculate the diff. The sequence is:

```txt
generate diff
generate plan
review plan
approve plan
backup source and target
apply that exact approved plan
```

### Export Script

```bash
mimic export-script --plan=plans/2026-07-29-staging-to-prod.approved.json --format=mongodb-js
```

This command should generate a reviewable script from an approved plan that can be stored in Git when a project wants a versioned artifact outside the Go CLI.

## Recommended Workflow

The intended workflow is:

```txt
1. Make and test configuration changes in staging.
2. Run validate against source and target.
3. Run diff from source to target.
4. Generate a proposed plan file.
5. Review the plan manually.
6. Approve only the operations that should be applied.
7. Create backups of both source and target using the approved plan checksum.
8. Validate that the backup artifacts exist and match the YAML config and approved plan checksums.
9. Apply the exact approved plan to target with explicit confirmation.
10. Verify counts, indexes, and critical application behavior.
11. Store the original plan, approved plan, backup metadata, and audit log.
12. Optionally export a script or migration artifact for Git.
```

In production, the minimum safe command sequence should be:

```bash
mimic validate --config=mimic.yml
mimic diff --config=mimic.yml
mimic plan --config=mimic.yml --out=plans/2026-07-29-staging-to-prod.json
mimic review --plan=plans/2026-07-29-staging-to-prod.json
mimic approve --plan=plans/2026-07-29-staging-to-prod.json --out=plans/2026-07-29-staging-to-prod.approved.json
mimic backup --config=mimic.yml --plan=plans/2026-07-29-staging-to-prod.approved.json --out=backups/2026-07-29-staging-to-prod
mimic apply --plan=plans/2026-07-29-staging-to-prod.approved.json --backup=backups/2026-07-29-staging-to-prod --confirm=production
```

## Failure Handling Requirement

If an apply operation fails, Mimic must prevent partial target changes for every operation class it supports.

The tool must enforce this with multiple layers:

- use MongoDB transactions for data writes when the target deployment supports them;
- group compatible writes into transactional batches;
- avoid mixing non-transactional operations with transactional document updates in the same apply batch;
- run preflight validations before writing;
- require a target backup before apply;
- verify the target backup before apply;
- record every attempted operation and its result in an audit log;
- refuse to apply any approved plan that contains operations without a safe rollback or restore strategy.

For normal document writes, the preferred strategy is:

```txt
transaction starts
approved insert/update/delete operations run
transaction commits only if every operation succeeds
transaction aborts if any operation fails
```

For non-transactional or deployment-dependent operations, such as some index or collection changes, Mimic must either:

- run them in a separate apply phase with explicit approval;
- require a maintenance window or write pause;
- require a tested restore command from the target backup;
- or refuse to apply them.

Important limitation:

```txt
Restoring a full target backup can also remove writes made by the live application after the backup was created.
```

Because of that, full backup restore is only safe when the target is in a controlled maintenance window, or when the operator explicitly accepts that recovery model.

Mimic must fail closed. If it cannot keep the target unchanged after a failure, or cannot restore it safely under the current conditions, it must not start the apply.

## Interactive Confirmation

Before applying an approved plan, Mimic must show a final summary and require textual confirmation.

Example:

```txt
Target: production
Source: staging
Approved plan: plans/2026-07-29-staging-to-prod.approved.json
Backup: backups/2026-07-29-staging-to-prod

Operations:
  inserts: 2
  updates: 5
  deletes: 0
  indexes: 1

Safety:
  approved plan checksum: valid
  backup checksum: valid
  deletes: disabled
  rollback strategy: transaction for document writes
  non-transactional operations: index phase requires explicit confirmation

Type "apply production" to continue:
```

The command must abort unless the confirmation text matches exactly.

`apply` must never prompt with a weak yes/no confirmation for production writes.

## Important Restrictions

### Dry-Run by Default

No command should modify data unless `apply` is used with an approved plan, a verified backup, and explicit confirmation:

```bash
--confirm=production
```

### Approved Plan Required

`apply` must refuse raw plans.

The only valid input for `apply` is an approved plan generated by `mimic approve`.

The approved plan must be immutable during apply. Mimic should verify its checksum before writing.

### Mandatory Allowlist

The tool must not compare or modify every collection by default.

It may only operate on collections defined in the YAML configuration file.

### Deletes Disabled by Default

If a document exists in target but not in source, the tool may report it, but it must not delete it unless the config allows deletes.

Example:

```yaml
allowDeletes: true
```

Deletes should also be configurable per collection.

### Do Not Use `_id` by Default

`_id` must be ignored by default.

It may only be used as a stable key if explicitly configured.

### Unique Indexes Require Pre-Validation

Before creating a unique index, the tool must check for duplicates in the target database.

If duplicates exist, the plan must fail before any write is applied.

### References Between Collections

MongoDB does not enforce relationships like a SQL database. If a document references another collection, the tool must support rules to:

- keep the reference as-is;
- map the reference by a stable key;
- warn if the referenced document does not exist in target;
- block the operation if the reference is required.

### Arrays Need Explicit Strategy

Not all arrays should be compared the same way.

Possible strategies:

```txt
preserveOrder: order matters
sort: sort before comparing
set: compare as a set without duplicates
replace: replace the full array
mergeByKey: merge array objects using an internal key
```

### Runtime Fields Should Be Ignored

Fields like these are usually runtime state and should be ignored:

```txt
createdAt
updatedAt
lastLogin
lastSeen
usageCount
runtimeStats
temporaryToken
cache
```

### Audit Log

Every real apply execution must create an audit record.

The audit should include:

```txt
date
user or machine user
source
target
original plan checksum
approved plan checksum
source backup path or snapshot id
target backup path or snapshot id
insert count
update count
delete count
index count
attempted operations
applied operations
skipped operations
errors
rollback status
restore status, if any
```

The audit log can be stored as a local file and also in a target collection:

```txt
mimic_runs
```

## Comparison Modes

### Document-Level Diff

Compares full documents after normalizing ignored fields.

Best for small configuration documents.

### Field-Level Diff

Generates `$set` and `$unset` operations by field.

Best for avoiding full document replacement.

### Index Diff

Compares expected indexes from the config against existing target indexes.

The tool may create missing indexes, but it must not drop extra indexes unless explicitly configured.

## Supported Operations

Initial scope:

```txt
insertOne
updateOne with $set
updateOne with $unset
createIndex
```

Future scope:

```txt
deleteOne
dropIndex
rename field
transform field
map references
merge arrays by key
```

## Recommended Protections

- block execution if source and target point to the same URI/database;
- require `--confirm=production` for production writes;
- show counts before and after apply;
- support `--max-operations` to avoid accidental mass changes;
- support `--collections` to limit a run;
- fail if a configured collection has no stable key;
- fail if source or target has duplicated stable keys;
- never print database credentials in logs;
- save and approve a plan before apply;
- require backups before apply;
- verify backup metadata before apply;
- apply operations in a deterministic order;
- refuse operations that cannot be rolled back or restored safely.

## Suggested Go Project Structure

```txt
mimic/
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
      restore.go
      transactions.go
    diff/
      normalize.go
      documents.go
      collections.go
      indexes.go
    plan/
      operation.go
      build.go
      approve.go
      read.go
      write.go
      checksum.go
    apply/
      apply.go
      rollback.go
      validators.go
    audit/
      audit.go
    exporters/
      mongodb_js.go
  examples/
    mimic.yml
  go.mod
  go.sum
  README.md
```

## Final Definition

Mimic is a reusable Go CLI for comparing two MongoDB databases and promoting approved configurable data from one database to another.

It does not replace backups.
It does not replace versioned structural migrations.
It must not modify production data without explicit rules.

Its ideal role is:

```txt
promote configurable data and metadata between MongoDB environments in a safe, auditable, repeatable way.
```

For code-dependent structural changes, use versioned migrations.

For admin-made staging changes, use Mimic to generate a diff, create a proposed plan, review and approve selected operations, back up both databases, and apply the approved plan only when rollback or restore safety is available.
