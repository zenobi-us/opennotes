---
id: r1a2b3c4
title: Jot Groups Intent Resolution Research
created_at: 2026-03-01T18:20:15+10:30
updated_at: 2026-03-01T18:27:00+10:30
status: completed
epic_id: c5d7e9b1
---

# Jot Groups Intent Resolution Research

## Research Questions

1. **Group Selection Mechanics**: How does Jot currently resolve which group applies to a note creation request?
2. **Template + Path Generation**: Can group config auto-generate filename/path/frontmatter from minimal user input?
3. **Conflict Handling**: What happens when multiple groups could match? Is there deterministic priority?
4. **Required vs Inferred Inputs**: What must the user explicitly provide vs what can be inferred from group rules?
5. **Edge Cases**: Slug collisions, missing required fields, ambiguous intent — how are these handled today?
6. **Workflow Integration**: How do groups interact with the recently completed workflow system (`epic-b2f4e6a8`)?

## Summary

Groups in Jot use **glob-based path matching** to determine applicability. Currently, group resolution is **post-hoc** (evaluated after the path is determined) rather than **intent-driven** (helping determine the path). The system has **explicit conflict detection** for workflows but uses **first-match semantics** for resolution. Templates exist as a concept but are **not yet integrated with groups** for auto-generation. The workflow integration is **mature and deterministic**.

## Findings

### 1. Group Configuration Analysis

**Schema** (`internal/services/notebook.go:21-27`):
```go
type NotebookGroup struct {
    Name       string         `json:"name"`
    Globs      []string       `json:"globs"`
    Metadata   map[string]any `json:"metadata"`
    Template   string         `json:"template,omitempty"`
    WorkflowID string         `json:"workflow_id,omitempty"`
}
```

**Key Observations**:
- Groups are ordered arrays in `.jot.json`
- Each group has a `Name`, `Globs` (path patterns), optional `Metadata`, optional `Template`, and optional `WorkflowID`
- Templates are **string references** to entries in `templates` map (not inline content)
- The `Template` field exists on groups but is **not currently used** in note creation flow

**Evidence**: Real config in `/.jot.json` shows 9 groups with glob patterns like `**/epic-*.md`, `**/story-*.md`, etc.

### 2. Path Resolution Behavior

**Path Resolution Function** (`internal/services/note.go:331-346`):
```go
func ResolvePath(notebookRoot, inputPath, slugifiedTitle string) string {
    // Case 1: No path specified - use root + slugified title
    if inputPath == "" {
        return filepath.Join(notebookRoot, slugifiedTitle+".md")
    }
    // Case 2: Ends with "/" - explicit folder
    if strings.HasSuffix(inputPath, "/") {
        return filepath.Join(notebookRoot, inputPath, slugifiedTitle+".md")
    }
    // Case 3: Full filepath with .md extension
    if strings.HasSuffix(inputPath, ".md") {
        return filepath.Join(notebookRoot, inputPath)
    }
    // Case 4: Filepath without extension - auto-add .md
    return filepath.Join(notebookRoot, inputPath+".md")
}
```

**Key Finding**: **Path is determined BEFORE group matching**. Groups are consulted for:
1. Workflow enforcement (via `enforceWorkflowForCreate`)
2. NOT for path/filename generation

**Gap Identified**: Groups have a `Template` field but it is **never read** in the note creation flow (`cmd/notes_add.go`). The template flag (`--template`) only references the notebook-level `templates` map.

### 3. Frontmatter Generation

**Current Implementation** (`cmd/notes_add.go:168-182`):
```go
func generateFrontmatter(title string, customData map[string]interface{}) string {
    frontmatter := make(map[string]interface{})
    if title != "" {
        frontmatter["title"] = title
    }
    frontmatter["created"] = time.Now().Format(time.RFC3339)
    
    // Merge custom data from --data flags
    for k, v := range customData {
        frontmatter[k] = v
    }
    ...
}
```

**Key Finding**: 
- Frontmatter is generated from title + `--data` flags only
- **Group metadata is NOT injected** into frontmatter automatically
- The `Metadata` field on groups exists but is **not used** for note creation

### 4. Conflict Resolution

**Workflow Assignment Resolution** (`internal/services/workflow_assignment.go:18-95`):

The system has **explicit conflict detection**:

```go
if len(workflowIDs) > 1 {
    // CONFLICT: Multiple groups match with DIFFERENT workflows
    return WorkflowAssignmentResult{
        Resolved: false,
        Diagnostics: []WorkflowDiagnostic{{
            Code:    "workflow.assignment_conflict",
            Message: fmt.Sprintf("conflicting workflow assignments for note path %s: %s", ...),
        }},
    }
}
```

**Resolution Rules**:
1. All matching groups' `WorkflowID`s are collected
2. If multiple **distinct** workflow IDs match → **CONFLICT ERROR** (blocks operation)
3. If all matching groups have the **same** workflow ID → **FIRST GROUP selected** (deterministic)
4. If no groups match → allowed (no workflow enforcement)

**Glob Matching** (`internal/services/semantic_search.go:407-417`):
```go
func globMatch(pattern, value string) bool {
    re := regexp.QuoteMeta(pattern)
    re = strings.ReplaceAll(re, `\*\*`, `.*`)      // ** = any path
    re = strings.ReplaceAll(re, `\*`, `[^/]*`)    // * = single segment
    re = strings.ReplaceAll(re, `\?`, `.`)        // ? = single char
    ...
}
```

### 5. Integration Points (Groups ↔ Workflows)

**Flow for Note Creation** (`cmd/workflow_enforce.go:12-44`):

```
User: jot notes add "Sprint Planning" meetings/
                    ↓
1. Parse arguments → title="Sprint Planning", path="meetings/"
                    ↓
2. ResolvePath() → notePath = "meetings/sprint-planning.md"
                    ↓
3. enforceWorkflowForCreate(nb, notePath, metadata)
        ↓
    3a. ResolveWorkflowAssignment(relPath, groups, workflows)
        - Iterate ALL groups
        - For each group, check if ANY glob matches relPath
        - Collect matching groups with workflow_id
        - Detect conflicts (multiple distinct workflow IDs)
        ↓
    3b. EnforceWorkflowOnMutation(...)
        - Get workflow definition
        - Validate initial state matches workflow.initial_state
        - Check metadata satisfies state schema
                    ↓
4. If blocked → return error with diagnostic
   If allowed → proceed to write file
```

**Key Integration Points**:
- Groups are the **binding layer** between file paths and workflows
- Workflow enforcement happens **before file creation** (fail-fast)
- State transitions are validated against workflow definition

### 6. Edge Cases

| Scenario | Current Behavior | Evidence |
|----------|------------------|----------|
| **Slug collision** | Error: "note already exists: {path}" | `cmd/notes_add.go:91` |
| **Empty title after slugify** | Error: "title produces empty filename after slugification" | `cmd/notes_add.go:74` |
| **No title, no path** | Auto-generate timestamp: `2026-03-01-182700.md` | `cmd/notes_add.go:83-85` |
| **No matching group** | Allowed (no workflow enforcement) | `workflow_lifecycle.go:40-45` |
| **Multiple groups, same workflow** | First match wins, allowed | `workflow_assignment.go` |
| **Multiple groups, different workflows** | Conflict error, operation blocked | `workflow_assignment.go:54-66` |
| **Missing required fields in schema** | Blocked with diagnostic | Schema validation in `EvaluateWorkflow` |
| **Invalid initial state** | Blocked: "transition X -> Y not allowed" | `workflow_lifecycle.go:97-112` |

## References

- Codebase: `internal/services/notebook.go` — group config loading, NotebookGroup struct
- Codebase: `internal/services/note.go` — ResolvePath function, ParseDataFlags
- Codebase: `internal/services/workflow_assignment.go` — ResolveWorkflowAssignment, conflict detection
- Codebase: `internal/services/workflow_lifecycle.go` — EnforceWorkflowOnMutation
- Codebase: `cmd/notes_add.go` — note creation flow, generateFrontmatter
- Codebase: `cmd/workflow_enforce.go` — enforceWorkflowForCreate wrapper
- Codebase: `internal/services/semantic_search.go` — globMatch implementation
- Config: `/.jot.json` — real-world group configuration with 9 groups and 6 workflows
- Epic: [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Related stories:
  - [story-9c0d1e2a](story-9c0d1e2a-group-driven-note-creation-verification.md)
  - [story-ad1e2f3b](story-ad1e2f3b-natural-language-create-task-intent-mapping.md)

## Gaps Identified for Future Work

1. **Group.Template unused**: Field exists but note creation ignores it
2. **Group.Metadata not injected**: Groups can declare metadata but it's not auto-applied to notes
3. **No intent-driven path inference**: User must specify path; groups don't suggest paths
4. **No reverse resolution**: Can't ask "which group should handle a task note?" and get path suggestion
5. **Template placeholders limited**: Only `{{title}}` supported, no date/group/workflow placeholders
