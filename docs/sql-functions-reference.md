# SQL Functions Reference

Quick reference for DuckDB markdown extension functions available in OpenNotes.

## Table Functions

### `read_markdown(glob, ...)`

Reads markdown files matching a glob pattern.

**Syntax:**
```sql
read_markdown(glob_pattern, include_filepath := true)
```

**Parameters:**
- `glob_pattern` (string): File pattern (e.g., `**/*.md`, `notes/*.md`)
- `include_filepath` (boolean, optional): Include filepath column (default: false)

**Returns:**
- `content` (string): Full markdown content including frontmatter
- `filepath` (string): Absolute file path (if include_filepath=true)  
- `metadata` (map): Frontmatter parsed as key-value pairs

**Examples:**
```sql
-- Basic usage
SELECT * FROM read_markdown('**/*.md')

-- With file paths
SELECT * FROM read_markdown('**/*.md', include_filepath := true)

-- Specific directory
SELECT * FROM read_markdown('projects/*.md', include_filepath := true)
```

## Scalar Functions

### `md_stats(content)`

Returns statistics about markdown content.

**Syntax:**
```sql
md_stats(content_string)
```

**Parameters:**
- `content_string` (string): Markdown content to analyze

**Returns:** Struct with:
- `word_count` (integer): Number of words
- `character_count` (integer): Number of characters  
- `line_count` (integer): Number of lines

**Examples:**
```sql
-- Get word count
SELECT (md_stats(content)).word_count as words
FROM read_markdown('**/*.md')

-- Get all stats
SELECT 
    (md_stats(content)).word_count as words,
    (md_stats(content)).character_count as chars,
    (md_stats(content)).line_count as lines
FROM read_markdown('**/*.md')

-- Filter by word count
SELECT * FROM read_markdown('**/*.md', include_filepath := true)
WHERE (md_stats(content)).word_count > 500
```

### `md_extract_links(content)`

Extracts all markdown links from content.

**Syntax:**
```sql
md_extract_links(content_string)
```

**Parameters:**
- `content_string` (string): Markdown content to analyze

**Returns:** Array of structs with:
- `text` (string): Link text/title
- `url` (string): Link URL/destination

**Examples:**
```sql
-- Extract all links
SELECT UNNEST(md_extract_links(content)) as link_info
FROM read_markdown('**/*.md')

-- Get link text and URLs separately  
SELECT 
    filepath,
    link.text,
    link.url
FROM read_markdown('**/*.md', include_filepath := true),
     LATERAL UNNEST(md_extract_links(content)) AS link

-- Count links per file
SELECT 
    filepath,
    array_length(md_extract_links(content)) as link_count
FROM read_markdown('**/*.md', include_filepath := true)
```

### `md_extract_code_blocks(content)`

Extracts code blocks from markdown content.

**Syntax:**
```sql
md_extract_code_blocks(content_string)
```

**Parameters:**
- `content_string` (string): Markdown content to analyze

**Returns:** Array of structs with:
- `language` (string): Programming language (if specified)
- `code` (string): Code content

**Examples:**
```sql
-- Extract all code blocks
SELECT UNNEST(md_extract_code_blocks(content)) as code_info  
FROM read_markdown('**/*.md')

-- Get Python code only
SELECT 
    filepath,
    cb.code
FROM read_markdown('**/*.md', include_filepath := true),
     LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
WHERE cb.language = 'python'

-- Count code blocks by language
SELECT 
    cb.language,
    COUNT(*) as count
FROM read_markdown('**/*.md'),
     LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
GROUP BY cb.language
ORDER BY count DESC
```

### `md_extract_headers(content)`

Extracts headers from markdown content.

**Syntax:**
```sql
md_extract_headers(content_string)
```

**Parameters:**
- `content_string` (string): Markdown content to analyze

**Returns:** Array of structs with:
- `level` (integer): Header level (1-6)
- `text` (string): Header text

**Examples:**
```sql
-- Extract all headers
SELECT UNNEST(md_extract_headers(content)) as header_info
FROM read_markdown('**/*.md')

-- Get top-level headers only
SELECT 
    filepath,
    header.text
FROM read_markdown('**/*.md', include_filepath := true),
     LATERAL UNNEST(md_extract_headers(content)) AS header  
WHERE header.level = 1

-- Count headers by level
SELECT 
    header.level,
    COUNT(*) as count
FROM read_markdown('**/*.md'),
     LATERAL UNNEST(md_extract_headers(content)) AS header
GROUP BY header.level
ORDER BY header.level
```

## Standard SQL Functions

These standard SQL functions are particularly useful with markdown content:

### String Functions

#### `LIKE` and `ILIKE`
```sql
-- Case-sensitive pattern matching
SELECT * FROM read_markdown('**/*.md') WHERE content LIKE '%TODO%'

-- Case-insensitive pattern matching  
SELECT * FROM read_markdown('**/*.md') WHERE content ILIKE '%todo%'
```

#### `LOWER()` and `UPPER()`
```sql
-- Convert to lowercase for comparison
SELECT * FROM read_markdown('**/*.md') 
WHERE LOWER(content) LIKE '%meeting%'
```

#### `LENGTH()`
```sql
-- Content length
SELECT filepath, LENGTH(content) as content_length
FROM read_markdown('**/*.md', include_filepath := true)
```

#### `SUBSTRING()`
```sql
-- First 100 characters of content
SELECT filepath, SUBSTRING(content, 1, 100) as preview
FROM read_markdown('**/*.md', include_filepath := true)
```

#### `SPLIT()` and `STRING_SPLIT()`
```sql
-- Split metadata tags
SELECT 
    filepath,
    UNNEST(string_split(metadata['tags'], ',')) as tag
FROM read_markdown('**/*.md', include_filepath := true)
WHERE metadata['tags'] IS NOT NULL
```

### Array Functions

#### `array_length()`
```sql
-- Count links per file
SELECT 
    filepath,
    array_length(md_extract_links(content)) as link_count
FROM read_markdown('**/*.md', include_filepath := true)
```

#### `UNNEST()`
```sql
-- Expand arrays into rows
SELECT 
    filepath,
    UNNEST(md_extract_links(content)) as link
FROM read_markdown('**/*.md', include_filepath := true)
```

### Aggregate Functions

#### `COUNT()`, `SUM()`, `AVG()`
```sql
-- Statistics across all notes
SELECT 
    COUNT(*) as total_notes,
    AVG((md_stats(content)).word_count) as avg_words,
    SUM((md_stats(content)).word_count) as total_words
FROM read_markdown('**/*.md')
```

#### `MAX()`, `MIN()`
```sql
-- Find longest and shortest notes
SELECT 
    MAX((md_stats(content)).word_count) as longest_note,
    MIN((md_stats(content)).word_count) as shortest_note
FROM read_markdown('**/*.md')
```

### Date Functions

#### `CURRENT_DATE`, `NOW()`
```sql
-- Recent notes (if dates in metadata)
SELECT * FROM read_markdown('**/*.md', include_filepath := true)
WHERE metadata['date']::DATE >= CURRENT_DATE - INTERVAL 7 DAY
```

## Working with Metadata

Frontmatter metadata is accessible as a map. Common patterns:

### Access Fields
```sql
-- Access specific metadata fields
SELECT 
    filepath,
    metadata['title'] as title,
    metadata['author'] as author,
    metadata['tags'] as tags,
    metadata['date'] as date
FROM read_markdown('**/*.md', include_filepath := true)
```

### Check for Fields
```sql
-- Notes with titles
SELECT * FROM read_markdown('**/*.md', include_filepath := true)
WHERE metadata['title'] IS NOT NULL

-- Notes with tags
SELECT * FROM read_markdown('**/*.md', include_filepath := true)  
WHERE metadata['tags'] IS NOT NULL
```

### Type Conversion
```sql
-- Convert metadata to appropriate types
SELECT 
    filepath,
    metadata['date']::DATE as date,
    metadata['word_goal']::INTEGER as word_goal
FROM read_markdown('**/*.md', include_filepath := true)
WHERE metadata['date'] IS NOT NULL
```

## Common Patterns

### Content Search
```sql
-- Case-insensitive content search
SELECT filepath FROM read_markdown('**/*.md', include_filepath := true)
WHERE LOWER(content) LIKE '%search_term%'

-- Multiple search terms (AND)
SELECT filepath FROM read_markdown('**/*.md', include_filepath := true)
WHERE content LIKE '%term1%' AND content LIKE '%term2%'

-- Multiple search terms (OR)  
SELECT filepath FROM read_markdown('**/*.md', include_filepath := true)
WHERE content LIKE '%term1%' OR content LIKE '%term2%'
```

### Statistics and Analysis
```sql
-- Word count distribution
SELECT 
    CASE 
        WHEN (md_stats(content)).word_count < 100 THEN 'Short'
        WHEN (md_stats(content)).word_count < 500 THEN 'Medium'  
        ELSE 'Long'
    END as length_category,
    COUNT(*) as note_count
FROM read_markdown('**/*.md')
GROUP BY length_category
```

### Complex Queries with CTEs
```sql
-- Using Common Table Expressions for complex analysis
WITH note_stats AS (
    SELECT 
        filepath,
        (md_stats(content)).word_count as words,
        array_length(md_extract_links(content)) as links,
        array_length(md_extract_code_blocks(content)) as code_blocks
    FROM read_markdown('**/*.md', include_filepath := true)
)
SELECT * FROM note_stats 
WHERE words > 1000 AND (links > 5 OR code_blocks > 2)
```

## Function Reference Quick Lookup

| Function | Purpose | Returns |
|----------|---------|---------|
| `read_markdown()` | Read markdown files | Table: content, filepath, metadata |
| `md_stats()` | Content statistics | Struct: word_count, character_count, line_count |
| `md_extract_links()` | Extract links | Array: [{text, url}, ...] |
| `md_extract_code_blocks()` | Extract code | Array: [{language, code}, ...] |
| `md_extract_headers()` | Extract headers | Array: [{level, text}, ...] |

## Error Reference

| Error | Cause | Solution |
|-------|-------|----------|
| "File or directory does not exist" | No files match glob | Check file pattern |
| "Invalid glob pattern" | Malformed pattern | Use proper glob syntax |
| "Function does not exist" | Typo in function name | Check function spelling |
| "Cannot access field" | Invalid metadata field | Check available metadata keys |

---

For more detailed examples and usage patterns, see the [SQL Guide](sql-guide.md).