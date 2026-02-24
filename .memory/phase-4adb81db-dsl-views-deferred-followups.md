---
id: 4adb81db
title: DSL Views Deferred Follow-up Phase
type: "phase"
created_at: 2026-02-20T19:20:00+10:30
updated_at: 2026-02-20T19:20:00+10:30
status: planning
epic_id: f661c068
start_criteria: Core DSL views implementation is stable and deferred follow-up tasks are approved for execution
end_criteria: All deferred follow-up tasks for view ergonomics and config hierarchy are completed and verified
---

# DSL Views Deferred Follow-up Phase

## Overview

This phase groups deferred items from `plan-b4e2f7a1-dsl-views-implementation.md` into executable tasks focused on usability and completeness of the views workflow.

## Deliverables

- CLI save/delete operations for custom views
- Runtime override flags for executing saved views
- OR syntax support in DSL grammar
- Named parameter substitution in view definitions
- Global views support in user config hierarchy

## Tasks

- [task-6c834006](task-6c834006-view-save-delete-cli-flags.md)
- [task-ddbaa84b](task-ddbaa84b-view-cli-override-flags.md)
- [task-a82526f2](task-a82526f2-dsl-or-syntax-support.md)
- [task-684a9a73](task-684a9a73-view-parameter-substitution.md)
- [task-e19963c7](task-e19963c7-global-views-config-support.md)

## Dependencies

- Existing DSL views core behavior from `plan-b4e2f7a1`
- Parser and Bleve translation stability for advanced query features
- Config service read/write path for notebook and global view definitions

## Next Steps

1. Confirm implementation order with human reviewer.
2. Execute tasks one-by-one with verification gates.
3. Update docs and examples after each feature lands.
