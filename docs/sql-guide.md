# SQL Query Guide for OpenNotes

This guide shows you how to use the `--sql` flag to run custom SQL queries against your notebook files using DuckDB's powerful markdown extension.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Available Functions](#available-functions)
3. [Schema Overview](#schema-overview)
4. [Common Query Patterns](#common-query-patterns)
5. [Troubleshooting](#troubleshooting)
6. [Security Model](#security-model)
7. [Performance Tips](#performance-tips)

## Getting Started

### Basic Syntax

Use the `--sql` flag with the search command:

```bash
opennotes search --sql "SELECT * FROM read_markdown('**/*.md') LIMIT 5"
```

### Your First Query

List all notes in your notebook:

```bash
opennotes search --sql "SELECT filepath, content FROM read_markdown('**/*.md', include_filepath:=true)"
```

**Output format:**
```
filepath                    content
-------------------------   ------------------------------------------
notes/project-ideas.md      # Project Ideas\n\nSome ideas for new...
notes/meeting-notes.md      # Meeting Notes\n\nDiscussed the new...
notes/todo.md              # Todo List\n\n- [ ] Finish report...

3 rows
```

## Available Functions

### Table Functions

#### `read_markdown(glob, include_filepath:=boolean)`
Reads markdown files matching the glob pattern.

**Parameters:**
- `glob` (string): File pattern (e.g., `**/*.md`, `notes/*.md`)
- `include_filepath` (boolean, optional): Include filepath column

**Returns:** Table with columns:
- `content` (string): Full markdown content
- `filepath` (string): Full file path (if include_filepath=true)
- `metadata` (map): Frontmatter as key-value pairs

**Examples:**
```sql
-- All notes with file paths
SELECT * FROM read_markdown('**/*.md', include_filepath:=true)

-- Notes in specific directory
SELECT * FROM read_markdown('projects/*.md')

-- Single file
SELECT * FROM read_markdown('README.md')
```

### Scalar Functions

#### `md_stats(content)`
Returns statistics about markdown content.

**Returns:** Struct with:
- `word_count` (integer): Number of words
- `character_count` (integer): Number of characters
- `line_count` (integer): Number of lines

**Example:**
```sql
SELECT 
    filepath,
    (md_stats(content)).word_count as words,
    (md_stats(content)).line_count as lines
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE (md_stats(content)).word_count < 100
```

#### `md_extract_links(content)`
Extracts all markdown links from content.

**Returns:** Array of structs with:
- `text` (string): Link text
- `url` (string): Link URL

**Example:**
```sql
SELECT 
    filepath,
    UNNEST(md_extract_links(content)) as link_info
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE array_length(md_extract_links(content)) > 0
```

#### `md_extract_code_blocks(content)`
Extracts code blocks from content.

**Returns:** Array of structs with:
- `language` (string): Programming language
- `code` (string): Code content

**Example:**
```sql
SELECT 
    filepath,
    cb.language,
    cb.code
FROM read_markdown('**/*.md', include_filepath:=true),
     LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
WHERE cb.language = 'python'
```

## Schema Overview

### `read_markdown()` Columns

| Column | Type | Description |
|--------|------|-------------|
| `content` | string | Full markdown content including frontmatter |
| `filepath` | string | Absolute file path (if include_filepath=true) |
| `metadata` | map | Frontmatter parsed as key-value pairs |

### Frontmatter Access

Access frontmatter fields using map syntax:

```sql
SELECT 
    filepath,
    metadata['title'] as title,
    metadata['tags'] as tags,
    metadata['date'] as date
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE metadata['title'] IS NOT NULL
```

## Common Query Patterns

### 1. Find Notes by Content

```sql
-- Case-insensitive search
SELECT filepath FROM read_markdown('**/*.md', include_filepath:=true)
WHERE LOWER(content) LIKE '%meeting%'

-- Multiple keywords (AND)
SELECT filepath FROM read_markdown('**/*.md', include_filepath:=true)
WHERE content LIKE '%project%' AND content LIKE '%deadline%'

-- Multiple keywords (OR)
SELECT filepath FROM read_markdown('**/*.md', include_filepath:=true)
WHERE content LIKE '%todo%' OR content LIKE '%task%'
```

### 2. Find Notes by Metadata

```sql
-- Notes with specific tag
SELECT filepath, metadata['title'] as title
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE metadata['tags'] LIKE '%work%'

-- Recent notes
SELECT filepath, metadata['date'] as date
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE metadata['date'] >= '2024-01-01'
ORDER BY metadata['date'] DESC

-- Notes by author
SELECT filepath, metadata['author'] as author
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE metadata['author'] = 'John Doe'
```

### 3. Analyze Content Statistics

```sql
-- Word count analysis
SELECT 
    filepath,
    (md_stats(content)).word_count as words
FROM read_markdown('**/*.md', include_filepath:=true)
ORDER BY words DESC
LIMIT 10

-- Average words per note
SELECT 
    COUNT(*) as note_count,
    AVG((md_stats(content)).word_count) as avg_words,
    MAX((md_stats(content)).word_count) as max_words
FROM read_markdown('**/*.md')

-- Find short notes that need more content
SELECT filepath, (md_stats(content)).word_count as words
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE (md_stats(content)).word_count < 50
ORDER BY words ASC
```

### 4. Code Block Analysis

```sql
-- Find all Python code
SELECT 
    filepath,
    cb.code
FROM read_markdown('**/*.md', include_filepath:=true),
     LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
WHERE cb.language = 'python'

-- Count code blocks by language
SELECT 
    cb.language,
    COUNT(*) as block_count
FROM read_markdown('**/*.md'),
     LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
GROUP BY cb.language
ORDER BY block_count DESC

-- Find notes with most code blocks
SELECT 
    filepath,
    array_length(md_extract_code_blocks(content)) as code_blocks
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE array_length(md_extract_code_blocks(content)) > 0
ORDER BY code_blocks DESC
```

### 5. Link Analysis

```sql
-- Find all external links
SELECT 
    filepath,
    link.text,
    link.url
FROM read_markdown('**/*.md', include_filepath:=true),
     LATERAL UNNEST(md_extract_links(content)) AS link
WHERE link.url LIKE 'http%'

-- Find broken internal links (simple check)
SELECT 
    filepath,
    link.url as potentially_broken
FROM read_markdown('**/*.md', include_filepath:=true),
     LATERAL UNNEST(md_extract_links(content)) AS link
WHERE link.url LIKE './%' OR link.url LIKE '../%'

-- Count links per note
SELECT 
    filepath,
    array_length(md_extract_links(content)) as link_count
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE array_length(md_extract_links(content)) > 5
ORDER BY link_count DESC
```

### 6. Complex Analysis with CTEs

```sql
-- Most active writing days
WITH daily_stats AS (
    SELECT 
        metadata['date'] as date,
        COUNT(*) as notes_written,
        SUM((md_stats(content)).word_count) as words_written
    FROM read_markdown('**/*.md')
    WHERE metadata['date'] IS NOT NULL
    GROUP BY metadata['date']
)
SELECT * FROM daily_stats
WHERE words_written > 1000
ORDER BY date DESC

-- Tag analysis
WITH tag_stats AS (
    SELECT 
        UNNEST(string_split(metadata['tags'], ',')) as tag,
        filepath
    FROM read_markdown('**/*.md', include_filepath:=true)
    WHERE metadata['tags'] IS NOT NULL
)
SELECT 
    TRIM(tag) as tag,
    COUNT(*) as usage_count
FROM tag_stats
GROUP BY TRIM(tag)
ORDER BY usage_count DESC
```

## Troubleshooting

### Common Errors

#### "Only SELECT queries are allowed"
**Cause:** Trying to use INSERT, UPDATE, DELETE, or other write operations.
**Solution:** Use only SELECT or WITH statements.

```bash
# ❌ This fails
opennotes search --sql "DELETE FROM markdown"

# ✅ This works
opennotes search --sql "SELECT * FROM read_markdown('**/*.md')"
```

#### "File or directory does not exist"
**Cause:** No files match your glob pattern.
**Solution:** Check your file pattern and notebook path.

```bash
# ❌ No .md files found
opennotes search --sql "SELECT * FROM read_markdown('*.txt')"

# ✅ Correct pattern
opennotes search --sql "SELECT * FROM read_markdown('**/*.md')"
```

#### "keyword 'DROP' is not allowed"
**Cause:** Using blocked dangerous keywords.
**Solution:** Remove dangerous keywords from your query.

#### Query times out after 30 seconds
**Cause:** Query is too complex or dataset too large.
**Solution:** Add LIMIT clauses or simplify the query.

```bash
# ❌ May timeout on large notebooks
opennotes search --sql "SELECT * FROM read_markdown('**/*.md')"

# ✅ Limited results
opennotes search --sql "SELECT * FROM read_markdown('**/*.md') LIMIT 100"
```

### Debug Tips

1. **Start simple:** Begin with `SELECT * FROM read_markdown('**/*.md') LIMIT 5`
2. **Check your pattern:** Use specific glob patterns to limit files
3. **Use LIMIT:** Always limit results during testing
4. **Check metadata:** Use `SELECT metadata FROM read_markdown('**/*.md') LIMIT 1` to see available fields

## Security Model

### Read-Only Access
- Only SELECT and WITH (Common Table Expression) queries allowed
- No data modification possible (INSERT, UPDATE, DELETE blocked)
- No schema changes possible (CREATE, ALTER, DROP blocked)
- No system access (PRAGMA, ATTACH blocked)

### Validation
Queries are validated before execution to block:
- `INSERT`, `UPDATE`, `DELETE`
- `DROP`, `CREATE`, `ALTER`
- `TRUNCATE`, `REPLACE`
- `ATTACH`, `DETACH`
- `PRAGMA`

### Timeout Protection
- All queries have a 30-second timeout
- Prevents runaway queries from blocking the system
- Long-running queries are automatically cancelled

### Isolation
- Separate read-only database connection
- No access to OpenNotes internal tables
- Cannot affect notebook files on disk

## Performance Tips

### Use Specific Glob Patterns
```sql
-- ❌ Slow: searches all files
SELECT * FROM read_markdown('**/*')

-- ✅ Fast: searches only markdown files
SELECT * FROM read_markdown('**/*.md')

-- ✅ Faster: searches specific directory
SELECT * FROM read_markdown('work-notes/*.md')
```

### Limit Results
```sql
-- ❌ Returns all results (potentially thousands)
SELECT * FROM read_markdown('**/*.md')

-- ✅ Returns manageable number
SELECT * FROM read_markdown('**/*.md') LIMIT 50
```

### Filter Early
```sql
-- ❌ Processes all notes then filters
SELECT * FROM (
    SELECT *, (md_stats(content)).word_count as words
    FROM read_markdown('**/*.md')
) WHERE words > 1000

-- ✅ Filters during processing
SELECT *, (md_stats(content)).word_count as words
FROM read_markdown('**/*.md')
WHERE (md_stats(content)).word_count > 1000
```

### Use Appropriate Indexes
DuckDB automatically optimizes many queries, but you can help by:
- Filtering on metadata fields early
- Using specific patterns instead of broad searches
- Limiting result sets

---

## More Resources

- [DuckDB SQL Reference](https://duckdb.org/docs/sql/introduction)
- [DuckDB Markdown Extension](https://github.com/duckdb/duckdb/blob/main/extension/markdown/README.md)
- [OpenNotes CLI Reference](./cli-reference.md)

Need help? Check the troubleshooting section above or open an issue on GitHub.