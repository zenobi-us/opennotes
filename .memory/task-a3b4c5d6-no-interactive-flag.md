---
id: a3b4c5d6
title: Add --no-interactive Flag and Fallback
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: todo
epic_id: c5d7e9b1
phase_id: 3
story_id: j5e6f7a8
assigned_to: null
---

# Add --no-interactive Flag and Fallback

## Objective

Add a `--no-interactive` flag and implement fallback behavior that uses notebook default or errors when interactive mode is disabled.

## Related Story

- [story-j5e6f7a8](story-j5e6f7a8-interactive-group-selection.md) — Interactive Group Selection
- Directly implements AC#5 (Can be skipped with `--no-interactive` flag)
- Directly implements AC#6 (Non-interactive fallback uses notebook default or errors)

## Related Phase

- **Phase 3: User Experience** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Depends on task-f2a3b4c5 (trigger conditions)

## Steps

1. Add `--no-interactive` flag to notes add command:
   ```go
   var noInteractive bool
   notesAddCmd.Flags().BoolVar(&noInteractive, "no-interactive", false, 
       "Disable interactive prompts (use default_group or error)")
   ```

2. Add environment variable support:
   ```go
   func init() {
       if os.Getenv("JOT_NO_INTERACTIVE") == "1" {
           noInteractive = true
       }
   }
   ```

3. Add `default_group` field to notebook config:
   ```go
   type NotebookConfig struct {
       // existing fields...
       DefaultGroup string `json:"default_group,omitempty"`
   }
   ```

4. Implement non-interactive fallback:
   ```go
   func (s *NoteService) GetGroupNonInteractive(notebook *NotebookConfig) (*GroupConfig, error) {
       if notebook.DefaultGroup != "" {
           group := s.FindGroupByName(notebook.DefaultGroup)
           if group != nil {
               return group, nil
           }
       }
       return nil, fmt.Errorf(
           "no group specified and interactive mode disabled. " +
           "Use --type flag or set default_group in notebook config")
   }
   ```

5. Wire fallback into command flow:
   ```go
   if !ShouldShowInteractiveSelector(ctx) {
       group, err = s.GetGroupNonInteractive(notebook)
       if err != nil {
           return err
       }
   }
   ```

## Unit Tests

- `TestNotesAdd_NoInteractiveFlag`: flag disables prompt → supports AC#5
- `TestNotesAdd_EnvVarNoInteractive`: `JOT_NO_INTERACTIVE=1` disables prompt → supports AC#5
- `TestGetGroupNonInteractive_UsesDefault`: default_group used → supports AC#6
- `TestGetGroupNonInteractive_ErrorsWithoutDefault`: no default → helpful error → supports AC#6

## Expected Outcome

CI pipelines and scripts can use `--no-interactive` with predictable fallback behavior.

## Actual Outcome

_To be filled after completion_

## Lessons Learned

_To be filled after completion_
