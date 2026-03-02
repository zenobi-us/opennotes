---
id: i4d5e6f7
title: Group Content Templates
created_at: 2026-03-02T13:43:00+10:30
updated_at: 2026-03-02T13:43:00+10:30
status: proposed
epic_id: c5d7e9b1
priority: medium
story_points: 5
test_coverage: none
---

# Group Content Templates

## User Story

**As a** Jot user  
**I want** groups to define content templates with frontmatter  
**So that** new notes start with the right metadata structure

## Acceptance Criteria

- [ ] Group's `template` field processed via gomplate
- [ ] Template can include frontmatter with dynamic values
- [ ] Same `jot` namespace functions available
- [ ] `.title` and other inputs accessible in template
- [ ] Falls back to minimal template if none specified
- [ ] Template syntax errors reported clearly

## Context

When creating notes of a specific type, users expect consistent structure. A "task" note should have status, priority, and due date fields. A "meeting" note should have attendees and action items sections. This story enables groups to define content templates that are automatically applied when notes are created.

## Out of Scope

- Template inheritance (child templates extending parent)
- External template file references (templates must be inline in config)
- Post-creation template updates (only applies at creation time)

## Tasks

_To be populated during task breakdown_

## Test Specification

### E2E Tests

| AC# | Criterion | Test file/case | Status |
|-----|-----------|----------------|--------|
| 1 | Group's `template` field processed via gomplate | template_test.go / TestTemplateProcessing | ⬜ |
| 2 | Template can include frontmatter | template_test.go / TestFrontmatterInTemplate | ⬜ |
| 3 | Same jot namespace functions available | template_test.go / TestJotFunctionsInTemplate | ⬜ |
| 4 | `.title` and inputs accessible | template_test.go / TestTemplateInputs | ⬜ |
| 5 | Falls back to minimal template | template_test.go / TestMinimalFallback | ⬜ |
| 6 | Template syntax errors reported clearly | template_test.go / TestSyntaxErrorReporting | ⬜ |

### Unit Test Coverage (via Tasks)

_To be populated as tasks are completed_

## Notes

- Template context variables:
  - `.title` - The note title from user input
  - `.filename` - Generated filename (after pattern resolution)
  - `.created_at` - Creation timestamp
  - `.group` - Group name/metadata
- Default minimal template:
  ```markdown
  ---
  title: {{ .title }}
  created_at: {{ jot.Now "2006-01-02T15:04:05Z07:00" }}
  ---
  
  # {{ .title }}
  ```
- Syntax errors should include line number and context from template
- Consider validating templates at notebook load time (not just on use)
