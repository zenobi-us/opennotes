---
id: b8c9d0e1
title: Filename Collision Detection
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: todo
epic_id: c5d7e9b1
phase_id: 2
story_id: h3c4d5e6
assigned_to: null
---

# Filename Collision Detection

## Objective

Implement collision detection that raises an error when a generated filename already exists, rather than overwriting or auto-suffixing.

## Related Story

- [story-h3c4d5e6](story-h3c4d5e6-group-filename-patterns.md) — Group Filename Patterns
- Directly implements AC#6 (Filename collisions detected and error raised)

## Related Phase

- **Phase 2: Schema & Resolution** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Depends on task-f6a7b8c9 (filename generation exists)

## Steps

1. Add collision check to note creation flow:
   ```go
   func (s *NoteService) CreateNote(group *GroupConfig, title string, content string) error {
       filename, err := s.GenerateFilename(group, title)
       if err != nil {
           return err
       }
       
       fullPath := filepath.Join(group.Directory, filename)
       
       // Check for collision
       if _, err := os.Stat(fullPath); err == nil {
           return &FilenameCollisionError{
               Path:     fullPath,
               Filename: filename,
               Suggestion: "Use a different title or add unique identifier to filename_format",
           }
       }
       
       // ... continue with creation
   }
   ```

2. Create custom error type for collisions:
   ```go
   type FilenameCollisionError struct {
       Path       string
       Filename   string
       Suggestion string
   }

   func (e *FilenameCollisionError) Error() string {
       return fmt.Sprintf("filename collision: %s already exists. %s", 
           e.Filename, e.Suggestion)
   }
   ```

3. Handle error appropriately in CLI:
   - Display clear error message
   - Suggest solutions (change title, or use NanoID in format)
   - Return appropriate exit code

4. Consider adding `--force` flag for future (out of scope for now)

## Unit Tests

- `TestCreateNote_CollisionDetected`: existing file blocks creation → supports AC#6
- `TestCreateNote_NoCollision`: new filename succeeds → supports AC#6
- `TestFilenameCollisionError_Message`: error includes helpful suggestion → supports AC#6

## Expected Outcome

Users get clear error messages when filename collisions occur, with suggestions for resolution.

## Actual Outcome

_To be filled after completion_

## Lessons Learned

_To be filled after completion_
