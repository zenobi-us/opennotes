# SQL Quick Reference for Jot

This guide provides a progressive learning path for querying your notes with SQL. Start simple, build confidence, then tackle advanced patterns. All examples are tested and ready to copy-paste.

## Table of Contents

1. [SQL Basics for Note Queries](#sql-basics-for-note-queries)
2. [Progressive Learning Path](#progressive-learning-path)
3. [Level 1: Basic Queries](#level-1-basic-queries)
4. [Level 2: Content Search](#level-2-content-search)
5. [Level 3: Metadata Analysis](#level-3-metadata-analysis)
6. [Level 4: Complex Queries](#level-4-complex-queries)
7. [DuckDB Markdown Functions](#duckdb-markdown-functions)
8. [Performance Tips](#performance-tips)
9. [Common Mistakes](#common-mistakes)
10. [When to Use SQL](#when-to-use-sql)

---

## SQL Basics for Note Queries

### What is SQL?

SQL is a language for querying structured data. Jot uses **DuckDB**, an embedded SQL engine, to query your markdown files.

Think of SQL as powerful filtering and analysis that would take hours in a text editor—but executes in milliseconds.

### SELECT Statement Structure

All SQL queries follow this pattern:

```sql
SELECT <what-to-return>
FROM <where-to-get-data>
WHERE <filters>
ORDER BY <sort-order>
LIMIT <how-many>
```

**Components explained**:
- **SELECT**: Which columns to return
- **FROM**: Which table/function to query (usually `read_markdown()`)
- **WHERE**: Filter conditions (optional)
- **ORDER BY**: Sort order (optional)
- **LIMIT**: Maximum results (optional)

### read_markdown() Function

`read_markdown()` is your entry point—it reads all markdown files matching a pattern:

```bash
# Basic syntax
jot notes search --sql "SELECT * FROM read_markdown('**/*.md')"

# With file paths included
jot notes search --sql "SELECT * FROM read_markdown('**/*.md', include_filepath:=true)"
```

**Parameters**:
- `'**/*.md'` - File pattern (required)
  - `'*.md'` - All .md in notebook root only
  - `'**/*.md'` - All .md recursively (includes subdirectories)
  - `'projects/*.md'` - Only in projects folder
- `include_filepath:=true` - Include file path in results (optional)

### Your First Query

List all your notes:

```bash
jot notes search --sql "SELECT * FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 10"
```

Output:
```json
[
  {
    "file_path": "/path/to/my-notes/README.md",
    "content": "# My Notes\n\nThis is my..."
  },
  {
    "file_path": "/path/to/my-notes/projects/alpha.md",
    "content": "# Project Alpha\n\n..."
  }
]
```

**Breaking it down**:
- `SELECT *` - Return all columns (file_path and content)
- `FROM read_markdown('**/*.md', include_filepath:=true)` - From all .md files
- `LIMIT 10` - Return maximum 10 results

---

## Progressive Learning Path

### Your Learning Journey

| Level | Topic | Time | Difficulty | Objective |
|-------|-------|------|-----------|-----------|
| **Level 1** | Basic Queries | 5 min | Beginner | List all notes, count files |
| **Level 2** | Content Search | 10 min | Beginner | Find notes by text, use LIKE |
| **Level 3** | Metadata Analysis | 15 min | Intermediate | Extract stats, analyze structure |
| **Level 4** | Complex Queries | 20 min | Advanced | Joins, aggregations, conditions |

**Progression strategy**:
1. Copy-paste Level 1 examples to verify they work
2. Modify examples to match your data
3. Combine patterns from multiple examples
4. Progress to next level only after examples work

**⚠️ Important**: Each level assumes you've completed the previous level.

---

## Level 1: Basic Queries

**Duration**: 5 minutes  
**Goal**: Understand SELECT, FROM, and basic filtering  
**Difficulty**: Beginner

### Query 1.1: List All Notes

```bash
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)"
```

**What it does**: Returns list of all markdown files in your notebook

**Modify it**: 
- Limit to specific folder: `'projects/**/*.md'` instead of `'**/*.md'`
- Include content: Add `content` to SELECT: `SELECT file_path, content FROM ...`

---

### Query 1.2: Count Total Notes

```bash
jot notes search --sql "SELECT COUNT(*) as total_notes FROM read_markdown('**/*.md')"
```

**What it does**: Returns single number—total count of notes

**Output**:
```json
[{"total_notes": 157}]
```

**Modify it**:
- Count specific folder: `read_markdown('projects/**/*.md')`
- Count nested folders: `read_markdown('daily/**/*.md')`

---

### Query 1.3: List Notes Limit and Offset

```bash
# First 10 notes
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 10"

# Next 10 notes (skip first 10)
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 10 OFFSET 10"
```

**What it does**: Returns paginated results (useful for large collections)

---

### Query 1.4: Sort Notes by Filename

```bash
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) ORDER BY file_path"
```

**What it does**: Lists all notes sorted alphabetically

**Modify it**:
- Reverse order: `ORDER BY file_path DESC`
- Sort by filename only: `ORDER BY file_path DESC` then parse in script

---

### Query 1.5: List Notes in Specific Folder

```bash
# All notes in projects folder
jot notes search --sql "SELECT file_path FROM read_markdown('projects/**/*.md', include_filepath:=true)"

# Only in projects root (not subfolders)
jot notes search --sql "SELECT file_path FROM read_markdown('projects/*.md', include_filepath:=true)"

# All daily notes from 2024
jot notes search --sql "SELECT file_path FROM read_markdown('daily/2024-*.md', include_filepath:=true)"
```

**What it does**: Filters notes by file pattern

---

## Level 2: Content Search

**Duration**: 10 minutes  
**Goal**: Find notes by text content using LIKE and ILIKE  
**Difficulty**: Beginner

### Query 2.1: Find Notes by Exact Text

```bash
jot notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%TODO%'"
```

**What it does**: Finds all notes containing exactly "TODO"

**Modify it**:
- Change search term: Replace `'%TODO%'` with your term
- Different text: `LIKE '%deadline%'`

**Important**: `LIKE` is case-sensitive. Use `ILIKE` for case-insensitive search.

---

### Query 2.2: Case-Insensitive Search

```bash
jot notes search --sql \
  "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true) WHERE content ILIKE '%python%' LIMIT 10"
```

**What it does**: Finds notes containing "python", "Python", or "PYTHON"

**Modify it**:
- Change term: Replace `'%python%'` with your search term
- Include more context: Add other columns: `SELECT file_path, content`
- Limit results: Change `LIMIT 10` to `LIMIT 20`, etc.

---

### Query 2.3: Find Notes with Multiple Keywords

```bash
jot notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content ILIKE '%deadline%' AND content ILIKE '%urgent%' LIMIT 20"
```

**What it does**: Finds notes containing BOTH "deadline" AND "urgent"

**Modify it**:
- Match either term: Replace `AND` with `OR`
- Add more conditions: Add more `AND content ILIKE '%term%'`

---

### Query 2.4: Find Notes Without Specific Text

```bash
jot notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content NOT LIKE '%archived%' LIMIT 10"
```

**What it does**: Finds notes that DON'T contain "archived"

**Use cases**: Filter out old notes, find incomplete items, etc.

---

### Query 2.5: Search in Specific Folders

```bash
# Find "bug" in projects only
jot notes search --sql \
  "SELECT file_path FROM read_markdown('projects/**/*.md', include_filepath:=true) WHERE content ILIKE '%bug%' LIMIT 20"

# Find "meeting" in daily notes from 2024
jot notes search --sql \
  "SELECT file_path FROM read_markdown('daily/2024-*.md', include_filepath:=true) WHERE content ILIKE '%meeting%' LIMIT 20"
```

**What it does**: Combines file pattern filtering with content search

**Combine with all Level 2 queries**: Replace `'**/*.md'` with specific pattern

---

### Query 2.6: Find Notes with Specific Format

```bash
# Notes with checked checkboxes [x]
jot notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%[x]%'"

# Notes with unchecked checkboxes [ ]
jot notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%[ ]%'"

# Notes with code blocks
jot notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%```%'"

# Notes with links
jot notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%[%](%'  LIMIT 20"
```

**What it does**: Search for structural elements in markdown

---

## Level 3: Metadata Analysis

**Duration**: 15 minutes  
**Goal**: Extract and analyze note statistics  
**Difficulty**: Intermediate

### Query 3.1: Get Word Count per Note

```bash
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count as words FROM read_markdown('**/*.md', include_filepath:=true) ORDER BY words DESC LIMIT 10"
```

**What it does**: Returns file path and word count, sorted by longest first

**Output**:
```json
[
  {"file_path": "projects/alpha-spec.md", "words": 2847},
  {"file_path": "projects/beta-overview.md", "words": 1523},
  {"file_path": "daily/2024-01-15.md", "words": 342}
]
```

---

### Query 3.2: Total Statistics About Collection

```bash
jot notes search --sql \
  "SELECT COUNT(*) as total_notes, SUM((md_stats(content)).word_count) as total_words, AVG((md_stats(content)).word_count) as avg_words FROM read_markdown('**/*.md')"
```

**What it does**: Collection-wide statistics (total files, total words, average)

**Output**:
```json
[
  {
    "total_notes": 157,
    "total_words": 142857,
    "avg_words": 910
  }
]
```

---

### Query 3.3: Find Notes by Length

```bash
# Long notes (over 2000 words)
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count as words FROM read_markdown('**/*.md', include_filepath:=true) WHERE (md_stats(content)).word_count > 2000 ORDER BY words DESC"

# Short notes (under 100 words)
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count as words FROM read_markdown('**/*.md', include_filepath:=true) WHERE (md_stats(content)).word_count < 100"

# Medium notes (500-1500 words)
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count as words FROM read_markdown('**/*.md', include_filepath:=true) WHERE (md_stats(content)).word_count BETWEEN 500 AND 1500"
```

**What it does**: Find notes by length criteria

---

### Query 3.4: Heading Count per Note

```bash
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).heading_count as headings FROM read_markdown('**/*.md', include_filepath:=true) ORDER BY headings DESC LIMIT 15"
```

**What it does**: Find most structured notes (by heading count)

---

### Query 3.5: Line Count Statistics

```bash
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).line_count as lines FROM read_markdown('**/*.md', include_filepath:=true) ORDER BY lines DESC LIMIT 10"
```

**What it does**: See longest notes by line count

---

### Query 3.6: Analyze Folder Statistics

```bash
# Stats for all notes in projects folder
jot notes search --sql \
  "SELECT COUNT(*) as total, SUM((md_stats(content)).word_count) as total_words FROM read_markdown('projects/**/*.md')"

# Compare folders
jot notes search --sql \
  "SELECT 'projects' as folder, COUNT(*) as count, SUM((md_stats(content)).word_count) as words FROM read_markdown('projects/**/*.md')
   UNION
   SELECT 'daily' as folder, COUNT(*) as count, SUM((md_stats(content)).word_count) as words FROM read_markdown('daily/**/*.md')"
```

**What it does**: Understand structure and distribution of your collection

---

## Level 4: Complex Queries

**Duration**: 20 minutes  
**Goal**: Combine patterns, use advanced SQL features  
**Difficulty**: Advanced

### Query 4.1: Find Incomplete Tasks

```bash
jot notes search --sql \
  "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%[ ]%' AND content NOT LIKE '%[x]%' LIMIT 20"
```

**What it does**: Finds notes with unchecked but no checked items (likely incomplete)

**Enhancement**: Filter to specific folder:
```bash
jot notes search --sql \
  "SELECT file_path FROM read_markdown('projects/**/*.md', include_filepath:=true) WHERE content LIKE '%[ ]%'"
```

---

### Query 4.2: Find Recently Modified Notes (by filename date pattern)

```bash
# Daily notes from last 7 days (if named YYYY-MM-DD)
jot notes search --sql \
  "SELECT file_path FROM read_markdown('daily/2024-01-*.md', include_filepath:=true) ORDER BY file_path DESC LIMIT 7"
```

**What it does**: Find recent notes based on naming pattern

**Adapt it**: Modify date pattern to match your naming convention

---

### Query 4.3: Combine Content and Metadata Filters

```bash
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count as words 
   FROM read_markdown('**/*.md', include_filepath:=true) 
   WHERE content ILIKE '%TODO%' AND (md_stats(content)).word_count > 100 
   ORDER BY words DESC"
```

**What it does**: Find long TODO items (complex enough to track)

**What's happening**:
- Find notes with "TODO" (content filter)
- AND have more than 100 words (metadata filter)
- Sort longest first

---

### Query 4.4: Pattern Matching with Complex Conditions

```bash
# Find code examples in documentation
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count as words
   FROM read_markdown('reference/**/*.md', include_filepath:=true)
   WHERE content LIKE '%```%' AND (md_stats(content)).word_count > 500
   ORDER BY words DESC"

# Find outdated documentation
jot notes search --sql \
  "SELECT file_path
   FROM read_markdown('docs/**/*.md', include_filepath:=true)
   WHERE (content LIKE '%FIXME%' OR content LIKE '%TODO%' OR content LIKE '%OUTDATED%')
   AND (md_stats(content)).word_count > 50"
```

**What it does**: Practical patterns for real workflows

---

### Query 4.5: Group and Aggregate Results

```bash
-- Count notes by first-level folder (projects, daily, archive)
SELECT 
  SUBSTRING(file_path, 1, POSITION('/' IN file_path) - 1) as folder,
  COUNT(*) as note_count,
  SUM((md_stats(content)).word_count) as total_words,
  AVG((md_stats(content)).word_count) as avg_words
FROM read_markdown('**/*.md', include_filepath:=true)
GROUP BY SUBSTRING(file_path, 1, POSITION('/' IN file_path) - 1)
ORDER BY note_count DESC
```

**What it does**: Analyzes structure of notebook by folder

**Advanced note**: Uses string functions to extract folder names

---

### Query 4.6: Export for Processing

```bash
# Get all content for external processing
jot notes search --sql \
  "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 100"

# Export just file paths for scripting
jot notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)" | jq -r '.[].file_path'

# Complex export
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count FROM read_markdown('**/*.md', include_filepath:=true) ORDER BY word_count DESC" \
  | jq 'map({path: .file_path, words: .word_count})'
```

**What it does**: Prepare data for piping to other tools

---

## DuckDB Markdown Functions

### md_stats() - Extract Statistics

Returns metadata about markdown content:

```bash
jot notes search --sql \
  "SELECT (md_stats(content)).word_count, (md_stats(content)).line_count FROM read_markdown('**/*.md') LIMIT 5"
```

**Available fields**:
- `.word_count` - Number of words
- `.line_count` - Number of lines
- `.heading_count` - Number of headings
- `.code_block_count` - Number of code blocks
- `.link_count` - Number of links

---

### read_markdown() - Read Files

Reads markdown files with optional parameters:

```bash
-- Basic: all .md files
read_markdown('**/*.md')

-- Include file paths
read_markdown('**/*.md', include_filepath:=true)

-- Specific pattern
read_markdown('projects/**/*.md')
```

**Patterns**:
- `*.md` - Notebook root only
- `**/*.md` - All files recursively
- `folder/*.md` - Specific folder
- `**/folder/*.md` - Folder at any level

---

## Performance Tips

### Write Efficient Queries

```sql
-- ❌ Inefficient: Calculates stats for all notes
SELECT * FROM read_markdown('**/*.md') WHERE (md_stats(content)).word_count > 1000

-- ✅ Efficient: Use LIMIT early
SELECT file_path FROM read_markdown('**/*.md') LIMIT 1000

-- ✅ Better: Filter by pattern first
SELECT file_path FROM read_markdown('projects/**/*.md') 
WHERE content ILIKE '%important%'
```

### Use Specific Patterns

```sql
-- ❌ Broad pattern
read_markdown('**/*.md')

-- ✅ Specific when possible
read_markdown('projects/**/*.md')
```

### LIMIT Your Results

```sql
-- Always use LIMIT for testing
SELECT * FROM read_markdown('**/*.md') LIMIT 10

-- Once verified, increase as needed
SELECT * FROM read_markdown('**/*.md') LIMIT 1000
```

---

## Common Mistakes

### Mistake 1: Case Sensitivity with LIKE

```sql
-- ❌ Case-sensitive (misses "TODO", "todo")
WHERE content LIKE '%TODO%'

-- ✅ Case-insensitive
WHERE content ILIKE '%todo%'
```

### Mistake 2: Missing Wildcards in LIKE

```sql
-- ❌ Won't match anything (exact match only)
WHERE content LIKE 'TODO'

-- ✅ Matches "TODO" anywhere in content
WHERE content LIKE '%TODO%'
```

### Mistake 3: Forgetting include_filepath

```sql
-- ❌ Returns content only (hard to know which note)
SELECT * FROM read_markdown('**/*.md')

-- ✅ Includes file paths for context
SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true)
```

### Mistake 4: Stats Function Syntax

```sql
-- ❌ Wrong syntax (missing parentheses)
WHERE md_stats(content).word_count > 1000

-- ✅ Correct syntax
WHERE (md_stats(content)).word_count > 1000
```

### Mistake 5: Not Using LIMIT for New Queries

```sql
-- ❌ Risk: Could return thousands of rows
SELECT * FROM read_markdown('**/*.md') WHERE content LIKE '%anything%'

-- ✅ Always LIMIT first
SELECT * FROM read_markdown('**/*.md') WHERE content LIKE '%anything%' LIMIT 10
```

---

## When to Use SQL

### Use SQL When You Need:

✅ **Search across all notes quickly**
```bash
Find all notes about "project alpha" in under 100ms
```

✅ **Statistical analysis**
```bash
Which notes are longest? Most structured? Most recently updated?
```

✅ **Complex filtering**
```bash
Find incomplete tasks that are also long (over 500 words)
```

✅ **Automation**
```bash
Export note list for script processing
```

✅ **Batch operations**
```bash
Get list of all notes matching criteria
```

### Use Regular Search When You Need:

✅ **Quick title keyword search (fieldless)**
```bash
jot notes search "project alpha"
```

✅ **Simple body/content lookup (explicit)**
```bash
jot notes search "body:deadline"
```

✅ **Browse recent notes**
```bash
jot notes list
```

---

## Practice Exercises

**Exercise 1: Basic Query**
```bash
# Try it: List first 5 notes in your notebook
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 5"
```

**Exercise 2: Content Search**
```bash
# Try it: Find all notes with "TODO" in them
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content ILIKE '%todo%'"
```

**Exercise 3: Metadata Analysis**
```bash
# Try it: Find your longest notes
jot notes search --sql "SELECT file_path, (md_stats(content)).word_count as words FROM read_markdown('**/*.md', include_filepath:=true) ORDER BY words DESC LIMIT 5"
```

**Exercise 4: Combine Everything**
```bash
# Try it: Find long TODO notes (complex tasks)
jot notes search --sql "SELECT file_path, (md_stats(content)).word_count FROM read_markdown('**/*.md', include_filepath:=true) WHERE content ILIKE '%todo%' AND (md_stats(content)).word_count > 500"
```

---

## Related Guides

- 📚 **[SQL Query Guide](sql-guide.md)** - Comprehensive SQL reference and advanced patterns
- 📋 **[SQL Functions Reference](sql-functions-reference.md)** - Complete function documentation
- 🚀 **[Getting Started for Power Users](getting-started-power-users.md)** - 15-minute onboarding with SQL
- 📖 **[Import Workflow Guide](import-workflow-guide.md)** - Set up your first notebook
- 🤖 **[Automation & JSON Integration](json-sql-guide.md)** - Pipe SQL results to other tools

---

## Next Steps

1. **Copy a Level 1 example** and run it against your notebook
2. **Modify it** to match your notes (change search terms, folder patterns)
3. **Progress to Level 2** once comfortable with results
4. **Build your own queries** by combining patterns
5. **Check [SQL Query Guide](sql-guide.md)** for advanced patterns

Happy querying! 🎯
