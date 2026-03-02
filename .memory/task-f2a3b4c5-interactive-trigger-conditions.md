---
id: f2a3b4c5
title: Interactive Prompt Trigger Conditions
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: todo
epic_id: c5d7e9b1
phase_id: 3
story_id: j5e6f7a8
assigned_to: null
---

# Interactive Prompt Trigger Conditions

## Objective

Implement the logic that determines when to show the interactive group selector: no type, no path, multiple groups, and TTY available.

## Related Story

- [story-j5e6f7a8](story-j5e6f7a8-interactive-group-selection.md) — Interactive Group Selection
- Directly implements AC#1 (Interactive prompt shown when: no `--type`, no path, multiple groups exist)

## Related Phase

- **Phase 3: User Experience** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Depends on task-e1f2a3b4 (interactive UI exists)

## Steps

1. Create condition checker:
   ```go
   type InteractiveContext struct {
       TypeFlag       string
       ExplicitPath   string
       Groups         []GroupConfig
       IsTTY          bool
       NoInteractive  bool
   }

   func ShouldShowInteractiveSelector(ctx InteractiveContext) bool {
       // Check all conditions
       if ctx.NoInteractive {
           return false
       }
       if ctx.TypeFlag != "" {
           return false  // Type explicitly specified
       }
       if ctx.ExplicitPath != "" {
           return false  // Path explicitly specified
       }
       if len(ctx.Groups) <= 1 {
           return false  // Only one or no groups - no choice to make
       }
       if !ctx.IsTTY {
           return false  // Not a terminal - can't show interactive UI
       }
       return true
   }
   ```

2. Add TTY detection:
   ```go
   func IsTTY() bool {
       return term.IsTerminal(int(os.Stdin.Fd()))
   }
   ```

3. Wire into notes add command:
   ```go
   if ShouldShowInteractiveSelector(ctx) {
       group, err = SelectGroupInteractively(notebook.Groups)
   } else {
       // Use default or error
   }
   ```

4. Document conditions clearly in help text

## Unit Tests

- `TestShouldShowInteractive_NoType_NoPath_MultiGroups_TTY`: returns true → supports AC#1
- `TestShouldShowInteractive_TypeProvided`: returns false → supports AC#1
- `TestShouldShowInteractive_PathProvided`: returns false → supports AC#1
- `TestShouldShowInteractive_SingleGroup`: returns false → supports AC#1
- `TestShouldShowInteractive_NotTTY`: returns false → supports AC#1

## Expected Outcome

Interactive selector only appears when appropriate conditions are met.

## Actual Outcome

_To be filled after completion_

## Lessons Learned

_To be filled after completion_
