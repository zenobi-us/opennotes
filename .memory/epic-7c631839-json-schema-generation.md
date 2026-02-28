---
id: 7c631839
title: JSON Schema Generation from Go Structs
created_at: 2026-03-01T01:45:00+10:30
updated_at: 2026-03-01T01:45:00+10:30
status: proposed
---

# JSON Schema Generation from Go Structs

## Background

The `jot.schema.json` file was manually created in commit `3e8fcbf` to provide IDE validation for `.jot.json` notebook configuration files. The commit message listed unfulfilled "Next steps" including automated schema generation, but this was never implemented.

Currently:
- Schema is hand-written and can drift from actual Go struct definitions
- No automated way to regenerate schema when structs change
- Schema validation is not enforced at runtime

Source of truth Go structs in `internal/services/notebook.go`:
- `StoredNotebookConfig` — root configuration
- `NotebookGroup` — group definitions with globs and metadata
- `NotebookGroupWorkflow` — workflow definitions (new in workflows epic)

## Vision/Goal

Implement automated JSON schema generation from Go structs to ensure the schema file always reflects the actual configuration structure. The schema should be regeneratable via a simple command and optionally validated at runtime.

## Success Criteria

1. **Single Source of Truth**: Go structs are the canonical definition; schema is derived
2. **Reproducible Generation**: Running `mise run schema:generate` produces identical output
3. **CI Integration**: Schema drift is detected in CI (generated vs committed mismatch)
4. **IDE Support Preserved**: Generated schema provides equivalent or better IDE autocomplete
5. **Runtime Validation (Optional)**: `jot notebook validate` command checks config against schema

## Phases

1. **Phase 1: Research & Tooling Selection** — evaluate Go JSON schema libraries
2. **Phase 2: Generation Implementation** — implement `go generate` or mise task
3. **Phase 3: CI Integration** — add schema drift detection to GitHub Actions
4. **Phase 4: Runtime Validation** — optional `jot notebook validate` command

## Stories

- [story-s1a2b3c4](story-s1a2b3c4-schema-generation-tooling.md) — As a maintainer, I want automated schema generation
- [story-s2b3c4d5](story-s2b3c4d5-schema-ci-drift-detection.md) — As a maintainer, I want CI to catch schema drift
- [story-s3c4d5e6](story-s3c4d5e6-notebook-init-schema-copy.md) — As a user, I want `jot init` to include the schema file
- [story-s4d5e6f7](story-s4d5e6f7-notebook-validate-command.md) — As a user, I want to validate my config against the schema

## Dependencies

- Requires selection of a pure-Go JSON schema generation library (e.g., `invopop/jsonschema`)
- Must preserve existing schema features: examples, descriptions, patterns
- Must handle custom types like `map[string]any` gracefully

## Research References

- Git history analysis: schema was manually created, generation was planned but not implemented
- Relevant commit: `3e8fcbf` (feat(schema): add JSON schema for notebook configuration)
- Branch `feat/json-schema-generation` contains only manual schema, no generation code

## Open Questions

1. Should we embed the schema in the binary for runtime validation?
2. Should `jot init` copy the schema file to the notebook directory or reference a URL?
3. How do we handle `map[string]any` fields (metadata, additionalProperties)?
