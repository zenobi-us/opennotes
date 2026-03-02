---
id: f6a7b8c9
title: Gomplate Template Engine Integration
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: todo
epic_id: c5d7e9b1
phase_id: 2
story_id: h3c4d5e6
assigned_to: null
---

# Gomplate Template Engine Integration

## Objective

Integrate the gomplate library to process filename_format and template fields, setting up the rendering pipeline with custom function namespace.

## Related Story

- [story-h3c4d5e6](story-h3c4d5e6-group-filename-patterns.md) — Group Filename Patterns
- Directly implements AC#2 (Processed via gomplate with `jot` namespace functions)
- Directly implements AC#4 (Top-level aliases work: `{{ .title | slug }}`)

## Related Phase

- **Phase 2: Schema & Resolution** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Depends on task-e5f6a7b8 (schema field exists)
- Foundation for jot namespace functions (task-a7b8c9d0)

## Steps

1. Add gomplate library dependency:
   ```bash
   go get github.com/hairyhenderson/gomplate/v4
   ```

2. Create `internal/services/template.go` for template engine:
   ```go
   package services

   import (
       "github.com/hairyhenderson/gomplate/v4"
       "github.com/hairyhenderson/gomplate/v4/data"
   )

   type TemplateEngine struct {
       funcs template.FuncMap
   }

   func NewTemplateEngine() *TemplateEngine {
       return &TemplateEngine{
           funcs: make(template.FuncMap),
       }
   }
   ```

3. Implement render function:
   ```go
   func (te *TemplateEngine) Render(tmpl string, data map[string]interface{}) (string, error) {
       // Parse template
       // Apply custom funcs
       // Execute with data
       // Return result
   }
   ```

4. Register top-level aliases for convenience:
   ```go
   te.funcs["slug"] = Slug  // so {{ .title | slug }} works
   ```

5. Create filename generation function using template engine:
   ```go
   func (s *NoteService) GenerateFilename(group *GroupConfig, title string) (string, error) {
       format := group.FilenameFormat
       if format == "" {
           format = DefaultFilenameFormat
       }
       return s.templateEngine.Render(format, map[string]interface{}{
           "title": title,
       })
   }
   ```

## Unit Tests

- `TestTemplateEngine_BasicRender`: `{{ .title }}` → "my title" → supports AC#2
- `TestTemplateEngine_TopLevelSlug`: `{{ .title | slug }}` → "my-title" → supports AC#4
- `TestTemplateEngine_ChainedPipe`: `{{ .title | slug | upper }}` → "MY-TITLE" → supports AC#2
- `TestGenerateFilename_UsesGroupFormat`: group format used when set → supports AC#2

## Expected Outcome

A working template engine that can process gomplate templates with custom functions and aliases.

## Actual Outcome

_To be filled after completion_

## Lessons Learned

_To be filled after completion_
