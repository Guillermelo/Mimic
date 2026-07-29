# Testing Environment

This repository uses disposable local MongoDB containers for integration testing.

## Start MongoDB Test Databases

```bash
docker compose up -d
```

This starts:

```txt
source MongoDB: mongodb://localhost:27018/mimic_source
target MongoDB: mongodb://localhost:27019/mimic_target
```

## Configure CLI Environment

PowerShell:

```powershell
$env:STAGING_MONGO_URI = "mongodb://localhost:27018/mimic_source"
$env:PRODUCTION_MONGO_URI = "mongodb://localhost:27019/mimic_target"
```

Bash:

```bash
export STAGING_MONGO_URI="mongodb://localhost:27018/mimic_source"
export PRODUCTION_MONGO_URI="mongodb://localhost:27019/mimic_target"
```

## Local Config

Use:

```bash
examples/mimic.yml
```

The seeded source and target databases intentionally differ:

- `settings.commission_percentage` differs between source and target.
- `settings.dynamic_pricing_enabled` exists only in source.
- `roles.dispatcher` exists only in source.
- Runtime fields differ and should be ignored by default.

## Stop And Reset

```bash
docker compose down -v
```

The `-v` flag removes container volumes so the next startup reseeds both databases.
