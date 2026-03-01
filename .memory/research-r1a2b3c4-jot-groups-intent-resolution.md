---
id: r1a2b3c4
title: Jot Groups Intent Resolution Research
created_at: 2026-03-01T18:20:15+10:30
updated_at: 2026-03-01T19:16:00+10:30
status: in-progress
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

---

## User Feedback Direction (2026-03-01)

User provided new direction:
1. Template/metadata should only be used for **creation** (not matching)
2. Metadata field may be unnecessary — can be frontmatter in template using gotemplate `{{ variable | semantics }}`
3. **Need deeper research** on reverse lookup: how to support "create a task" → system suggests path
4. Consider **gomplate** for advanced templating (more functions, filters)

---

## New Findings: Gomplate as Templating Solution

### What is Gomplate?

**gomplate** is a powerful Go template renderer that extends `text/template` with 200+ functions across 20+ namespaces. It's designed for rendering config files, documentation, and any text-based output from templates.

**Key Features Over text/template**:

| Feature | text/template | gomplate |
|---------|---------------|----------|
| **Functions** | ~15 built-in | 200+ across namespaces |
| **Date/Time** | Manual formatting | `time.Now`, `time.Parse`, timezone-aware |
| **Strings** | Basic | `strings.Slug`, `strings.KebabCase`, `strings.CamelCase`, `strings.Trunc` |
| **UUID** | None | `uuid.V4`, `uuid.V5` |
| **Random** | None | `random.String`, `random.Number` |
| **Crypto** | None | `crypto.SHA256`, HMAC, bcrypt |
| **Datasources** | None | JSON, YAML, env vars, vault, consul |
| **Math** | None | `math.Add`, `math.Ceil`, etc. |
| **Regex** | None | `regexp.Find`, `regexp.Replace`, `regexp.Split` |

### Integration with Go

gomplate is a pure Go library (MIT license) that can be used as a dependency:

```go
import "github.com/hairyhenderson/gomplate/v5"

// Simple rendering
renderer := gomplate.NewRenderer(gomplate.RenderOptions{})
out, err := renderer.RenderTemplates(ctx, []gomplate.Template{
    {Name: "note", Text: `---
title: {{ .title }}
created: {{ time.Now | time.Format "2006-01-02T15:04:05Z07:00" }}
slug: {{ .title | strings.Slug }}
---`},
})
```

**Integration Complexity**: Low — it's designed for embedding. The `Renderer` API accepts templates as strings and returns rendered output.

### Semantic Functions Useful for Note Creation

**For Filenames/Paths**:
- `strings.Slug` — Convert title to URL-safe slug (perfect for filenames)
- `strings.KebabCase` — "Sprint Planning" → "sprint-planning"
- `strings.SnakeCase` — "Sprint Planning" → "sprint_planning"
- `strings.CamelCase` — "sprint planning" → "sprintPlanning"
- `strings.Trunc` — Truncate long titles for filename limits

**For Dates**:
- `time.Now` — Current timestamp
- `time.Format` — Any format: `{{ time.Now | time.Format "2006-01-02" }}`
- `time.Parse` — Parse user input dates
- `time.RFC3339` — ISO 8601 format constant

**For IDs/Uniqueness**:
- `uuid.V4` — Generate unique IDs
- `random.String 8` — Random alphanumeric strings (for hashids)

**For Data Processing**:
- `data.ToYAML` — Convert map to YAML (frontmatter generation)
- `data.ToJSON` — Convert to JSON
- `conv.ToString` — Type coercion

### Example: Template-Driven Note Creation

```yaml
# .jot.json template config
templates:
  task:
    content: |
      ---
      id: {{ random.String 8 }}
      title: {{ .title }}
      type: task
      created: {{ time.Now | time.Format "2006-01-02T15:04:05Z07:00" }}
      status: todo
      {{ if .assignee }}assignee: {{ .assignee }}{{ end }}
      ---

      # {{ .title }}

      ## Objective

      ## Steps
      - [ ] 

      ## Expected Outcome
    path: ".memory/task-{{ random.String 8 }}-{{ .title | strings.Slug }}.md"
```

This eliminates the need for separate `Metadata` field — everything is frontmatter in the template.

### Recommendation

**Use gomplate** for template rendering. It provides:
1. Clean separation: templates define structure, groups define location rules
2. Rich functions without custom code
3. Path generation in templates (not just content)
4. Pure Go, no CGO, MIT license

---

## New Findings: Reverse Lookup / Intent Resolution

### The Problem

User says "create a task" — how does the system know:
1. Which group applies?
2. What path to use?
3. What template to apply?

Current system: User must specify path → system finds matching group (post-hoc)
Desired system: User specifies intent → system resolves path (proactive)

### How Other Tools Solve This

#### Obsidian Templater Plugin

**Folder Templates**:
- Configure: "For folder `/tasks/`, always use template `task.md`"
- Works on **new file creation** trigger
- Deepest folder match wins (specificity)
- Limitation: User must navigate to folder first

**File Regex Templates**:
- Configure: regex `^daily/.*` → use `daily-note.md` template
- Tests against file path
- First match wins (order matters)
- Limitation: Still requires user to know path

**Key Insight**: Obsidian solves path→template, not intent→path

#### Notion Databases

- User clicks "+ New" in a specific database view
- Database **knows its type** (task, meeting, etc.)
- Template is bound to database, not path
- No ambiguity: user action implies context

**Key Insight**: Selection is explicit (click database), not inferred

### Proposed Solutions for Jot

#### Option 1: Explicit Type Flag (`--type`)

```bash
jot notes add "Sprint Planning" --type task
# System looks up group with type="task", uses its path pattern + template
```

**Group Config**:
```json
{
  "name": "tasks",
  "type": "task",                          // NEW: type alias
  "globs": [".memory/task-*.md"],
  "template": "task",
  "path_pattern": ".memory/task-{hash}-{slug}.md"  // NEW: path generator
}
```

**Pros**: Explicit, no ambiguity
**Cons**: User must know available types

#### Option 2: Aliases on Groups

```json
{
  "name": "tasks",
  "aliases": ["task", "todo", "action"],   // NEW: fuzzy matching
  "globs": [".memory/task-*.md"],
  "template": "task"
}
```

```bash
jot notes add "Sprint Planning" --as todo
# System fuzzy-matches "todo" → "tasks" group
```

**Pros**: Flexible naming, memorable shortcuts
**Cons**: Potential collisions, must define aliases

#### Option 3: Interactive Selection (No Path Given)

```bash
jot notes add "Sprint Planning"
# No path → prompt user
? Select note type:
  > task (.memory/task-*.md)
    story (.memory/story-*.md)
    research (.memory/research-*.md)
```

**Pros**: Discoverable, no memorization
**Cons**: Extra interaction, slower

#### Option 4: Keyword Detection in Title (Risky)

```bash
jot notes add "Task: Sprint Planning"
# Detects "Task:" prefix → infers task group
```

**Pros**: Natural language feel
**Cons**: Fragile, language-dependent, false positives

### UX Flow Comparison

| Scenario | Option 1 (--type) | Option 2 (aliases) | Option 3 (interactive) |
|----------|-------------------|-------------------|------------------------|
| Power user | `--type task` | `--as t` | Skip prompt with flag |
| Discovery | `jot groups list-types` | `jot groups list --aliases` | Automatic |
| Ambiguity | Error if unknown type | Error if no match | User selects |
| Scriptable | ✅ Yes | ✅ Yes | ❌ Requires flag |

### Recommended Approach: Hybrid (Options 1 + 3)

1. **Add `--type` / `--as` flag** to `notes add`:
   - Maps to group by `type` field or `aliases` array
   - Group provides `path_pattern` (template variables for path generation)
   
2. **Interactive fallback** when no path AND no type given:
   - List groups that have templates defined
   - User selects → path generated from pattern
   
3. **Path pattern in group config** (NEW field):
   ```json
   "path_pattern": ".memory/task-{{ random 8 }}-{{ slug .title }}.md"
   ```
   - Gomplate renders this to generate actual path
   - Groups become "note type definitions"

### Data Flow with Reverse Lookup

```
User: jot notes add "Sprint Planning" --type task
                    ↓
1. Parse → title="Sprint Planning", type="task"
                    ↓
2. ResolveGroupByType("task")
    - Iterate groups, find first with type="task" or aliases contains "task"
    - Returns group: {template: "task", path_pattern: ".memory/task-{hash}-{slug}.md"}
                    ↓
3. RenderPathPattern(group.path_pattern, {title: "Sprint Planning"})
    - gomplate renders → ".memory/task-a8f3b2c1-sprint-planning.md"
                    ↓
4. RenderTemplate(group.template, {title: "Sprint Planning", ...})
    - gomplate renders template content
                    ↓
5. Write file to generated path
```

### Schema Changes Required

```go
type NotebookGroup struct {
    Name        string         `json:"name"`
    Type        string         `json:"type,omitempty"`        // NEW: canonical type name
    Aliases     []string       `json:"aliases,omitempty"`     // NEW: alternative names
    Globs       []string       `json:"globs"`                 // Keep: for post-hoc matching
    PathPattern string         `json:"path_pattern,omitempty"` // NEW: gomplate template for path
    Template    string         `json:"template,omitempty"`    // Keep: content template ref
    WorkflowID  string         `json:"workflow_id,omitempty"`
    // REMOVE: Metadata — handled by template frontmatter instead
}
```

---

## Updated Recommendations

### Simplification: Remove Metadata Field

User is correct — `Metadata` field on groups is redundant. Templates can define frontmatter directly with gomplate variables. Remove from schema.

### Template-Only for Creation

Groups serve two purposes:
1. **Matching** (globs): Post-hoc association for existing notes
2. **Creation** (template, path_pattern, type): Proactive generation for new notes

Keep these separate. `Globs` are for matching. `PathPattern` + `Template` are for creation.

### Next Steps

1. Add `type` and `aliases` fields to `NotebookGroup`
2. Add `path_pattern` field (gomplate template string)
3. Remove `Metadata` field
4. Integrate gomplate as template renderer
5. Implement `ResolveGroupByType()` function
6. Add `--type` flag to `notes add`
7. Add interactive selection fallback

---

## References (New)

- [gomplate documentation](https://docs.gomplate.ca/)
- [gomplate strings functions](https://docs.gomplate.ca/functions/strings/) — includes `Slug`, `KebabCase`
- [gomplate time functions](https://docs.gomplate.ca/functions/time/) — includes `Now`, `Format`
- [gomplate uuid functions](https://docs.gomplate.ca/functions/uuid/)
- [gomplate Go library](https://pkg.go.dev/github.com/hairyhenderson/gomplate/v5)
- [Obsidian Templater settings](https://silentvoid13.github.io/Templater/settings.html) — folder templates, regex templates
