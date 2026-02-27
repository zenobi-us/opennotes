# Notes Search Command

Search notes with title-focused fieldless matching, boolean filters, DSL pipe syntax, or semantic retrieval.

## Overview

Jot supports these search workflows:

1. **Fieldless Search (title-only)** (`jot notes search "..."`)  
   Fieldless input is normalized to `title:<query>` (or `title:"multi word"`).
   Use explicit DSL (for example `body:query`) to search outside title.
2. **DSL Filter + Directives** (`jot notes search "filter | directives"`)  
   Filter expression with optional sort/limit/offset directives.
3. **Semantic Search** (`jot notes search semantic`)  
   Keyword / semantic / hybrid retrieval with optional explain output and `--and/--or/--not` filters.

---

## Quick Start

```bash
# Fieldless title search
jot notes search "meeting"

# DSL boolean operators
jot notes search "tag:workflow NOT status:archived"

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
| `invalid field: X`        | Unsupported field name                | Use supported fields listed above |
| `failed to parse filter`  | Invalid DSL filter syntax             | Fix DSL expression (`field:value`) |
| `value too long`          | Semantic condition value exceeds limit | Shorten value                      |

---

## Related Docs

- [Semantic Search Guide](../semantic-search-guide.md)
- [Getting Started (Power Users)](../getting-started-power-users.md)
- [Notebook Discovery](../notebook-discovery.md)
