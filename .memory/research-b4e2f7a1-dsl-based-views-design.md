---
id: b4e2f7a1
title: "Research: DSL-based views design for opennotes"
type: "research"
created_at: 2026-02-17T18:47:00+10:30
updated_at: 2026-02-18T20:58:00+10:30
status: completed
epic_id: f661c068
phase_id: null
related_task_id: null
assigned_to: research-dsl-views-session
---

# Research: DSL-based views design for opennotes

## Research Questions

### 1. How can builtin views be expressed as DSL queries?
The search DSL (`internal/search/parser/`) supports Gmail-style queries with fields: `tag`, `title`, `path`, `created`, `modified`, `body`, `status`. Can each builtin view be represented as a single DSL query string?

- **today**: `modified:>=2026-02-17` — straightforward date filter. Does the DSL support `>=` on date fields correctly? Does it resolve relative dates like `{{today}}`?
- **recent**: Needs all notes sorted by modified date DESC, limit 20. The DSL has no `sort` or `limit` concepts — how to handle this?
- **kanban**: Needs notes grouped by `status` field. The DSL has no `group by` concept — is this even expressible, or does it need post-query Go logic?
- **untagged**: Needs notes where tags are empty/nil. The DSL supports `tag:value` but not "tag is absent" (`-tag:*` or similar). What negation patterns exist?

### 2. What gaps exist in the current DSL grammar?
Current DSL supports: field matching, comparison operators (`>`, `>=`, `<`, `<=`), negation (`-term`, `-field:value`), quoted strings, date literals, free text.

Missing for views:
- **Sorting**: No `sort:field` or `order:desc` syntax
- **Limits**: No `limit:N` syntax
- **Grouping**: No `group:field` syntax (needed for kanban)
- **Wildcard negation**: Can you express "field has no value" (e.g., `tag:*` then negate it)?
- **Relative dates**: Template variables like `{{today}}` — should these be in the DSL or resolved before parsing?
- **Existence checks**: "has:tag" or "missing:tag" patterns

Should these be added to the DSL grammar, or handled as view-level metadata outside the query string?

### 3. How should ViewDefinition/ViewQuery types be redesigned?
Current `core.ViewQuery` has SQL-oriented fields: `SelectColumns`, `GroupBy`, `Having`, `AggregateColumns`, `Distinct`. These are dead code.

Proposed direction: A view could be a DSL query string + optional metadata (sort, limit, group). Questions:
- Should `ViewDefinition.Query` just be a `string` (the DSL query)?
- Or a struct with `Query string`, `Sort string`, `Limit int`, `GroupBy string`?
- How do parameters (`{{today}}`, `{{param_name}}`) get resolved — before or after DSL parsing?
- Should `ViewCondition` be removed entirely in favor of raw DSL strings?

### 4. How should user-created custom views (saved queries) work?
Users should be able to save DSL queries as named views. Questions:
- Where are custom views stored? In notebook `.opennotes.json`? In global config? Both (with precedence)?
- What's the format? `{ "name": "my-view", "query": "tag:work status:todo", "sort": "modified:desc", "limit": 50 }`?
- Can users override builtin views?
- How does `opennotes notes view --list` discover and display custom vs builtin views?
- Should there be a `opennotes notes view --save <name> <query>` command?

### 5. What about graph-based views (orphans, broken-links)?
These views require traversing note links — they can't be expressed as search queries. Questions:
- Keep them as special cases in `view_special.go`? (Current approach, already working)
- Should they be flagged differently in `ViewDefinition` (e.g., `type: "special"` vs `type: "query"`)?
- Could they eventually be index-backed (index link relationships in Bleve)?

### 6. How to clean up dead SQL code?
`internal/services/view.go` has ~1000+ lines including `GenerateSQL()`, SQL validation, aggregate function handling. Questions:
- Remove all SQL code in one pass or incrementally?
- What code is still useful? (Template variable resolution `{{today}}`, parameter handling, view discovery/loading)
- What tests in `view_test.go` test SQL behavior vs. general view logic?

## Execution Plan

### Phase 1: Map the existing DSL pipeline (codemapper)

Use the `codemapper` skill to build a complete picture of how queries flow through the system.

**Actions:**
- `cm trace "cmd" "search"` — trace how CLI commands reach the search service
- `cm query "Parse" --format ai` — find all Parse-related symbols in the parser package
- `cm query "BuildQuery" --format ai` — understand how QueryConditions become search.Query AST
- `cm callers "SearchService.BuildQuery"` — who calls BuildQuery today?
- `cm query "ViewDefinition" --format ai` — map all view type usage
- `cm stats internal/services/view.go` — quantify dead SQL code vs. reusable code
- `cm trace "internal/search" "internal/services"` — map dependencies between search and services

**Deliverable:** A clear map of: DSL string → parser → search.Query AST → Bleve query → results. Identify exactly which connection points the view system needs to plug into.

### Phase 2: Explore design options (brainstorming)

Use the `brainstorming` skill to explore the design space before committing to an approach.

**Key design tensions to explore:**
- **DSL purity vs. view metadata**: Should sort/limit/group be DSL syntax extensions, or separate view-level config alongside a query string?
- **Thin views vs. rich views**: Is a view just a saved query string, or does it need its own execution model?
- **Grammar extension cost**: What's the impact of adding `sort:`, `limit:`, `group:` to the Participle grammar vs. keeping them out?
- **Template resolution timing**: Resolve `{{today}}` before DSL parsing (simple string substitution) or make the parser date-aware?
- **User experience**: `notes view today` vs. `notes search "modified:>=today"` — what's the value-add of named views?

**Deliverable:** 2-3 concrete design options with tradeoffs articulated.

### Phase 3: Design the CLI surface (creating-cli-tools)

Use the `creating-cli-tools` skill to design the user-facing commands.

**Questions to resolve:**
- How does `notes view <name>` execute a saved query?
- How does `notes view --save <name> "<query>"` work?
- What flags does `notes view` need? (`--format`, `--sort`, `--limit`, `--group`?)
- How do users list, edit, delete saved views?
- How do parameters work? (`notes view my-view --param sprint=Q1`)
- Output format: reuse `notes search` output rendering or separate?

**Deliverable:** CLI specification for the view command family.

### Phase 4: Validate the design (architect-reviewer)

Use the `architect-reviewer` skill to validate the chosen approach.

**Review criteria:**
- Does the design compose well with existing search infrastructure?
- Is it backwards-compatible with the working `orphans`/`broken-links` views?
- Does it scale to complex user queries without becoming another SQL?
- Is the ViewDefinition type clean and minimal?
- Are there edge cases that break the model?

**Deliverable:** Go/no-go recommendation with any design adjustments.

### Phase 5: Plan the cleanup (refactoring-specialist)

Use the `refactoring-specialist` skill to plan safe removal of dead SQL code.

**Actions:**
- Identify which code in `view.go` is reusable (template resolution, view discovery, parameter handling)
- Identify which code is dead SQL (GenerateSQL, SQL validation, aggregate functions, escapeSQL, etc.)
- Plan incremental removal strategy (don't break tests that test non-SQL behavior)
- Identify which tests in `view_test.go` are SQL-specific vs. general

**Deliverable:** A file-by-file cleanup plan with safe removal order.

### Phase 6: Write implementation plan (writing-plans)

Use the `writing-plans` skill to produce the final implementation task(s).

**Deliverable:** Detailed implementation plan with exact file paths, code examples, and verification steps — ready for an engineer with zero context to execute.

## Summary

**All phases complete.** The DSL-based views design research is finished with a comprehensive implementation plan ready for execution.

- **Phase 1-2**: Mapped DSL pipeline, identified 3 query paths (text search, boolean query, Participle DSL parser with zero production callers)
- **Phase 3**: Designed CLI surface with pipe syntax (`filter | directives`)
- **Phase 4**: Architecture validation — GO ✅
- **Phase 5**: SQL cleanup plan — ~500 lines removable, ~550 lines reusable
- **Phase 6**: Implementation plan created — 10 TDD tasks ready for execution

**Selected Design**: Option C (Hybrid — Pipe Syntax). Views use DSL query strings with format `"filter DSL | directives"`. Example: `"modified:>=today | sort:modified:desc"`.

**Next Step**: Execute [plan-b4e2f7a1-dsl-views-implementation.md](plan-b4e2f7a1-dsl-views-implementation.md) using `superpowers:executing-plans` skill.

## Findings

### Phase 1: DSL Pipeline Map

#### Three Query Pipelines

**Pipeline 1: Text/Fuzzy Search** (used by `notes search`)
```
CLI args → cmd/notes_search.go
         → NoteService.SearchNotes(query, fuzzy)
         → NoteService.getAllNotes() → Index.Find(empty FindOpts)
         → SearchService.FuzzySearch() or .TextSearch() → in-memory filtering
         → []Note results
```

**Pipeline 2: Boolean Query** (used by `notes search query`)
```
CLI flags (--and, --or, --not) → cmd/notes_search_query.go
         → SearchService.ParseConditions(andFlags, orFlags, notFlags)
         → []QueryCondition → NoteService.SearchWithConditions(conditions)
         → SearchService.BuildQuery(ctx, conditions)
         → *search.Query AST → Index.Find(FindOpts{Query: query})
         → bleve.TranslateQuery(query) → Bleve execution → results
```

**Pipeline 3: DSL Parser** (exists but UNUSED in production)
```
DSL string → parser.New().Parse(input)
           → Participle lexer+parser → queryAST
           → convert() → *search.Query AST
           → Index.FindByQueryString — ZERO production callers
```

#### Key Connection Points for Views

1. **Input**: Convert view definition into `search.FindOpts` (contains `*search.Query` + sort + limit + offset)
2. **Execution**: Call `Index.FindByQueryString(ctx, queryString, opts)` — already exists, ideal entry point

#### Dead Code Inventory in `view.go` (1354 lines)

| Code Block | Lines | Status |
|-----------|-------|--------|
| `initializeBuiltinViews()` | 48-157 | **Dead SQL** — all definitions use SQL syntax |
| `GetView()` + view loading | 159-257 | **Reusable** — view discovery hierarchy works |
| `ResolveTemplateVariables()` | 258-390 | **Reusable** — backend-agnostic string templating |
| `ValidateViewDefinition()` | 447-676 | **Dead SQL** — validates SQL fields/operators |
| `ParseViewParameters()` | 804-834 | **Reusable** — parses `key=value` strings |
| `GenerateSQL()` | 867-1032 | **Dead SQL** — full SQL generator |
| `GroupResults()` | 1072-1196 | **Partly reusable** — app-level grouping by field |
| `ListAllViews()` | 1198-1354 | **Reusable** — view listing and discovery |

### Phase 2: Design Options

#### Option A: Thin Views — DSL String + Metadata Sidecar ⭐ RECOMMENDED

```go
type ViewDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  []ViewParameter `json:"parameters,omitempty"`
    Query       string          `json:"query"`              // DSL string
    Sort        string          `json:"sort,omitempty"`      // "modified:desc"
    Limit       int             `json:"limit,omitempty"`     // 20
    GroupBy     string          `json:"group_by,omitempty"`  // "status"
    Type        string          `json:"type,omitempty"`      // "query" or "special"
}
```

Builtin examples:
- `today`: `{ "query": "modified:>=today", "sort": "modified:desc" }`
- `recent`: `{ "query": "", "sort": "modified:desc", "limit": 20 }`
- `kanban`: `{ "query": "status:*", "group_by": "status" }`
- `untagged`: `{ "query": "missing:tag", "sort": "created:desc" }`
- `orphans`: `{ "type": "special" }`

**Pros**: Zero grammar changes, `FindOpts` already has sort/limit/offset, correct separation of concerns (filtering vs presentation), 90% of wiring exists.
**Cons**: Sort/limit aren't in the query string, two "languages".

#### Option B: Rich DSL — Sort/Limit as Grammar Extensions

Extend Participle grammar so sort/limit/group are first-class DSL tokens:

```go
type ViewDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  []ViewParameter `json:"parameters,omitempty"`
    Query       string          `json:"query"`              // Full DSL string including directives
    Type        string          `json:"type,omitempty"`      // "query" or "special"
}
```

Builtin examples:
- `today`: `{ "query": "modified:>=today sort:modified:desc" }`
- `recent`: `{ "query": "sort:modified:desc limit:20" }`
- `kanban`: `{ "query": "has:status group:status sort:path:asc" }`
- `untagged`: `{ "query": "missing:tag sort:created:desc" }`
- `orphans`: `{ "type": "special" }`

User-created examples:
- `{ "name": "work-inbox", "query": "tag:work status:todo sort:created:desc limit:50" }`
- `{ "name": "stale-drafts", "query": "status:draft modified:<this-month sort:modified:asc" }`

CLI usage:
```bash
opennotes notes view today                    # executes builtin
opennotes notes search "tag:work sort:modified:desc limit:10"  # ad-hoc with directives
opennotes notes view --save my-view "tag:work status:todo sort:created:desc"
```

**Pros**: Single self-contained query string. `ViewDefinition` is minimal — just name + query string. Users can copy/paste/share entire queries. Saved views and ad-hoc search use identical syntax.
**Cons**: Grammar changes risk regressions. `sort`/`limit`/`group` collide with lexer Field token (needs new `Directive` token type). Conflates filtering with presentation in the AST. `group:status` is post-query Go logic pretending to be a query directive.

#### Option C: Hybrid — Pipe Syntax ⭐ SELECTED

The query string has two halves separated by `|`: the left side is the filter DSL (parsed by Participle), the right side is presentation directives (parsed by simple key:value splitter). The pipe is processed *before* DSL parsing.

```go
type ViewDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  []ViewParameter `json:"parameters,omitempty"`
    Query       string          `json:"query"`              // "filter DSL | directives"
    Type        string          `json:"type,omitempty"`      // "query" or "special"
}
```

Builtin examples:
- `today`: `{ "query": "modified:>=today | sort:modified:desc" }`
- `recent`: `{ "query": "| sort:modified:desc limit:20" }`
- `kanban`: `{ "query": "has:status | group:status sort:path:asc" }`
- `untagged`: `{ "query": "missing:tag | sort:created:desc" }`
- `orphans`: `{ "type": "special" }`

User-created examples:
- `{ "name": "work-inbox", "query": "tag:work status:todo | sort:created:desc limit:50" }`
- `{ "name": "stale-drafts", "query": "status:draft modified:<this-month | sort:modified:asc" }`
- `{ "name": "all-by-title", "query": "| sort:title:asc" }` (no filter, just sorted listing)
- `{ "name": "weekly-review", "query": "modified:>=this-week | sort:modified:desc group:status" }`

CLI usage:
```bash
opennotes notes view today                                      # executes builtin
opennotes notes search "tag:work | sort:modified:desc limit:10" # ad-hoc with pipe
opennotes notes view --save my-view "tag:work status:todo | sort:created:desc"
```

Execution flow:
```
ViewDefinition.Query
  → ResolveTemplateVariables(query) → "modified:>=2026-02-17 | sort:modified:desc"
  → SplitPipe(resolved) → filterPart="modified:>=2026-02-17", directivesPart="sort:modified:desc"
  → parser.Parse(filterPart) → *search.Query
  → parseDirectives(directivesPart) → sort, limit, groupBy
  → FindOpts{Query: q, Sort: sort, Limit: limit}
  → Index.Find(ctx, opts)
  → if groupBy: post-process results in Go
```

Directive grammar (simple, not Participle):
```
directives  = directive (" " directive)*
directive   = "sort:" field (":" ("asc"|"desc"))?
            | "limit:" number
            | "group:" field
            | "offset:" number
```

**Pros**: Single self-contained query string — users copy/paste/share complete views. No Participle grammar changes — the DSL stays pure for filtering. Pipe is a clear visual separator between "what to find" and "how to present". Familiar shell metaphor (`grep pattern | sort`). `ViewDefinition` is minimal — just name + query string. Directives parser is trivial (~30 lines of Go, no Participle needed). The filter part can be empty (just `| sort:modified:desc`) for "all notes, sorted".
**Cons**: `|` character needs to be documented. Two parsers (but the directives parser is trivial). Not standard in Gmail-style search (but this isn't Gmail — it's a CLI tool where pipes are natural).

### Recommendation

**Option C (Hybrid — Pipe Syntax)** selected by project owner. Rationale:
1. Best UX — single self-contained query string that users can share, save, and compose
2. Clean visual separation between filtering and presentation (`filter | directives`)
3. No Participle grammar changes — DSL stays pure, directives parser is trivial
4. Natural fit for a CLI tool where pipe metaphor is well-understood
5. `ViewDefinition` type is minimal — just name + query string (no sidecar fields)
6. Identical syntax works in both `notes view` (saved) and `notes search` (ad-hoc)

### DSL Grammar Gaps

| Gap | Severity | Fix |
|-----|----------|-----|
| Existence checks (`has:tag`, `missing:tag`) | **Critical** | Add new expression types to grammar |
| Wildcard field values (`status:*`) | **Critical** | Add Wildcard token or `has:`/`missing:` keywords |
| OR syntax (`tag:work OR tag:personal`) | **Medium** | `OrExpr` exists in AST but parser can't produce it |
| Relative dates in DSL | **Already supported** | Bleve's `parseDate()` handles `today`, `yesterday`, etc. |

### Phase 3: CLI Surface Design

#### Design Principles Applied

Following clig.dev guidelines and existing opennotes CLI patterns:
- **Composability**: Views reuse the same output rendering as `notes search` and `notes list`
- **Discoverability**: `notes view` with no args lists all views; `--help` shows full syntax
- **Consistency**: Same `--format`, `--notebook` flags as other `notes` subcommands
- **Script-friendly**: `--json` output for piping; exit codes for scripting
- **Human-first**: Default output uses glamour markdown rendering via `TuiRender`

#### CLI Specification: `notes view` Command Family

##### USAGE

```
opennotes notes view [name] [flags]
opennotes notes view --save <name> "<query>"
opennotes notes view --delete <name>
opennotes notes view --list [--format list|json]
```

##### Subcommand Semantics

| Invocation | Behavior | Idempotent | State Change |
|-----------|----------|------------|--------------|
| `notes view` | List all available views | Yes | None |
| `notes view <name>` | Execute named view, display results | Yes | None |
| `notes view --save <name> "<query>"` | Save a custom view to notebook config | No | Writes `.opennotes.json` |
| `notes view --delete <name>` | Delete a custom view from notebook config | No | Writes `.opennotes.json` |
| `notes view --list` | List all available views (explicit) | Yes | None |

##### Args & Flags Table

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `[name]` | — | positional | — | No | View name to execute |
| `--format` | `-f` | string | `list` | No | Output format: `list`, `table`, `json` |
| `--param` | `-p` | string | — | No | View parameters: `key=value,key2=value2` |
| `--list` | `-l` | bool | `false` | No | List all available views |
| `--save` | `-s` | string | — | No | Save query as named view |
| `--delete` | — | string | — | No | Delete a named custom view |
| `--sort` | — | string | — | No | Override view sort: `field:asc\|desc` |
| `--limit` | — | int | `0` | No | Override view result limit |
| `--group` | — | string | — | No | Override view group-by field |
| `--notebook` | `-n` | string | — | No | Notebook path (inherited from parent) |

##### Query String Syntax (Pipe Convention from Phase 2 — Option C)

```
<filter DSL> | <directives>
```

- **Left of pipe**: DSL filter query (parsed by Participle parser)
- **Right of pipe**: Presentation directives (simple `key:value` parser)
- **Pipe is optional**: `tag:work` is valid (no directives)
- **Filter is optional**: `| sort:modified:desc limit:20` is valid (all notes, sorted)

Directives:
- `sort:<field>:<asc|desc>` — Sort results (default: `asc`)
- `limit:<n>` — Limit result count
- `group:<field>` — Group results by field value
- `offset:<n>` — Skip first N results (for pagination)

##### View Resolution Order (Precedence)

1. **Notebook views** (`.opennotes.json` → `views` section) — highest priority
2. **Global views** (`~/.config/opennotes/config.json` → `views` section)
3. **Built-in views** (hardcoded in Go) — lowest priority

Users can override built-in views by defining a view with the same name in notebook or global config.

##### Built-in Views (New DSL Definitions)

```json
{
  "today": {
    "name": "today",
    "description": "Notes created or updated today",
    "query": "modified:>=today | sort:modified:desc"
  },
  "recent": {
    "name": "recent",
    "description": "Recently modified notes (last 20)",
    "query": "| sort:modified:desc limit:20"
  },
  "kanban": {
    "name": "kanban",
    "description": "Notes grouped by status",
    "query": "has:status | group:status sort:title:asc"
  },
  "untagged": {
    "name": "untagged",
    "description": "Notes without any tags",
    "query": "missing:tag | sort:created:desc"
  },
  "orphans": {
    "name": "orphans",
    "description": "Notes with no incoming links",
    "type": "special"
  },
  "broken-links": {
    "name": "broken-links",
    "description": "Notes with broken references",
    "type": "special"
  }
}
```

##### `--save` Behavior

```bash
opennotes notes view --save work-inbox "tag:work status:todo | sort:created:desc limit:50"
```

1. Validates the query string (parses filter DSL + directives)
2. Rejects if name conflicts with a built-in view (unless `--force`)
3. Writes to notebook `.opennotes.json` under `views.<name>`:
   ```json
   {
     "views": {
       "work-inbox": {
         "name": "work-inbox",
         "description": "Work inbox — todo items sorted by creation",
         "query": "tag:work status:todo | sort:created:desc limit:50"
       }
     }
   }
   ```
4. Confirms: `View "work-inbox" saved to notebook config`

##### `--delete` Behavior

```bash
opennotes notes view --delete work-inbox
```

1. Rejects if name is a built-in view (cannot delete built-ins)
2. Removes from notebook `.opennotes.json` → `views` section
3. Confirms: `View "work-inbox" deleted from notebook config`

##### Output Rules

| Stream | Content |
|--------|---------|
| stdout | View results (notes list), view listing, JSON output |
| stderr | Errors, warnings, diagnostics |

| Format | Behavior |
|--------|----------|
| `list` (default) | Uses `TuiRender("note-list", ...)` — same as `notes list` / `notes search` |
| `table` | ASCII table with columns: path, title, modified, tags |
| `json` | JSON array of note objects (same schema as `notes list --format json`) |

For `--list`:
| Format | Behavior |
|--------|----------|
| `list` (default) | Grouped by origin (built-in, global, notebook) with descriptions |
| `json` | `{"views": [...]}` JSON array |

For grouped views (kanban):
| Format | Behavior |
|--------|----------|
| `list` | Notes grouped under status headers |
| `json` | `{"groups": {"todo": [...], "done": [...]}}` |

##### Error & Exit Code Map

| Exit Code | Condition |
|-----------|-----------|
| `0` | Success |
| `1` | View not found, query parse error, notebook not found |
| `2` | Invalid usage (bad flag combination, missing required args) |

| Error | Message |
|-------|---------|
| View not found | `Error: view "xyz" not found. Run 'opennotes notes view --list' to see available views.` |
| Invalid query | `Error: failed to parse query: <parser error detail>` |
| Invalid directive | `Error: unknown directive "foo" in "| foo:bar". Valid: sort, limit, group, offset` |
| Delete built-in | `Error: cannot delete built-in view "today". Built-in views can only be overridden.` |
| Save conflict | `Error: "today" is a built-in view. Use --force to override in notebook config.` |

##### Integration with `notes search`

The pipe syntax also works in `notes search` for ad-hoc queries with directives:

```bash
# Current (no change needed):
opennotes notes search "meeting"
opennotes notes search --fuzzy "mtng"

# New (pipe syntax in search):
opennotes notes search "tag:work | sort:modified:desc limit:10"
```

This means `notes search` and `notes view` share the same query execution pipeline. The only difference is that `notes view` resolves a named view definition first, then executes it identically to `notes search`.

##### Example Invocations

```bash
# List all views
opennotes notes view
opennotes notes view --list --format json

# Execute built-in views
opennotes notes view today
opennotes notes view recent --format table
opennotes notes view kanban
opennotes notes view untagged --format json

# Execute with runtime overrides
opennotes notes view today --limit 5
opennotes notes view recent --sort title:asc
opennotes notes view kanban --group priority

# Execute with parameters
opennotes notes view my-sprint --param sprint=Q1-S3

# Save custom views
opennotes notes view --save work-inbox "tag:work status:todo | sort:created:desc limit:50"
opennotes notes view --save stale-drafts "status:draft modified:<this-month | sort:modified:asc"
opennotes notes view --save all-sorted "| sort:title:asc"

# Delete custom view
opennotes notes view --delete work-inbox

# Ad-hoc search with pipe syntax
opennotes notes search "tag:work | sort:modified:desc limit:10"

# Pipe to external tools
opennotes notes view today --format json | jq '.[].title'
```

##### New `ViewDefinition` Type (Proposed)

```go
type ViewDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  []ViewParameter `json:"parameters,omitempty"`
    Query       string          `json:"query"`           // "filter DSL | directives"
    Type        string          `json:"type,omitempty"`  // "query" (default) or "special"
}
```

Key change: `Query` becomes a plain `string` (was `ViewQuery` struct with SQL fields). The pipe-separated query string encodes everything. `Type` distinguishes DSL-queryable views from special graph-traversal views (orphans, broken-links).

##### CLI Flag Overrides vs. Query Directives

When a user provides both a view's query directives AND CLI flags, CLI flags win:

```bash
# View definition: "| sort:modified:desc limit:20"
# CLI override: --limit 5 --sort title:asc
# Effective: sort=title:asc, limit=5 (CLI flags override)
```

Precedence: CLI flags > query directives > defaults

### Phase 4: Architecture Validation

**Skill Applied**: `architect-reviewer`

#### Executive Summary

**GO** ✅ — The pipe syntax design (Option C) is architecturally sound and ready for implementation.

The design composes cleanly with existing infrastructure, maintains backwards compatibility, and avoids the complexity trap of becoming "SQL 2.0". Three minor adjustments are recommended before implementation.

#### Review Criteria Assessment

##### 1. Composition with Existing Search Infrastructure ✅ PASS

**Finding**: The design maps directly to existing `FindOpts` struct with no modifications needed.

| Pipe Directive | Maps To | Status |
|---------------|---------|--------|
| Filter DSL (left of `\|`) | `FindOpts.Query *search.Query` | ✅ Exists |
| `sort:field:dir` | `FindOpts.Sort SortSpec` | ✅ Exists |
| `limit:n` | `FindOpts.Limit int` | ✅ Exists |
| `offset:n` | `FindOpts.Offset int` | ✅ Exists |
| `group:field` | Post-query Go logic | ✅ Pattern exists in `SpecialViewExecutor` |

**Evidence**: `internal/search/options.go` already defines `SortSpec` with `SortByCreated`, `SortByModified`, `SortByTitle`, `SortByRelevance`. The execution path is:

```
Query string → SplitPipe() → filter | directives
                              ↓          ↓
                        Parser.Parse()  parseDirectives()
                              ↓          ↓
                        *search.Query   sort, limit, offset, groupBy
                              ↓          ↓
                        FindOpts{Query: q, Sort: s, Limit: l, Offset: o}
                              ↓
                        Index.Find(ctx, opts)
                              ↓
                        if groupBy: groupResults(results, groupBy)
```

**Verdict**: Zero changes to `FindOpts` or `Index` interface required. The design plugs into existing infrastructure.

##### 2. Backwards Compatibility with Special Views ✅ PASS

**Finding**: The `Type: "special"` field cleanly separates DSL-queryable views from graph-traversal views.

| View | Type | Execution Path |
|------|------|----------------|
| `today`, `recent`, `kanban`, `untagged` | `"query"` (default) | DSL parser → `Index.Find()` |
| `orphans`, `broken-links` | `"special"` | `SpecialViewExecutor.Execute*View()` |

**Evidence**: `internal/services/view_special.go` already implements `ExecuteBrokenLinksView()` and `ExecuteOrphansView()` with Go logic that cannot be expressed as search queries. These views traverse note links to find broken references.

**Verdict**: No changes to special view handling. The `Type` field provides clean dispatch routing.

##### 3. Complexity Scaling (Not Becoming SQL 2.0) ✅ PASS

**Finding**: The design has deliberate constraints that prevent feature creep.

| Feature | In Scope | Out of Scope (Intentional) |
|---------|----------|---------------------------|
| Filtering | ✅ Field matching, comparisons, negation, free text | ❌ JOINs, subqueries |
| Sorting | ✅ Single field, asc/desc | ❌ Multi-field sort, expressions |
| Pagination | ✅ limit + offset | ❌ Cursors, keyset pagination |
| Grouping | ✅ Single field post-query grouping | ❌ Aggregations, HAVING, nested groups |
| Presentation | ✅ Directives separated by pipe | ❌ Projection (select columns) |

**Key constraint**: The pipe syntax explicitly separates "what to find" (DSL) from "how to present" (directives). This prevents the DSL from evolving toward SQL expressiveness.

**Verdict**: The design has clear boundaries. Complexity is bounded by the directive grammar, not the DSL grammar.

##### 4. ViewDefinition Type Cleanliness ✅ PASS

**Proposed type** (from Phase 3):
```go
type ViewDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  []ViewParameter `json:"parameters,omitempty"`
    Query       string          `json:"query"`           // "filter DSL | directives"
    Type        string          `json:"type,omitempty"`  // "query" (default) or "special"
}
```

**Comparison to current type**:
```go
// CURRENT (SQL-oriented, dead code)
type ViewDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  []ViewParameter `json:"parameters,omitempty"`
    Query       ViewQuery       `json:"query"`  // ← Complex struct with 10 SQL fields
}
```

**Improvement metrics**:
- Fields removed: 10 (SelectColumns, GroupBy, Having, AggregateColumns, Distinct, Conditions, OrderBy, etc.)
- New fields: 1 (`Type`)
- Net complexity reduction: 90%

**Verdict**: The new type is minimal and self-documenting. A view is just a name + query string.

##### 5. Edge Cases Analysis ⚠️ MINOR ISSUES IDENTIFIED

| Edge Case | Handling | Status |
|-----------|----------|--------|
| Empty filter (`\| sort:modified:desc`) | Valid — means "all notes, sorted" | ✅ Designed |
| No pipe (`tag:work`) | Valid — filter only, default presentation | ✅ Designed |
| Multiple pipes (`a \| b \| c`) | First pipe splits; remainder is directives | ⚠️ Needs explicit rule |
| Pipe in quoted string (`title:"a \| b"`) | Should NOT split on this pipe | ⚠️ Needs careful implementation |
| Unknown directive (`foo:bar`) | Error: "unknown directive 'foo'" | ✅ Designed |
| Conflicting directives (`limit:10 limit:20`) | Last wins, or error? | ⚠️ Needs decision |
| Case sensitivity (`Sort:modified` vs `sort:modified`) | Directives should be case-insensitive | ⚠️ Needs decision |

#### Recommended Adjustments

##### Adjustment 1: Pipe Splitting Rule (REQUIRED)

**Issue**: Pipe character inside quoted strings should not trigger split.

**Recommendation**: Split on first unquoted `|` only. Implementation:
```go
func SplitQuery(query string) (filter, directives string) {
    // Use a simple state machine to track quote state
    // Only split on '|' when not inside quotes
    inQuote := false
    for i, ch := range query {
        if ch == '"' { inQuote = !inQuote }
        if ch == '|' && !inQuote {
            return strings.TrimSpace(query[:i]), strings.TrimSpace(query[i+1:])
        }
    }
    return query, "" // No pipe found
}
```

**Examples: where quote-aware splitting applies (split should happen):**
- `tag:work | sort:modified:desc`
  - filter: `tag:work`
  - directives: `sort:modified:desc`
- `status:todo modified:>=today | limit:20`
  - filter: `status:todo modified:>=today`
  - directives: `limit:20`
- `| sort:title:asc limit:50`
  - filter: *(empty = all notes)*
  - directives: `sort:title:asc limit:50`

**Examples: where quote-aware splitting does NOT apply at that pipe (no split at quoted `|`):**
- `title:"A | B" tag:work`
  - no split; entire string is filter
- `body:"foo | bar baz"`
  - no split; entire string is filter
- `title:"roadmap | q1" | sort:modified:desc`
  - first `|` inside quotes is ignored
  - split on second (unquoted) `|`

**Ambiguous/edge cases to define explicitly:**
- `title:"unterminated | quote`
  - treat as parse error (unclosed quote)
- `tag:work | sort:title:asc | limit:10`
  - split on first unquoted `|`; reject extra `|` in directives
- `title:'A | B' | sort:...`
  - if single quotes are unsupported by DSL, this should not protect `|`; document behavior

##### Adjustment 2: Directive Conflict Rule (REQUIRED)

**Issue**: What happens with `limit:10 limit:20`?

**Recommendation**: Last directive wins (consistent with CLI flag behavior where later flags override earlier ones). Document this explicitly.

##### Adjustment 3: Case Insensitivity (RECOMMENDED)

**Issue**: Should `Sort:modified` work?

**Recommendation**: Directives are case-insensitive. Normalize to lowercase during parsing. This matches user expectations from CLI tools.

#### Architecture Fitness Functions

To maintain architectural integrity during implementation, add these checks:

1. **No SQL imports in view.go**: `grep -c "database/sql" internal/services/view.go` should return 0
2. **ViewDefinition.Query is string**: Type assertion test in `view_test.go`
3. **Special views use Type field**: All special views have `Type: "special"` in definition
4. **Directive grammar is closed**: New directives require explicit approval (prevent feature creep)

#### Technical Debt Assessment

| Item | Current State | After Implementation |
|------|---------------|---------------------|
| Dead SQL code in `view.go` | ~600 lines | 0 lines (Phase 5 cleanup) |
| ViewQuery struct | 10 SQL-oriented fields | Removed entirely |
| ViewCondition struct | SQL-oriented | Removed entirely |
| Test coverage for SQL paths | ~40% | N/A (code removed) |

#### Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Grammar gaps block built-in views | Medium | High | Phase 5 adds `has:`/`missing:` before view implementation |
| Pipe syntax confuses users | Low | Low | Document clearly; familiar to CLI users |
| Directive parser bugs | Low | Medium | Comprehensive test coverage; simple grammar |
| Migration breaks saved views | N/A | N/A | No migration needed — views are new |

#### Final Recommendation

**GO** ✅ — Proceed to Phase 5 (SQL cleanup planning) and Phase 6 (implementation planning).

**Pre-requisites for implementation**:
1. Add `has:` / `missing:` keywords to DSL grammar (Critical gap from Phase 2)
2. Implement `SplitQuery()` with quote-aware pipe splitting
3. Remove dead SQL code (Phase 5 cleanup)

**Post-implementation validation**:
1. All built-in views execute correctly
2. Custom view save/delete works
3. `notes search` supports pipe syntax
4. CLI flag overrides work correctly
5. Special views (`orphans`, `broken-links`) unchanged

### Phase 5: SQL Cleanup Plan

**Skill Applied**: `refactoring-specialist`

#### Executive Summary

The `view.go` file (1355 lines) contains ~600 lines of dead SQL code and ~400 lines of reusable code. The `view_test.go` file (1804 lines) is 50% SQL-specific tests that can be removed. This section provides a safe, incremental removal strategy with dependency ordering.

#### Code Classification: `internal/services/view.go`

| Lines | Function(s) | Status | Rationale |
|-------|-------------|--------|-----------|
| 48-157 | `initializeBuiltinViews()` | 🔴 **REPLACE** | All definitions use SQL-oriented `ViewQuery` struct. Must be rewritten with new DSL format. |
| 159-257 | `GetView`, `loadNotebookView`, `loadGlobalView` | 🟢 **KEEP** | View discovery hierarchy is backend-agnostic. |
| 258-390 | `ResolveTemplateVariables`, `resolveDayArithmetic`, `resolveWeekMonthArithmetic`, `resolveEnvironmentVariables`, date helpers | 🟢 **KEEP** | Template resolution is pure string manipulation, backend-agnostic. |
| 447-507 | `ValidateViewDefinition` | 🔴 **REPLACE** | Validates SQL-oriented fields (`Conditions`, `Having`, `AggregateColumns`). Needs new validator for DSL query string. |
| 509-566 | `validateViewCondition`, `validateHavingCondition` | 🔴 **REMOVE** | SQL condition validation, no replacement needed. |
| 567-602 | `validateAggregateFunction` | 🔴 **REMOVE** | SQL aggregates not used in new design. |
| 603-618 | `validateViewParameter` | 🟢 **KEEP** | Parameter type validation is backend-agnostic. |
| 620-677 | `ValidateParameters`, `validateParamType`, `isValidViewName` | 🟢 **KEEP** | Parameter handling is backend-agnostic. |
| 685-763 | `validateField` | 🔴 **REMOVE** | SQL field whitelist validation. |
| 764-783 | `validateOperator` | 🔴 **REMOVE** | SQL operator whitelist validation. |
| 785-834 | `ApplyParameterDefaults`, `ParseViewParameters` | 🟢 **KEEP** | Parameter handling is backend-agnostic. |
| 837-865 | `FormatQueryValue`, `escapeSQL` | 🔴 **REMOVE** | SQL value formatting and escaping. |
| 867-1032 | `GenerateSQL` | 🔴 **REMOVE** | Full SQL query generator — core dead code. |
| 1033-1070 | `convertToJSONSafe` | 🟢 **KEEP** | JSON serialization helper. |
| 1072-1151 | `GroupResults` | 🟡 **ADAPT** | Grouping logic reusable, but signature may change (takes `[]Note` not `[]map[string]interface{}`). |
| 1152-1165 | `isArray` | 🟢 **KEEP** | Utility function. |
| 1166-1196 | `transformSQLGroupedResults` | 🔴 **REMOVE** | SQL-specific result transformation. |
| 1198-1354 | `ListAllViews`, `ListBuiltinViews`, `LoadAllGlobalViews`, `LoadAllNotebookViews` | 🟢 **KEEP** | View listing and discovery is backend-agnostic. |

**Summary:**
- **REMOVE**: ~500 lines (SQL-specific code)
- **REPLACE**: ~150 lines (need new DSL versions)
- **KEEP**: ~550 lines (backend-agnostic)
- **ADAPT**: ~80 lines (minor signature changes)

#### Code Classification: `internal/services/view_test.go`

| Lines | Test Group | Status | Rationale |
|-------|------------|--------|-----------|
| 16-112 | Builtin view structure tests | 🟡 **ADAPT** | Tests check SQL-oriented structure. Update to check DSL `Query` string. |
| 113-182 | `ResolveTemplateVariables_*` (basic) | 🟢 **KEEP** | Backend-agnostic template tests. |
| 184-290 | `ValidateViewDefinition_*` | 🔴 **REMOVE** | Tests SQL-oriented validation. |
| 291-336 | `ValidateParameters_*` | 🟢 **KEEP** | Parameter validation tests (backend-agnostic). |
| 337-428 | `ValidateParamType_*`, `ApplyParameterDefaults_*` | 🟢 **KEEP** | Backend-agnostic. |
| 431-464 | `ParseViewParameters_*` | 🟢 **KEEP** | Backend-agnostic. |
| 466-504 | `FormatQueryValue_*` | 🔴 **REMOVE** | SQL value formatting tests. |
| 506-655 | `GetView_*`, `Precedence_*` | 🟢 **KEEP** | Tests view discovery logic. |
| 658-1368 | `GenerateSQL_*` (all variants) | 🔴 **REMOVE** | All test SQL generation (~710 lines). |
| 1369-1645 | More `ResolveTemplateVariables_*` | 🟢 **KEEP** | Backend-agnostic template tests. |
| 1646-1804 | `GroupResults_*` | 🟡 **ADAPT** | Grouping tests may need signature updates. |

**Summary:**
- **REMOVE**: ~900 lines (50% of test file)
- **KEEP**: ~750 lines (42%)
- **ADAPT**: ~150 lines (8%)

#### Safe Incremental Removal Strategy

**Phase 5A: Remove Dead SQL Tests First (Safe, No Production Impact)**

1. Remove `FormatQueryValue_*` tests (lines 466-504)
2. Remove `GenerateSQL_*` tests (lines 658-1368)
3. Remove `ValidateViewDefinition_*` tests (lines 184-290)
4. Run `mise run test` — all remaining tests should pass

**Phase 5B: Remove Dead SQL Implementation (Leaf Functions First)**

Remove in this order (leaf → caller dependency order):

1. `escapeSQL()` — no callers after GenerateSQL removal
2. `FormatQueryValue()` — no callers
3. `validateOperator()` — called only by validateViewCondition
4. `validateField()` — called only by validateViewCondition, validateHavingCondition, GenerateSQL
5. `validateAggregateFunction()` — called only by ValidateViewDefinition, GenerateSQL
6. `validateHavingCondition()` — called only by ValidateViewDefinition
7. `validateViewCondition()` — called only by ValidateViewDefinition
8. `transformSQLGroupedResults()` — called only by GroupResults (SQL path)
9. `GenerateSQL()` — the big one, no callers after Bleve migration
10. Update `ValidateViewDefinition()` — remove SQL-specific field checks

**Phase 5C: Replace Builtin Views (Requires New DSL Implementation)**

This depends on Phase 6 implementation, but the outline is:

1. Change `ViewDefinition.Query` from `ViewQuery` struct to `string` (in `core/view.go`)
2. Rewrite `initializeBuiltinViews()` to use new DSL format
3. Update tests in lines 16-112 to check new structure

**Phase 5D: Adapt GroupResults (Minor Signature Change)**

1. Update `GroupResults()` to accept `[]core.Note` instead of `[]map[string]interface{}`
2. Remove SQL-specific path in GroupResults
3. Update `GroupResults_*` tests

#### Dependency Graph for Removal

```
GenerateSQL
├── ValidateParameters (KEEP)
├── ApplyParameterDefaults (KEEP)
├── ResolveTemplateVariables (KEEP)
├── validateField (REMOVE)
├── validateAggregateFunction (REMOVE)
├── validateHavingCondition (REMOVE)
│   └── validateOperator (REMOVE)
└── escapeSQL (REMOVE - unused after GenerateSQL gone)

ValidateViewDefinition
├── isValidViewName (KEEP)
├── validateViewCondition (REMOVE)
│   ├── validateField (REMOVE)
│   └── validateOperator (REMOVE)
├── validateHavingCondition (REMOVE)
├── validateAggregateFunction (REMOVE)
├── validateField (REMOVE)
└── validateViewParameter (KEEP)

GroupResults
├── transformSQLGroupedResults (REMOVE)
└── isArray (KEEP)
```

#### Verification Steps

After each removal phase:

1. `mise run test` — all tests pass
2. `mise run lint` — no lint errors
3. `mise run build` — binary compiles
4. `git diff --stat` — confirm only expected files changed
5. Commit with scope: `refactor(views): remove <description>`

#### Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Accidentally remove reusable code | Follow dependency graph strictly; remove leaves first |
| Break special views | `view_special.go` is separate file, not touched in Phase 5 |
| Break view discovery | View loading functions (`GetView`, `loadNotebookView`, etc.) are explicitly KEEP |
| Break template resolution | All `Resolve*` functions are explicitly KEEP |

#### Files to Modify

| File | Action |
|------|--------|
| `internal/services/view.go` | Remove ~500 lines, replace ~150 lines |
| `internal/services/view_test.go` | Remove ~900 lines, adapt ~150 lines |
| `internal/core/view.go` | Change `ViewQuery` struct (Phase 5C, depends on Phase 6) |

#### Post-Cleanup Metrics Target

| Metric | Before | After |
|--------|--------|-------|
| `view.go` lines | 1355 | ~700 |
| `view_test.go` lines | 1804 | ~850 |
| SQL-related imports | database/sql concepts | 0 |
| Dead code | ~600 lines | 0 |

### Phase 6: Implementation Plan

**Skill Applied**: `writing-plans`

#### Deliverable

Full implementation plan created: [plan-b4e2f7a1-dsl-views-implementation.md](plan-b4e2f7a1-dsl-views-implementation.md)

#### Plan Summary

The implementation plan contains **10 tasks** following TDD principles:

| Task | Description | Files Changed |
|------|-------------|---------------|
| 1 | Add `has:` and `missing:` keywords to DSL grammar | `internal/search/parser/` |
| 2 | Implement quote-aware pipe splitting | `internal/services/view_query.go` |
| 3 | Implement directives parser | `internal/services/view_query.go` |
| 4 | Update `ViewDefinition` type | `internal/core/view.go` |
| 5 | Rewrite builtin views with DSL | `internal/services/view.go` |
| 6 | Implement view query execution | `internal/services/view_executor.go` |
| 7 | Remove dead SQL code | `internal/services/view.go` |
| 8 | Update CLI `notes view` command | `cmd/notes_view.go` |
| 9 | Add pipe syntax to `notes search` | `cmd/notes_search.go` |
| 10 | Integration testing | `internal/services/view_integration_test.go` |

#### Key Implementation Decisions

1. **Grammar changes**: Only add `has:` and `missing:` keywords — no other Participle changes
2. **Pipe splitting**: Quote-aware splitting before DSL parsing
3. **Directives**: Simple `key:value` parser (~50 lines), not Participle
4. **ViewDefinition**: Replace `ViewQuery` struct with `string` field
5. **Execution**: Route through existing `Index.Find()` via `FindOpts`
6. **Special views**: Unchanged — dispatch via `Type: "special"`

#### Estimated Effort

- Total tasks: 10
- Estimated time: 4-6 hours (experienced developer)
- Each task is self-contained and can be committed independently

#### Execution Options

1. **Subagent-Driven (recommended)**: Fresh subagent per task with code review between tasks
2. **Parallel Session**: Open new session with `executing-plans` skill

## References

- DSL grammar: `internal/search/parser/grammar.go`
- DSL parser: `internal/search/parser/parser.go`
- Search service (query building): `internal/services/search.go`
- View service: `internal/services/view.go`
- View types: `internal/core/view.go`
- Special views: `internal/services/view_special.go`
- View command: `cmd/notes_view.go`
