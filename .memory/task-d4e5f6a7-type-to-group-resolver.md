---
id: d4e5f6a7
title: Type-to-Group Resolver
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: done
epic_id: c5d7e9b1
phase_id: 1
story_id: f1a2b3c4
assigned_to: null
---

# Type-to-Group Resolver

## Objective

Implement the resolution logic that maps a user-provided type name to a specific group configuration, including lookup via `type` field and `aliases` array.

## Related Story

- [story-f1a2b3c4](story-f1a2b3c4-type-based-note-creation.md) — Type-based Note Creation
- Directly implements AC#2 (Type maps to group via `type` or `aliases` field)
- Directly implements AC#3 (Note created in group's directory)
- Directly implements AC#4 (Error with helpful message if type not found)

## Related Phase

- **Phase 1: Foundation** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Depends on task-c3d4e5f6 (type flag exists)

## Steps

1. Add `type` and `aliases` fields to group schema (if not present):
   ```go
   type GroupConfig struct {
       // existing fields...
       Type    string   `json:"type,omitempty"`
       Aliases []string `json:"aliases,omitempty"`
   }
   ```

2. Create resolver function in `internal/services/group.go`:
   ```go
   func (s *NotebookService) ResolveGroupByType(typeName string) (*GroupConfig, error) {
       typeName = strings.ToLower(typeName) // case-insensitive
       
       for _, group := range s.notebook.Groups {
           // Check exact type match
           if strings.ToLower(group.Type) == typeName {
               return &group, nil
           }
           // Check aliases
           for _, alias := range group.Aliases {
               if strings.ToLower(alias) == typeName {
                   return &group, nil
               }
           }
       }
       
       return nil, fmt.Errorf("type %q not found. Available types: %s", 
           typeName, s.ListAvailableTypes())
   }
   ```

3. Implement `ListAvailableTypes()` for error messages:
   ```go
   func (s *NotebookService) ListAvailableTypes() string {
       // Returns comma-separated list of all types and aliases
   }
   ```

4. Wire resolver into `notes add` command flow

5. Use resolved group's directory for note creation path

## Unit Tests

- `TestResolveGroupByType_ExactMatch`: "task" matches group with `type: task` → supports AC#2
- `TestResolveGroupByType_AliasMatch`: "todo" matches group with `aliases: [todo, t]` → supports AC#2
- `TestResolveGroupByType_CaseInsensitive`: "TASK" matches "task" → supports AC#2
- `TestResolveGroupByType_NotFound`: "unknown" returns error with available types → supports AC#4
- `TestResolveGroupByType_UsesGroupDir`: resolved group's directory used → supports AC#3

## Expected Outcome

Users can run `jot notes add --type task "My Task"` and have the note created in the correct group directory automatically.

## Actual Outcome

✅ Successfully implemented type-to-group resolver:

- Added `Type` and `Aliases` fields to `NotebookGroup` struct
- Implemented `ResolveGroupByType()` with case-insensitive matching (type field takes precedence over aliases)
- Implemented `ListAvailableTypes()` for helpful error messages
- Implemented `GetGroupDirectory()` to extract directory from group globs
- Wired into `cmd/notes_add.go` - resolves `--type` flag and uses group's directory
- Added 18 comprehensive tests covering all scenarios
- All 362 tests pass

## Lessons Learned

- Type field should take precedence over aliases when both match to ensure explicit mappings are honored
- Extracting directory from glob patterns requires careful handling of `**/*.md` and nested paths
