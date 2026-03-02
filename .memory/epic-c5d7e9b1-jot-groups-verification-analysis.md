---
id: c5d7e9b1
title: Jot Groups Verification Analysis
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: in-progress
---

# Jot Groups Verification Analysis

## Vision/Goal

Verify whether Jot groups can let users issue intent-level commands (e.g., "create a task titled X with content Y") while Jot enforces naming/path/metadata conventions automatically.

## Success Criteria

- Clear mapping from user intent → group selection → filename/path/frontmatter generation.
- Deterministic behavior with conflict handling and fallback rules.
- Documented limits and required explicit inputs when ambiguity exists.
- Recommendation on whether group rules alone are sufficient or need workflow augmentation.

## Dependencies

- Notebook group config and template handling.
- Path resolution and slug generation behavior.
- Proposed workflow model for stronger validation when needed.

## Stories

Stories define WHAT needs to be built (phase-agnostic requirements):

### Foundation Stories
- [story-g2b3c4d5](story-g2b3c4d5-unicode-safe-slugify.md) — Unicode-safe slugify (CJK, emoji, accents) [P0]
- [story-f1a2b3c4](story-f1a2b3c4-type-based-note-creation.md) — Type-based note creation (`--type task`) [P0]

### Schema & Resolution Stories
- [story-h3c4d5e6](story-h3c4d5e6-group-filename-patterns.md) — Group filename patterns (gomplate) [P1]
- [story-9c0d1e2a](story-9c0d1e2a-group-driven-note-creation-verification.md) — Group driven note creation verification [verification]
- [story-ad1e2f3b](story-ad1e2f3b-natural-language-create-task-intent-mapping.md) — Natural language create task intent mapping [P1]

### User Experience Stories
- [story-i4d5e6f7](story-i4d5e6f7-group-content-templates.md) — Group content templates (frontmatter) [P2]
- [story-j5e6f7a8](story-j5e6f7a8-interactive-group-selection.md) — Interactive group selection prompt [P2]

## Phases

Phases define WHEN work happens and group TASKS (scheduled work):

### Phase 1: Foundation
- **Status**: todo
- **Start Criteria**: Epic approved, research complete
- **End Criteria**: Unicode slugify and type-based note creation are implemented and tested
- **Implements Stories**: story-g2b3c4d5, story-f1a2b3c4
- **Tasks**:
  - [ ] [task-a1b2c3d4](task-a1b2c3d4-integrate-slug-library.md) — Integrate gosimple/slug library
  - [ ] [task-b2c3d4e5](task-b2c3d4e5-slug-truncation-logic.md) — Slug truncation logic
  - [ ] [task-c3d4e5f6](task-c3d4e5f6-add-type-flag.md) — Add --type flag to notes add
  - [ ] [task-d4e5f6a7](task-d4e5f6a7-type-to-group-resolver.md) — Type-to-group resolver
- **Notes**: Core building blocks that other features depend on. Slugify must be complete before filename patterns.

### Phase 2: Schema & Resolution
- **Status**: planned
- **Start Criteria**: Phase 1 complete
- **End Criteria**: Group filename patterns work with gomplate; intent-to-group resolution documented
- **Implements Stories**: story-h3c4d5e6, story-9c0d1e2a, story-ad1e2f3b
- **Tasks**:
  - [ ] [task-e5f6a7b8](task-e5f6a7b8-filename-format-schema.md) — Add filename_format to group schema
  - [ ] [task-f6a7b8c9](task-f6a7b8c9-gomplate-integration.md) — Gomplate template engine integration
  - [ ] [task-a7b8c9d0](task-a7b8c9d0-jot-namespace-funcs.md) — Implement jot namespace functions
  - [ ] [task-b8c9d0e1](task-b8c9d0e1-filename-collision-detection.md) — Filename collision detection
- **Notes**: Depends on gomplate research. May require schema changes to group config.

### Phase 3: User Experience
- **Status**: planned
- **Start Criteria**: Phase 2 complete
- **End Criteria**: Content templates and interactive selection available; full intent-to-note pipeline tested
- **Implements Stories**: story-i4d5e6f7, story-j5e6f7a8
- **Tasks**:
  - [ ] [task-c9d0e1f2](task-c9d0e1f2-content-template-processing.md) — Content template processing
  - [ ] [task-d0e1f2a3](task-d0e1f2a3-content-template-fallback.md) — Content template fallback and error handling
  - [ ] [task-e1f2a3b4](task-e1f2a3b4-interactive-selection-ui.md) — Interactive group selection UI
  - [ ] [task-f2a3b4c5](task-f2a3b4c5-interactive-trigger-conditions.md) — Interactive prompt trigger conditions
  - [ ] [task-a3b4c5d6](task-a3b4c5d6-no-interactive-flag.md) — Add --no-interactive flag and fallback
- **Notes**: Polish and UX improvements. Can be deferred if core functionality is sufficient.

## Related Research

- [research-c5d7e9b1](research-c5d7e9b1-gomplate-custom-functions.md) — Gomplate custom functions research
- [research-r1a2b3c4](research-r1a2b3c4-jot-groups-intent-resolution.md) — Groups intent resolution research
