---
id: r1a2b3c4
title: Jot Groups Intent Resolution Research
created_at: 2026-03-01T18:20:15+10:30
updated_at: 2026-03-02T09:30:00+10:30
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

---

## PART 1: Existing Research (Groups & Workflow Integration)

[See previous sections from original research file - preserved below]

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

### 3. Workflow Assignment Resolution

**Resolution Rules** (`internal/services/workflow_assignment.go`):
1. All matching groups' `WorkflowID`s are collected
2. If multiple **distinct** workflow IDs match → **CONFLICT ERROR** (blocks operation)
3. If all matching groups have the **same** workflow ID → **FIRST GROUP selected** (deterministic)
4. If no groups match → allowed (no workflow enforcement)

---

## PART 2: Filename Generation Patterns Research (NEW)

### Research Questions (User Directed)

1. **Filename generation patterns** - how do other tools derive filename from title?
2. **Schema options for filename control** - what fields could we add?
3. **Fallback behavior** - when group has no filename config
4. **Edge cases** - collision handling, unicode/emoji, long titles

---

## Findings: How Other Tools Generate Filenames

### Tool Comparison Matrix

| Tool | Filename Strategy | ID Format | Date Format | Slug Style | Collision Handling |
|------|-------------------|-----------|-------------|------------|-------------------|
| **Obsidian** | Title as filename | None (opt-in Zettel ID via plugin) | None by default | Preserve case, spaces allowed | Append " 1", " 2" etc. |
| **Dendron** | Hierarchical + date/time | None | `y.MM.dd` (Luxon) for journal, `y.MM.dd.HHmmss` for scratch | Dot-separated hierarchy | Error on conflict |
| **Foam** | Title as filename | None | None | Preserve case | User must rename |
| **Zettlr** | Zettelkasten ID (timestamp) | `YYYYMMDDHHmmss` | Embedded in ID | ID-based, title optional | Unique by timestamp |
| **Bear** | Title as display name | Internal UUID | None visible | N/A (database) | N/A (database) |
| **Logseq** | Date-based pages or title | None | `YYYY-MM-DD` for journals | Preserve spaces | Merge pages |
| **Notable** | Title as filename | None | None | Preserve case | Append number |

### Pattern Categories

#### 1. **Pure Title → Slug** (Obsidian, Foam, Notable)
- User provides title, system converts to filename
- Simple, intuitive for users
- Problem: Non-ASCII characters, very long titles, collisions

**Obsidian example**:
```
Title: "Meeting Notes: Q1 2026"
Filename: "Meeting Notes- Q1 2026.md"  (preserves most chars, replaces : with -)
```

#### 2. **Date Prefix + Slug** (Dendron scratch/journal, Hugo)
- `YYYY-MM-DD-title-slug.md` or `YYYYMMDD-title.md`
- Provides chronological sorting
- Reduces collisions (date scopes)

**Dendron scratch note**:
```
dateFormat: "y.MM.dd.HHmmss"
Result: "scratch.2026.03.02.093015.md"
```

#### 3. **ID Prefix + Slug** (Zettelkasten, Zettlr)
- Unique ID guarantees no collision
- ID types: timestamp (`YYYYMMDDHHmmss`), UUID, nanoid
- Title becomes metadata, not filename

**Zettlr example**:
```
ID: 20260302093015
Filename: "20260302093015 Meeting Notes.md"
```

#### 4. **Type Prefix + Hash + Slug** (Current .memory/ convention)
- Pattern: `{type}-{8char_hash}-{slug}.md`
- Example: `task-a8f3b2c1-sprint-planning.md`
- Provides: type identification, uniqueness, readability

### Slug Style Comparison

| Style | Example | Pros | Cons |
|-------|---------|------|------|
| **kebab-case** | `sprint-planning` | URL-safe, readable, common | Multi-word separation only |
| **snake_case** | `sprint_planning` | Python/Ruby conventions | Less URL-friendly |
| **camelCase** | `sprintPlanning` | No separators, compact | Harder to read |
| **Preserve spaces** | `Sprint Planning` | Natural, like title | Problematic on some filesystems |
| **dot-separated** | `sprint.planning` | Dendron hierarchy | Confuses extensions |

### Date Format Options

| Format | Example | Use Case |
|--------|---------|----------|
| **ISO 8601** | `2026-03-02` | International standard, sortable |
| **Compact** | `20260302` | Shorter, still sortable |
| **Luxon/moment** | Custom via pattern | Flexible, localized |
| **RFC 3339** | `2026-03-02T09:30:00+10:30` | Full timestamp w/ timezone |
| **Unix timestamp** | `1740963000` | Unique, not human-readable |

### ID Generation Options

| Type | Example | Uniqueness | Readability | Length |
|------|---------|------------|-------------|--------|
| **UUID v4** | `550e8400-e29b-41d4-a716-446655440000` | Excellent | Poor | 36 |
| **nanoid** | `V1StGXR8_Z5jdHi6B` | Excellent | Fair | 21 (default) |
| **nanoid (8)** | `a8f3b2c1` | Good (for single notebook) | Good | 8 |
| **Timestamp** | `20260302093015` | Good (per-second) | Fair | 14 |
| **Short hash** | `c5d7e9b1` | Fair (collision risk) | Good | 8 |

---

## Findings: Schema Options for Filename Control

### Option A: Single `filename_format` Template (Recommended)

```json
{
  "name": "tasks",
  "type": "task",
  "globs": [".memory/task-*.md"],
  "filename_format": "task-{{ nanoid 8 }}-{{ slug .title }}.md"
}
```

**Pros**:
- Maximum flexibility with gomplate
- Single field, clear purpose
- Can express any pattern

**Cons**:
- Requires gomplate knowledge
- Complex patterns harder to validate

### Option B: Structured Fields

```json
{
  "name": "tasks",
  "type": "task",
  "filename": {
    "prefix": "task",
    "id_style": "nanoid8",
    "slug_style": "kebab",
    "date_format": null,
    "max_length": 50
  }
}
```

**Pros**:
- Explicit, discoverable options
- Easier validation
- IDE autocomplete friendly

**Cons**:
- More fields to maintain
- Less flexible than template
- Combinatorial complexity

### Option C: Hybrid (Preset + Override)

```json
{
  "name": "tasks",
  "type": "task",
  "filename_preset": "prefixed-id-slug",  // Predefined pattern
  "filename_format": null                  // Optional override
}
```

**Presets** could be:
- `slug-only`: `{{ slug .title }}.md`
- `date-slug`: `{{ date "2006-01-02" }}-{{ slug .title }}.md`
- `id-slug`: `{{ nanoid 8 }}-{{ slug .title }}.md`
- `prefixed-id-slug`: `{{ .type }}-{{ nanoid 8 }}-{{ slug .title }}.md`

**Pros**:
- Easy for common cases
- Full power when needed
- Progressive disclosure

**Cons**:
- Two ways to do things
- Preset names must be learned

### Recommendation: Option A with Smart Defaults

Use `filename_format` as the primary field, with sensible fallback behavior when absent.

---

## Findings: Fallback Behavior

### When Group Has No `filename_format`

**Proposed Resolution Chain**:

1. **Group-level**: `group.filename_format` (explicit)
2. **Notebook-level**: `notebook.default_filename_format` (NEW field in `.jot.json`)
3. **Global fallback**: Pure slugified title (`{{ slug .title }}.md`)

### Notebook-Level Default Example

```json
{
  "name": "my-notebook",
  "default_filename_format": "{{ date \"2006-01-02\" }}-{{ slug .title }}.md",
  "groups": [
    {
      "name": "tasks",
      "type": "task",
      "filename_format": "task-{{ nanoid 8 }}-{{ slug .title }}.md"
    },
    {
      "name": "notes",
      "type": "note"
      // No filename_format → uses notebook default
    }
  ]
}
```

### Global Fallback Behavior

When neither group nor notebook specifies format:

```go
// internal/services/filename.go (proposed)
const DefaultFilenameFormat = "{{ slug .title }}.md"

func ResolveFilename(group *NotebookGroup, notebook *NotebookConfig, title string) (string, error) {
    format := DefaultFilenameFormat
    
    if group != nil && group.FilenameFormat != "" {
        format = group.FilenameFormat
    } else if notebook.DefaultFilenameFormat != "" {
        format = notebook.DefaultFilenameFormat
    }
    
    return renderTemplate(format, map[string]any{"title": title, "type": group.Type})
}
```

---

## Findings: Edge Cases

### 1. Title Collision Handling

| Strategy | Behavior | Example |
|----------|----------|---------|
| **Error** (strict) | Fail if file exists | Error: "note already exists: sprint-planning.md" |
| **Append number** | Add suffix | `sprint-planning.md` → `sprint-planning-2.md` |
| **Append timestamp** | Add time suffix | `sprint-planning.md` → `sprint-planning-093015.md` |
| **Force unique ID** | Always include ID | `a8f3b2c1-sprint-planning.md` (no collision possible) |

**Current Jot behavior**: Error (strict) — `cmd/notes_add.go:91`

**Recommendation**: Keep strict behavior as default. If user wants auto-increment, they should use ID in filename_format.

### 2. Unicode/Emoji in Titles

**Current Jot slugify** (`internal/core/strings.go`):
```go
// Removes ALL non-ASCII characters
re := regexp.MustCompile(`[^a-z0-9\s-]`)
text = re.ReplaceAllString(text, "")
```

**Problem**:
- Title: "📝 Meeting Notes 会议"
- Current slug: "meeting-notes" (emoji and Chinese removed)
- If title is only emoji/unicode: **empty filename error**

**Comparison with gosimple/slug**:
```go
slug.Make("影師")        // → "ying-shi" (transliteration)
slug.Make("Hellö Wörld") // → "hello-world" (transliteration)
```

**Recommendation**: Replace custom slugify with `gosimple/slug` for better Unicode handling (transliteration to ASCII).

### 3. Very Long Titles

**Filesystem limits**:
- ext4: 255 bytes
- NTFS: 255 characters
- macOS HFS+: 255 UTF-16 code units

**Current behavior**: No truncation — can create invalid filenames.

**Recommendation**: Add max length to filename format:
```go
// Proposed: truncate slug to N characters
func truncateSlug(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    // Truncate at word boundary if possible
    ...
}
```

Template function: `{{ slug .title | trunc 50 }}`

### 4. Empty Title After Slugify

**Current behavior**: Error if title produces empty slug.

**Edge cases**:
- Title: "---" → slug: "" → Error
- Title: "???" → slug: "" → Error
- Title: "📝" → slug: "" → Error (with current slugify)

**Recommendation**: If slug is empty, fallback to timestamp-based filename.

```go
slug := Slugify(title)
if slug == "" {
    slug = time.Now().Format("2006-01-02-150405")
}
```

### 5. Special Characters Reserved by Filesystems

| Character | Windows | macOS | Linux |
|-----------|---------|-------|-------|
| `<` `>` `:` `"` `/` `\` `|` `?` `*` | ❌ Forbidden | ⚠️ Some issues | ✅ Mostly OK |
| Leading/trailing spaces | ❌ Stripped | ⚠️ Issues | ✅ OK |
| `.` as first char | ⚠️ Hidden | ⚠️ Hidden | ⚠️ Hidden |

**Recommendation**: The `gosimple/slug` library handles all of these by design (outputs only `a-z`, `0-9`, `-`, `_`).

---

## Proposed Schema Changes

### NotebookGroup (Updated)

```go
type NotebookGroup struct {
    Name           string   `json:"name"`
    Type           string   `json:"type,omitempty"`            // Canonical type name for --type flag
    Aliases        []string `json:"aliases,omitempty"`         // Alternative names for fuzzy matching
    Globs          []string `json:"globs"`                     // For post-hoc matching
    FilenameFormat string   `json:"filename_format,omitempty"` // Gomplate template for filename
    Template       string   `json:"template,omitempty"`        // Content template reference
    WorkflowID     string   `json:"workflow_id,omitempty"`
    // REMOVED: Metadata — use template frontmatter instead
}
```

### NotebookConfig (New Field)

```go
type NotebookConfig struct {
    // ... existing fields ...
    DefaultFilenameFormat string `json:"default_filename_format,omitempty"` // Fallback format
}
```

### Gomplate Functions Available

| Function | Example | Result |
|----------|---------|--------|
| `slug` | `{{ slug "Hello World!" }}` | `hello-world` |
| `nanoid N` | `{{ nanoid 8 }}` | `a8f3b2c1` |
| `date FMT` | `{{ date "2006-01-02" }}` | `2026-03-02` |
| `trunc N` | `{{ "long-title" \| trunc 5 }}` | `long-` |
| `now` | `{{ now.Format "20060102" }}` | `20260302` |

---

## Implementation Recommendations

### Phase 1: Minimal Viable Change
1. Add `type` field to `NotebookGroup`
2. Add `filename_format` field to `NotebookGroup`
3. Add `default_filename_format` to `NotebookConfig`
4. Replace custom `Slugify()` with `gosimple/slug`
5. Implement `--type` flag on `notes add`

### Phase 2: Enhanced UX
6. Add `aliases` field for fuzzy matching
7. Add interactive group selection when no path/type given
8. Add filename preview in dry-run mode

### Phase 3: Robustness
9. Add filename length validation/truncation
10. Add collision detection with suggested alternatives
11. Add Unicode transliteration via slug library

---

## References

- [gosimple/slug](https://pkg.go.dev/github.com/gosimple/slug) — Go slug library with Unicode transliteration
- [iancoleman/strcase](https://pkg.go.dev/github.com/iancoleman/strcase) — Case conversion (kebab, snake, camel)
- [Dendron Special Notes Config](https://wiki.dendron.so/notes/j1120e2f53301nhdn1wkp05) — dateFormat, addBehavior patterns
- [Zettelkasten Introduction](https://zettelkasten.de/introduction/) — ID-based note naming philosophy
- [Foam Note Properties](https://foambubble.github.io/foam/user/features/note-properties) — Frontmatter-driven metadata
- [gomplate documentation](https://docs.gomplate.ca/) — Template rendering with 200+ functions
- Jot codebase: `internal/core/strings.go` — Current slugify implementation
- Jot codebase: `internal/services/note.go` — ResolvePath function
