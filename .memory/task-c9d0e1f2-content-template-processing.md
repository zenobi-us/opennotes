---
id: c9d0e1f2
title: Content Template Processing
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: done
epic_id: c5d7e9b1
phase_id: 3
story_id: i4d5e6f7
assigned_to: null
---

# Content Template Processing

## Objective

Implement processing of group `template` field via gomplate to generate initial note content with dynamic frontmatter.

## Related Story

- [story-i4d5e6f7](story-i4d5e6f7-group-content-templates.md) — Group Content Templates
- Directly implements AC#1 (Group's `template` field processed via gomplate)
- Directly implements AC#2 (Template can include frontmatter with dynamic values)
- Directly implements AC#3 (Same `jot` namespace functions available)
- Directly implements AC#4 (`.title` and other inputs accessible in template)

## Related Phase

- **Phase 3: User Experience** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Depends on Phase 2 completion (gomplate + jot funcs exist)

## Steps

1. Add `template` field to group config if not present:
   ```go
   type GroupConfig struct {
       // existing fields...
       Template string `json:"template,omitempty"`
   }
   ```

2. Create content generation function:
   ```go
   func (s *NoteService) GenerateContent(group *GroupConfig, data NoteData) (string, error) {
       tmpl := group.Template
       if tmpl == "" {
           tmpl = DefaultContentTemplate
       }
       
       return s.templateEngine.Render(tmpl, map[string]interface{}{
           "title":      data.Title,
           "filename":   data.Filename,
           "created_at": data.CreatedAt,
           "group":      group.Name,
       })
   }
   ```

3. Define default minimal template:
   ```go
   const DefaultContentTemplate = `---
   title: {{ .title }}
   created_at: {{ jot.Now "2006-01-02T15:04:05Z07:00" }}
   ---

   # {{ .title }}
   `
   ```

4. Wire content generation into note creation flow:
   ```go
   content, err := s.GenerateContent(group, noteData)
   if err != nil {
       return fmt.Errorf("template error: %w", err)
   }
   ```

5. Ensure template context includes all documented variables

## Unit Tests

- `TestGenerateContent_BasicTemplate`: simple template renders → supports AC#1
- `TestGenerateContent_Frontmatter`: YAML frontmatter included → supports AC#2
- `TestGenerateContent_JotFuncs`: `{{ jot.Now ... }}` works in template → supports AC#3
- `TestGenerateContent_TitleAccess`: `{{ .title }}` accessible → supports AC#4
- `TestGenerateContent_FilenameAccess`: `{{ .filename }}` accessible → supports AC#4

## Expected Outcome

Notes are created with rendered content from group templates, including proper frontmatter.

## Actual Outcome

✅ Successfully implemented content template processing:

- `Template` field already existed in `NotebookGroup` (no schema change needed)
- Added `DefaultContentTemplate` constant with YAML frontmatter and `jot.Now`
- Added `GetTemplate()` method with fallback to default
- Added `GenerateContent()` function in template.go
- Integrated into `cmd/notes_add.go` - uses group template when `--type` provided
- Template data includes: `title`, `filename`, `group`, plus custom `--data` fields
- All jot functions work in content templates
- Added tests in template_test.go and notebook_test.go
- All 361 tests pass

## Lessons Learned

- The Template field was already in the schema - always check existing code before assuming new fields are needed
- Content templates benefit from the same template engine used for filenames - code reuse pays off
