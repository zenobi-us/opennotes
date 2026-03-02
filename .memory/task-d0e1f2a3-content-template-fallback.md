---
id: d0e1f2a3
title: Content Template Fallback and Error Handling
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: done
epic_id: c5d7e9b1
phase_id: 3
story_id: i4d5e6f7
assigned_to: null
---

# Content Template Fallback and Error Handling

## Objective

Implement fallback behavior when no template is specified, and provide clear error reporting for template syntax errors.

## Related Story

- [story-i4d5e6f7](story-i4d5e6f7-group-content-templates.md) — Group Content Templates
- Directly implements AC#5 (Falls back to minimal template if none specified)
- Directly implements AC#6 (Template syntax errors reported clearly)

## Related Phase

- **Phase 3: User Experience** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Depends on task-c9d0e1f2 (content template processing)

## Steps

1. Implement fallback logic:
   ```go
   func (s *NoteService) getTemplate(group *GroupConfig) string {
       if group.Template != "" {
           return group.Template
       }
       return DefaultContentTemplate
   }
   ```

2. Create syntax error wrapper with context:
   ```go
   type TemplateSyntaxError struct {
       Template string
       Line     int
       Column   int
       Message  string
   }

   func (e *TemplateSyntaxError) Error() string {
       return fmt.Sprintf("template syntax error at line %d, col %d: %s\n\n%s",
           e.Line, e.Column, e.Message, e.highlightError())
   }
   ```

3. Parse template errors to extract location:
   ```go
   func parseTemplateError(err error, tmpl string) *TemplateSyntaxError {
       // Extract line/column from gomplate error
       // Highlight the problematic line in output
   }
   ```

4. Consider template validation at notebook load time:
   ```go
   func (s *NotebookService) ValidateGroupTemplates() []error {
       var errors []error
       for _, group := range s.notebook.Groups {
           if group.Template != "" {
               if err := s.templateEngine.Validate(group.Template); err != nil {
                   errors = append(errors, fmt.Errorf("group %s: %w", group.Name, err))
               }
           }
       }
       return errors
   }
   ```

## Unit Tests

- `TestGenerateContent_NoTemplate`: empty template uses default → supports AC#5
- `TestGenerateContent_SyntaxError`: `{{ .title | invalid }}` gives clear error → supports AC#6
- `TestTemplateSyntaxError_LineNumber`: error includes line info → supports AC#6
- `TestValidateGroupTemplates_CatchesErrors`: validation finds bad templates → supports AC#6

## Expected Outcome

Users get helpful error messages with line numbers when template syntax is wrong, and minimal templates when none specified.

## Actual Outcome

✅ Successfully implemented template fallback and error handling:

- Created `TemplateSyntaxError` custom error type in `errors.go`
- Added `Validate()` method to `TemplateEngine` for syntax checking
- Wrapped parse/execute errors in `Render()` with `TemplateSyntaxError`
- Added `ValidateGroupTemplates()` to validate all group templates at once
- Validates both `Template` and `FilenameFormat` fields
- Supports `errors.As()` for type assertions
- Fallback behavior uses existing `GetTemplate()` and `GetFilenameFormat()` methods
- Comprehensive tests added
- All 359 tests pass

## Lessons Learned

- Wrapping errors at the point of failure with context (template source, original error) makes debugging much easier
- Validating templates upfront (at config load) can catch errors before note creation fails
