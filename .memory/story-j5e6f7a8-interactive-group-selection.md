---
id: j5e6f7a8
title: Interactive Group Selection
created_at: 2026-03-02T13:43:00+10:30
updated_at: 2026-03-02T13:43:00+10:30
status: proposed
epic_id: c5d7e9b1
priority: medium
story_points: 5
test_coverage: none
---

# Interactive Group Selection

## User Story

**As a** Jot user  
**I want to** be prompted to select a group when I don't specify `--type` or path  
**So that** I can still create notes quickly without memorizing types

## Acceptance Criteria

- [ ] Interactive prompt shown when: no `--type`, no path, multiple groups exist
- [ ] Groups displayed with name and description
- [ ] Arrow key navigation or number selection
- [ ] Selected group used for filename + template
- [ ] Can be skipped with `--no-interactive` flag
- [ ] Non-interactive fallback uses notebook default or errors

## Context

While power users will memorize type names, casual users benefit from guided selection. This story provides an interactive fallback that surfaces available groups when the user doesn't provide explicit type information.

## Out of Scope

- Fuzzy search/filtering in the selector
- Recently-used group ordering
- Group favorites or pinning

## Tasks

- [task-e1f2a3b4](task-e1f2a3b4-interactive-selection-ui.md) — Interactive group selection UI (Phase 3)
- [task-f2a3b4c5](task-f2a3b4c5-interactive-trigger-conditions.md) — Interactive prompt trigger conditions (Phase 3)
- [task-a3b4c5d6](task-a3b4c5d6-no-interactive-flag.md) — Add --no-interactive flag and fallback (Phase 3)

## Test Specification

### E2E Tests

| AC# | Criterion | Test file/case | Status |
|-----|-----------|----------------|--------|
| 1 | Interactive prompt shown when conditions met | interactive_test.go / TestPromptConditions | ⬜ |
| 2 | Groups displayed with name and description | interactive_test.go / TestGroupDisplay | ⬜ |
| 3 | Arrow key navigation or number selection | interactive_test.go / TestNavigation | ⬜ |
| 4 | Selected group used for filename + template | interactive_test.go / TestSelectionApplied | ⬜ |
| 5 | Can be skipped with `--no-interactive` flag | interactive_test.go / TestNoInteractiveFlag | ⬜ |
| 6 | Non-interactive fallback uses default or errors | interactive_test.go / TestNonInteractiveFallback | ⬜ |

### Unit Test Coverage (via Tasks)

_To be populated as tasks are completed_

## Notes

- Recommended library: `github.com/charmbracelet/huh` or `github.com/AlecAivazis/survey`
- Prompt trigger conditions:
  1. No `--type` flag provided
  2. No explicit path argument
  3. Notebook has more than one group defined
  4. TTY is available (stdin is terminal)
- Non-interactive mode (CI, pipes): use `default_group` from notebook config, or error if not set
- Consider `JOT_NO_INTERACTIVE=1` env var as alternative to flag
