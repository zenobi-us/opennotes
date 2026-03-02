---
id: e5f6a7b8
title: Add filename_format to Group Schema
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: done
epic_id: c5d7e9b1
phase_id: 2
story_id: h3c4d5e6
assigned_to: null
---

# Add filename_format to Group Schema

## Objective

Extend the group configuration schema to support a `filename_format` field that defines how filenames are generated using gomplate templates.

## Related Story

- [story-h3c4d5e6](story-h3c4d5e6-group-filename-patterns.md) — Group Filename Patterns
- Directly implements AC#1 (`filename_format` field added to group schema)
- Supports AC#5 (Fallback to `{{ slug .title }}.md` when no format specified)

## Related Phase

- **Phase 2: Schema & Resolution** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Depends on Phase 1 completion (slugify must exist)

## Steps

1. Add `filename_format` field to group config struct:
   ```go
   type GroupConfig struct {
       // existing fields...
       FilenameFormat string `json:"filename_format,omitempty"`
   }
   ```

2. Update JSON schema (`jot.schema.json`) if it exists:
   ```json
   "filename_format": {
       "type": "string",
       "description": "Gomplate template for generating filenames",
       "examples": [
           "task-{{ jot.NanoID 8 }}-{{ .title | slug }}.md",
           "{{ jot.Now \"2006-01-02\" }}-{{ .title | slug }}.md"
       ]
   }
   ```

3. Add schema validation for filename_format:
   - Must end with `.md`
   - Should not contain path separators
   - Basic template syntax validation (deferred to gomplate)

4. Define default format constant:
   ```go
   const DefaultFilenameFormat = "{{ .title | slug }}.md"
   ```

5. Add example to notebook config documentation

## Unit Tests

- `TestGroupConfig_FilenameFormatField`: field parses from JSON → supports AC#1
- `TestGroupConfig_FilenameFormatDefault`: empty field uses default → supports AC#5
- `TestGroupConfig_FilenameFormatValidation`: invalid formats rejected → supports AC#1

## Expected Outcome

Groups can define custom filename patterns in their config that will be processed by gomplate.

## Actual Outcome

✅ Successfully added `filename_format` to group schema:

- Added `FilenameFormat` field to `NotebookGroup` and `notebookGroupRaw` structs
- Defined `DefaultFilenameFormat` constant: `{{ .title | slug }}.md`
- Added `GetFilenameFormat()` helper with fallback to default
- Added `ValidateFilenameFormat()` validation (must end with `.md`, no path separators)
- Updated `jot.schema.json` with pattern validation and examples
- Added 12 comprehensive tests
- All 362 tests pass

## Lessons Learned

- Validation at schema level (must end with `.md`, no path separators) catches configuration errors early before template processing
