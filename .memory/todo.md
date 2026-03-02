# Jot TODO

## Epic: Jot Workflows (`b2f4e6a8`) ✅ COMPLETE — [NEEDS-HUMAN] review before archival
- [x] All 4 stories, 15 tasks complete. Learnings distilled: [learning-w4k9f2m1](learning-w4k9f2m1-jot-workflows-epic-complete.md)
- [x] Epic: [epic-b2f4e6a8](epic-b2f4e6a8-jot-workflows.md)

## Epic: JSON Schema Generation (`7c631839`) — proposed
- [ ] **P0** [story-s0z9y8x7](story-s0z9y8x7-storage-runtime-type-separation.md) — Separate storage types from runtime types
- [ ] **P1** [story-s1a2b3c4](story-s1a2b3c4-schema-generation-tooling.md) — Automated schema generation from Go structs
- [ ] **P1** [story-s2b3c4d5](story-s2b3c4d5-schema-ci-drift-detection.md) — CI schema drift detection
- [ ] **P2** [story-s3c4d5e6](story-s3c4d5e6-notebook-init-schema-copy.md) — Notebook init includes schema reference
- [ ] **P3** [story-s4d5e6f7](story-s4d5e6f7-notebook-validate-command.md) — Notebook config validation command

## Epic: Jot Groups Verification Analysis (`c5d7e9b1`) — IMPLEMENTATION COMPLETE

Epic: [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
Review: [research-f40dde4c](research-f40dde4c-codebase-alignment-review.md)

**[NEEDS-HUMAN] Confirm artifact status updates below**

### Phase 1: Foundation — ✅ IMPLEMENTED (code verified 2026-03-03)

| Task | Story | Description | Status |
|------|-------|-------------|--------|
| [task-a1b2c3d4](task-a1b2c3d4-integrate-slug-library.md) | g2b3c4d5 | Integrate gosimple/slug library | ✅ `internal/services/slug.go` |
| [task-b2c3d4e5](task-b2c3d4e5-slug-truncation-logic.md) | g2b3c4d5 | Slug truncation logic | ✅ `SlugWithMax()` |
| [task-c3d4e5f6](task-c3d4e5f6-add-type-flag.md) | f1a2b3c4 | Add --type flag to notes add | ✅ `cmd/notes_add.go:70` |
| [task-d4e5f6a7](task-d4e5f6a7-type-to-group-resolver.md) | f1a2b3c4 | Type-to-group resolver | ✅ implemented |

### Phase 2: Schema & Resolution — ✅ IMPLEMENTED (code verified 2026-03-03)

| Task | Story | Description | Status |
|------|-------|-------------|--------|
| [task-e5f6a7b8](task-e5f6a7b8-filename-format-schema.md) | h3c4d5e6 | Add filename_format to group schema | ✅ `NotebookGroup.FilenameFormat` |
| [task-f6a7b8c9](task-f6a7b8c9-gomplate-integration.md) | h3c4d5e6 | Template engine integration | ✅ Go text/template (not gomplate) |
| [task-a7b8c9d0](task-a7b8c9d0-jot-namespace-funcs.md) | h3c4d5e6 | Implement jot namespace functions | ✅ `JotNamespace` in template_funcs.go |
| [task-b8c9d0e1](task-b8c9d0e1-filename-collision-detection.md) | h3c4d5e6 | Filename collision detection | ✅ `CheckFilenameCollision()` |

### Phase 3: User Experience — ✅ IMPLEMENTED (code verified 2026-03-03)

| Task | Story | Description | Status |
|------|-------|-------------|--------|
| [task-c9d0e1f2](task-c9d0e1f2-content-template-processing.md) | i4d5e6f7 | Content template processing | ✅ `GenerateContent()` |
| [task-d0e1f2a3](task-d0e1f2a3-content-template-fallback.md) | i4d5e6f7 | Content template fallback and error handling | ✅ `DefaultContentTemplate` |
| [task-e1f2a3b4](task-e1f2a3b4-interactive-selection-ui.md) | j5e6f7a8 | Interactive group selection UI | ✅ `SelectGroupInteractively()` |
| [task-f2a3b4c5](task-f2a3b4c5-interactive-trigger-conditions.md) | j5e6f7a8 | Interactive prompt trigger conditions | ✅ `ShouldShowInteractiveSelector()` |
| [task-a3b4c5d6](task-a3b4c5d6-no-interactive-flag.md) | j5e6f7a8 | Add --no-interactive flag and fallback | ✅ Flag + env var |

### Pending Artifact Updates

After human review confirmation:
1. [ ] Update 7 story files: `status: proposed` → `status: completed`, `test_coverage: none` → `test_coverage: partial`
2. [ ] Update epic c5d7e9b1: `status: in-progress` → `status: completed`
3. [ ] Create distilled learnings artifact (like learning-w4k9f2m1 for workflows)

### Research
- [x] [research-c5d7e9b1](research-c5d7e9b1-gomplate-custom-functions.md) — Gomplate custom functions research
- [ ] [research-r1a2b3c4](research-r1a2b3c4-jot-groups-intent-resolution.md) — Groups intent resolution research

## Proposed / Parked Epics
- [ ] [epic-6e1f2a9c](epic-6e1f2a9c-cli-config-normalization-layer.md) — CLI Config Normalization Layer (future)

## Parked
- [ ] [task-9c4a2f8d](task-9c4a2f8d-github-actions-moonrepo-releases.md) — GitHub Actions CI/CD
- [ ] [plan-b4e2f7a1](plan-b4e2f7a1-dsl-views-implementation.md) — DSL views implementation (10 tasks)
