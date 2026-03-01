---
id: 7c631839
title: JSON Schema Generation from Go Structs
created_at: 2026-03-01T01:45:00+10:30
updated_at: 2026-03-01T12:20:00+10:30
status: proposed
---

# JSON Schema Generation from Go Structs

## Background

The `jot.schema.json` file was manually created in commit `3e8fcbf` to provide IDE validation for `.jot.json` notebook configuration files. The commit message listed unfulfilled "Next steps" including automated schema generation, but this was never implemented.

Currently:
- Schema is hand-written and can drift from actual Go struct definitions
- No automated way to regenerate schema when structs change
- Schema validation is not enforced at runtime
- **Storage and runtime types are not cleanly separated** (see Architecture section)

## Architecture: Storage vs Runtime Types

See [knowledge-config-transformation.md](knowledge-config-transformation.md) for detailed state machine.

**Current State:**
```
.jot.json  ──▶  StoredNotebookConfig  ──▶  NotebookConfig  ──▶  Notebook
(storage)       (storage type ✓)          (runtime type)       (full runtime)
                                          embeds storage       has services
```

**Key Insight**: Schema generation must target **storage types only**, not runtime types.

| Layer | Types | Schema Target? |
|-------|-------|----------------|
| **Storage** | `StoredNotebookConfig`, `NotebookGroup`, `WorkflowDefinition` | ✅ YES |
| **Runtime** | `NotebookConfig`, `Notebook`, `NoteService` | ❌ NO |

**Gap Identified**: Global config (`Config` in `config.go`) lacks storage/runtime separation. Both use the same struct.

## Vision/Goal

1. **Cleanly separate storage types** into a dedicated package for schema generation
2. **Auto-generate `jot.schema.json`** from storage types only
3. **Prevent drift** via CI enforcement
4. **Improve DX** with init schema reference and validate command

## Success Criteria

1. **Clear Type Boundaries**: Storage types are in a dedicated location, separate from runtime
2. **Single Source of Truth**: Go storage structs are canonical; schema is derived
3. **Reproducible Generation**: `mise run schema:generate` produces deterministic output
4. **CI Enforcement**: Schema drift is detected before merge
5. **IDE Support**: Generated schema provides autocomplete in VSCode/IntelliJ
6. **Optional Runtime Validation**: `jot notebook validate` checks config against schema

## Phases

1. **Phase 0: Type Separation** — Extract storage types to clean boundary
2. **Phase 1: Research & Tooling** — Evaluate Go JSON schema libraries
3. **Phase 2: Generation Implementation** — Implement mise task
4. **Phase 3: CI Integration** — Add schema drift detection
5. **Phase 4: UX Improvements** — Init schema copy, validate command

## Stories

| Priority | Story | Description |
|----------|-------|-------------|
| P0 | [story-s0z9y8x7](story-s0z9y8x7-storage-runtime-type-separation.md) | Separate storage types from runtime types |
| P1 | [story-s1a2b3c4](story-s1a2b3c4-schema-generation-tooling.md) | Automated schema generation from Go structs |
| P1 | [story-s2b3c4d5](story-s2b3c4d5-schema-ci-drift-detection.md) | CI schema drift detection |
| P2 | [story-s3c4d5e6](story-s3c4d5e6-notebook-init-schema-copy.md) | Notebook init includes schema reference |
| P3 | [story-s4d5e6f7](story-s4d5e6f7-notebook-validate-command.md) | Notebook config validation command |

## Storage Types Inventory

Types that should be schema generation targets:

```go
// Notebook config storage (.jot.json)
StoredNotebookConfig     // internal/services/notebook.go
NotebookGroup            // internal/services/notebook.go
NotebookGroupWorkflow    // internal/services/notebook.go
WorkflowDefinition       // internal/services/workflow_validation.go
WorkflowState            // internal/services/workflow_validation.go

// Global config storage (~/.config/jot/config.json)
Config                   // internal/services/config.go (⚠️ needs StoredGlobalConfig)
```

## Dependencies

- Requires selection of pure-Go JSON schema library (e.g., `invopop/jsonschema`)
- Must preserve existing schema features: examples, descriptions, patterns
- Must handle `map[string]any` fields (metadata, additionalProperties)
- Type separation should not break existing tests or API

## Research References

- Git history: schema manually created in `3e8fcbf`, generation planned but not implemented
- Branch `feat/json-schema-generation` contains only manual schema
- Architecture analysis: [knowledge-config-transformation.md](knowledge-config-transformation.md)

## Open Questions

1. **Package location**: `internal/schema/` or `internal/services/schema/` or `internal/config/`?
2. **Embed schema in binary** for runtime validation portability?
3. **Global config schema**: Should we also generate schema for `~/.config/jot/config.json`?
4. **Handling `map[string]any`**: Use `additionalProperties: true` or define allowed keys?
5. **jsonschema tags**: Use `jsonschema:"description=..."` or generate from doc comments?
