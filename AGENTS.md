# Project Working Rules

These rules must be checked before writing or changing code in this repository.

## Language And Comments

- All source code, identifiers, commit messages, test names, code comments, and user-facing CLI text must be written in English.
- Documentation may be bilingual only when it helps project communication, but technical reference files should prefer English.
- Comments must explain why something exists, not repeat what the code already says.

## Safety Model

- The tool must never modify a target MongoDB unless the user explicitly requests an apply flow.
- Dry-run behavior is the default for every workflow that can affect data.
- The tool must only operate on collections listed in the configuration file.
- Deletes, index drops, destructive transformations, and rollback behavior must be opt-in and independently tested.
- Connection strings and credentials must never be printed in logs, plans, errors, or test output.
- Applying a plan must use the exact plan file provided by the user. It must not recalculate the diff during apply.

## Data Rules

- `_id` must be ignored by default and must not be used as a stable key unless explicitly configured.
- Every configured collection must define at least one stable key.
- Duplicate stable keys in source or target must fail validation before planning or applying.
- Unique indexes require duplicate checks before creation.
- Array comparison must be strategy-driven. Do not add implicit array behavior without a config rule.
- Runtime and calculated fields must be ignored through defaults or collection-level configuration.

## Go Code Standards

- Keep packages small and purpose-specific.
- Prefer explicit domain types over unstructured maps at package boundaries.
- Use `context.Context` for I/O, MongoDB calls, and command execution paths.
- Return errors with useful context and without leaking secrets.
- Avoid global mutable state.
- Prefer standard library packages unless a dependency has clear value.
- Keep MongoDB driver usage behind internal packages so CLI and planning logic remain testable.

## Testing Requirements

- Add unit tests for config validation, normalization, diff generation, plan building, and safety validators.
- Add integration tests only against local disposable MongoDB instances.
- Tests must not require real staging or production credentials.
- Test fixtures must be deterministic and safe to rerun.
- Before finishing a coding task, run the narrowest relevant test command. For shared behavior, run the full suite.

## Change Process

1. Read this file.
2. Read the relevant README or docs for the area being changed.
3. Inspect existing package patterns before adding new structure.
4. Make the smallest coherent change.
5. Add or update tests for the changed behavior.
6. Run formatting and tests.
7. Report what changed, what was verified, and any remaining risk.

