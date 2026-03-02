---
id: e1f2a3b4
title: Interactive Group Selection UI
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: done
epic_id: c5d7e9b1
phase_id: 3
story_id: j5e6f7a8
assigned_to: null
---

# Interactive Group Selection UI

## Objective

Implement an interactive prompt using charmbracelet/huh that displays available groups and allows user selection via keyboard.

## Related Story

- [story-j5e6f7a8](story-j5e6f7a8-interactive-group-selection.md) — Interactive Group Selection
- Directly implements AC#2 (Groups displayed with name and description)
- Directly implements AC#3 (Arrow key navigation or number selection)
- Directly implements AC#4 (Selected group used for filename + template)

## Related Phase

- **Phase 3: User Experience** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Depends on Phase 2 completion (type resolution exists)

## Steps

1. Add charmbracelet/huh dependency:
   ```bash
   go get github.com/charmbracelet/huh
   ```

2. Create group selector function:
   ```go
   func (s *NoteService) SelectGroupInteractively(groups []GroupConfig) (*GroupConfig, error) {
       options := make([]huh.Option[*GroupConfig], len(groups))
       for i, g := range groups {
           label := g.Name
           if g.Description != "" {
               label = fmt.Sprintf("%s - %s", g.Name, g.Description)
           }
           options[i] = huh.NewOption(label, &groups[i])
       }

       var selected *GroupConfig
       form := huh.NewForm(
           huh.NewGroup(
               huh.NewSelect[*GroupConfig]().
                   Title("Select a note type:").
                   Options(options...).
                   Value(&selected),
           ),
       )

       if err := form.Run(); err != nil {
           return nil, err
       }
       return selected, nil
   }
   ```

3. Style the selector appropriately:
   - Group name in bold
   - Description in muted color
   - Clear selection indicator

4. Wire selected group into note creation:
   ```go
   group, err := s.SelectGroupInteractively(notebook.Groups)
   if err != nil {
       return err
   }
   // Use group for filename + template generation
   ```

## Unit Tests

- `TestSelectGroupInteractively_OptionsBuilt`: groups converted to options → supports AC#2
- `TestSelectGroupInteractively_DescriptionIncluded`: description shown → supports AC#2
- (Integration tests needed for actual UI interaction → supports AC#3, AC#4)

## Expected Outcome

Users see a beautiful interactive selector when group selection is needed.

## Actual Outcome

✅ Successfully implemented interactive group selection UI:

- Added `github.com/charmbracelet/huh v0.8.0` dependency
- Created `internal/services/prompt.go` with:
  - `SelectGroupInteractively()` - main interactive selector using huh forms
  - `BuildGroupSelectOptions()` - converts groups to huh options
  - `BuildGroupLabel()` - creates display labels (Name + Type if different)
- Created `internal/services/prompt_test.go` with tests for option building
- Note: `NotebookGroup` has `Name` and `Type` but no `Description` - adapted label format
- All 359 tests pass

## Lessons Learned

- charmbracelet/huh provides beautiful TUI forms with minimal code
- Always check the actual struct fields rather than assuming - `NotebookGroup` had `Type` not `Description`
- Focus unit tests on option building logic since full TUI interaction is hard to automate
