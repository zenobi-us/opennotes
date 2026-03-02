---
id: c3d4e5f6
title: Add --type Flag to Notes Add
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: todo
epic_id: c5d7e9b1
phase_id: 1
story_id: f1a2b3c4
assigned_to: null
---

# Add --type Flag to Notes Add

## Objective

Add a `--type` flag to the `jot notes add` command that accepts a type name (e.g., "task", "meeting") for intent-based note creation.

## Related Story

- [story-f1a2b3c4](story-f1a2b3c4-type-based-note-creation.md) — Type-based Note Creation
- Directly implements AC#1 (`--type` flag added to `notes add` command)
- Supports AC#5 (Works with existing `--notebook` flag)

## Related Phase

- **Phase 1: Foundation** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Foundation for type-based workflow

## Steps

1. Locate `cmd/notes_add.go` and add flag definition:
   ```go
   var noteType string
   notesAddCmd.Flags().StringVarP(&noteType, "type", "t", "", "Note type (maps to group)")
   ```

2. Update command help text to document the flag:
   ```
   Flags:
     -t, --type string   Create note using the specified type (e.g., task, meeting)
   ```

3. Pass type to the note creation flow (for task-d4e5f6a7 to use):
   ```go
   if noteType != "" {
       // Type resolution will happen in task-d4e5f6a7
   }
   ```

4. Ensure `--type` and `--notebook` flags can be used together

5. Ensure `--type` and explicit path are mutually exclusive (or type takes precedence)

## Unit Tests

- `TestNotesAdd_TypeFlagExists`: flag is registered → supports AC#1
- `TestNotesAdd_TypeWithNotebook`: `--type task --notebook work` parses correctly → supports AC#5
- `TestNotesAdd_TypeShortFlag`: `-t task` equivalent to `--type task` → supports AC#1

## Expected Outcome

The `--type` flag is available on `jot notes add`, ready to be wired to group resolution.

## Actual Outcome

_To be filled after completion_

## Lessons Learned

_To be filled after completion_
