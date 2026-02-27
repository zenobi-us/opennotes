# Notes Search Command

Search notes with title-focused fieldless matching, boolean filters, DSL pipe syntax, or semantic retrieval.

## Overview

Jot supports these search workflows:

1. **Fieldless Search (title-only)** (`jot notes search "..."`)  
   Fieldless input is normalized to `title:<query>` (or `title:"multi word"`).
   Use explicit DSL (for example `body:query`) to search outside title.
2. **Boolean Query Search** (`jot notes search query`)  
   Structured `--and/--or/--not` filtering on metadata, path, title, and links.
3. **DSL Pipe Syntax** (`jot notes search "filter | directives"`)  
   Filter expression + sort/limit/offset directives.
4. **Semantic Search** (`jot notes search semantic`)  
   Keyword / semantic / hybrid retrieval with optional explain output.

---

## Quick Start

```bash
# Fieldless title search
jot notes search "meeting"

# Boolean query
jot notes search query --and data.tag=workflow --not data.status=archived

# DSL + directives
jot notes search "tag:work | sort:modified:desc limit:10"

# Semantic hybrid search
jot notes search semantic "project planning discussions"
```

---

## Fieldless Search (Title-Only)

### Syntax

```bash
jot notes search [query]
```

### Examples

```bash
jot notes search "meeting"                 # normalized to title:meeting
jot notes search "fieldless text"          # normalized to title:"fieldless text"
jot notes search "body:meeting"            # explicit body search
jot notes search "todo" --notebook ~/notes
jot notes search
```

### Behavior

- Case-insensitive matching
- Fieldless queries search title only
- To search body/content, use explicit DSL field `body:<text>`
- No query returns all notes

---

## Boolean Query Search (`query` subcommand)

> ⚠️ **Deprecated path**
>
> `jot notes search query --and/--or/--not` is deprecated since `TODO(vX.Y.Z)` and scheduled for removal in `TODO(vX.(Y+1).0)`.
> Track rollout in `.memory/epic-9b7c2a4d-unified-search-dsl-deprecation.md`.
>
> **Migration target:** unified DSL via `jot notes search "<dsl>"`.

### Syntax

```bash
jot notes search query [--and field=value] [--or field=value] [--not field=value]
```

### Operators

| Operator | Meaning                   |
| -------- | ------------------------- |
| `--and`  | all conditions must match |
| `--or`   | any condition can match   |
| `--not`  | excludes matching notes   |

### Supported Fields

#### Metadata (`data.*`)

- `data.tag`, `data.tags`
- `data.status`, `data.priority`
- `data.assignee`, `data.author`
- `data.type`, `data.category`
- `data.project`, `data.sprint`

#### File/Title

- `path` (glob-enabled)
- `title`

#### Link Graph

- `links-to`
- `linked-by`

### Examples

```bash
# Basic metadata filtering
jot notes search query --and data.tag=workflow --and data.status=active

# OR logic
jot notes search query --or data.priority=high --or data.priority=critical

# Path filtering
jot notes search query --and path=projects/**/*.md --not path=archive/*

# Link graph filtering
jot notes search query --and links-to=docs/architecture.md
jot notes search query --and linked-by=planning/q1.md
```

### Migration to Unified DSL (recommended)

```bash
# AND conditions
jot notes search query --and data.tag=workflow --and data.status=active
jot notes search "tag:workflow status:active"

# OR conditions
jot notes search query --or data.priority=high --or data.priority=critical
jot notes search "priority:high OR priority:critical"

# Path + exclusion
jot notes search query --and path=projects/**/*.md --not path=archive/*
jot notes search "path:projects/** NOT path:archive/*"
```

---

## DSL Pipe Syntax (`search [query]`)

Use this when you want inline sorting/pagination directives.

### Syntax

```bash
jot notes search "<filter> | <directives>"
```

### Filter Examples

```bash
jot notes search "tag:work | sort:modified:desc"
jot notes search "status:todo | sort:created:asc limit:20"
jot notes search "| sort:title:asc"
```

### Filter Fields (DSL side)

- `tag:<value>`
- `status:<value>`
- `type:<value>`
- `title:<text>`
- `body:<text>`
- `path:<glob-or-prefix>`
- `created:>date`, `created:<date`
- `modified:>date`, `modified:<date`

### Boolean Operators

Use `OR` (case-insensitive) to join filters. Implicit AND still applies between whitespace-separated expressions and has higher precedence than OR.

```bash
# Either work or personal tagged notes
jot notes search "tag:work OR tag:personal"

# (tag:work AND status:todo) OR status:done
jot notes search "tag:work status:todo OR status:done"
```

### Directives

- `sort:<field>:<dir>` where field is `modified|created|title|path`, dir is `asc|desc`
- `limit:<n>`
- `offset:<n>`

---

## Semantic Search (`semantic` subcommand)

### Syntax

```bash
jot notes search semantic [query] [flags]
```

### Key Flags

- `--mode hybrid|keyword|semantic` (default: `hybrid`)
- `--top-k <n>` (default: `100`)
- `--explain`
- `--and`, `--or`, `--not` (same condition format as `query`)

### Examples

```bash
jot notes search semantic "meeting notes"
jot notes search semantic "workflow" --mode keyword --and data.status=active
jot notes search semantic "architecture" --mode hybrid --explain
```

For full semantic behavior and troubleshooting, see:

- [Semantic Search Guide](../semantic-search-guide.md)

---

## Glob Patterns

`path`, `links-to`, and `linked-by` support globs.

| Pattern | Meaning                  | Example     |
| ------- | ------------------------ | ----------- |
| `*`     | any chars (single level) | `docs/*.md` |
| `**`    | recursive path depth     | `**/*.md`   |
| `?`     | single character         | `task?.md`  |

---

## Common Errors

| Error                                | Cause                             | Fix                               |
| ------------------------------------ | --------------------------------- | --------------------------------- |
| `invalid field: X`                   | Unsupported field name            | Use supported fields listed above |
| `expected field=value`               | Missing `=`                       | Use `field=value` format          |
| `value too long`                     | Condition value exceeds limit     | Shorten value                     |
| `at least one condition is required` | `query` called with no conditions | Add `--and`, `--or`, or `--not`   |

---

## Related Docs

- [Semantic Search Guide](../semantic-search-guide.md)
- [Getting Started (Power Users)](../getting-started-power-users.md)
- [Notebook Discovery](../notebook-discovery.md)
