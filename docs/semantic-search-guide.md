# Semantic Search Guide

Jot includes an optional **semantic search** capability that finds notes by meaning rather than exact keyword matching. This is useful when you want to find conceptually related content.

## Overview

| Search Mode              | Best For                        | How It Works                               |
| ------------------------ | ------------------------------- | ------------------------------------------ |
| **Keyword** (default)    | Exact terms, quick lookups      | Bleve full-text index with BM25 ranking    |
| **Semantic**             | Conceptual queries, paraphrases | Vector embeddings via chromem-go           |
| **Hybrid** (recommended) | Best of both worlds             | Combines keyword + semantic with RRF merge |

## Quick Start

```bash
# Hybrid search (default mode)
jot notes search semantic "project planning discussions"

# Keyword-only mode (fastest, exact matching)
jot notes search semantic "meeting" --mode keyword

# Semantic-only mode (meaning-based)
jot notes search semantic "discussions about deadlines" --mode semantic

# With explain output to understand matches
jot notes search semantic "architecture" --explain
```

## When to Use Each Mode

### Regular Search (`notes search`)

- Fieldless lookups when you know title terms (`jot notes search "..."` => `title:...`)
- Explicit body/content lookups with `body:`
- **Performance-critical** scenarios (fastest)

```bash
jot notes search "TODO"
jot notes search "body:2024-01-15"
```

### Semantic Search (`notes search semantic`)

- **Conceptual queries**: "notes about team communication"
- **Paraphrases**: Finding "meeting notes" when content says "sync discussion"
- **Exploratory search**: When you don't know the exact terms used

```bash
jot notes search semantic "improving team workflow"
jot notes search semantic "architecture decisions"
```

## Command Reference

### Basic Syntax

```bash
jot notes search semantic [query] [flags]
```

### Flags

| Flag        | Default  | Description                                        |
| ----------- | -------- | -------------------------------------------------- |
| `--mode`    | `hybrid` | Retrieval mode: `hybrid`, `keyword`, or `semantic` |
| `--explain` | `false`  | Show match type and explanation for each result    |
| `--top-k`   | `100`    | Maximum candidates per source before merge         |
| `--and`     | -        | AND condition (field=value) - all must match       |
| `--or`      | -        | OR condition (field=value) - any can match         |
| `--not`     | -        | NOT condition (field=value) - excludes matches     |

### Examples with Filters

```bash
# Semantic search with tag filter
jot notes search semantic "project updates" --and data.tag=work

# Exclude archived notes
jot notes search semantic "meeting notes" --not data.status=archived

# Combine multiple filters
jot notes search semantic "decisions" \
  --and data.tag=architecture \
  --not data.status=archived
```

## Understanding Explain Output

The `--explain` flag shows how each result was matched:

```bash
jot notes search semantic "workflow automation" --explain
```

Output includes:

- **Match type**: `keyword`, `semantic`, or `both` (appeared in both retrievals)
- **Explanation**: Snippet showing why the note matched

Example output:

```
Found 3 note(s) using hybrid mode (explain):

[both] Meeting Notes - Project Sync
  File: notes/meetings/2024-01-15.md
  Why: "...discussed workflow automation for the build pipeline..."

[semantic] Development Processes
  File: notes/dev/processes.md
  Why: Semantic similarity to "workflow automation"

[keyword] CI/CD Setup Guide
  File: notes/dev/cicd.md
  Why: Contains "automation" keyword
```

## How It Works

### Hybrid Mode (Default)

1. **Keyword retrieval**: Searches Bleve full-text index
2. **Semantic retrieval**: Queries vector embeddings via chromem-go
3. **RRF merge**: Combines results using Reciprocal Rank Fusion
4. **Deduplication**: Notes appearing in both sources get boosted and labeled `both`

### Fallback Behavior

If the semantic backend is unavailable:

- **Hybrid mode**: Falls back to keyword-only results with a warning
- **Semantic mode**: Returns error suggesting `--mode keyword`

## Performance Considerations

| Mode     | Typical Latency | Notes                                         |
| -------- | --------------- | --------------------------------------------- |
| Keyword  | <5ms            | Fastest, uses Bleve index                     |
| Semantic | 50-200ms        | Depends on embedding model and notebook size  |
| Hybrid   | 50-200ms        | Parallel retrieval, slight overhead for merge |

### Tips for Large Notebooks

- Use **keyword mode** for quick lookups when you know the terms
- Use **hybrid mode** for exploratory searches
- Semantic indexing happens automatically on notebook operations
- First search after adding many notes may be slower due to indexing

## Troubleshooting

### "Semantic backend unavailable"

The semantic backend (chromem-go) may not be initialized. Try:

```bash
# Use keyword-only mode
jot notes search semantic "query" --mode keyword
```

### No results in semantic mode

- Try **hybrid mode** which combines keyword matching
- Check that your query is conceptual rather than very specific keywords
- Verify notes exist with `jot notes list`

### Different results between modes

This is expected! Each mode has different strengths:

- **Keyword** finds exact term matches
- **Semantic** finds conceptually similar content
- **Hybrid** combines both for balanced results

## Related Documentation

- [Getting Started](getting-started-basics.md)
- [Boolean Query Reference](getting-started-power-users.md)
- [Views System](views-guide.md)
