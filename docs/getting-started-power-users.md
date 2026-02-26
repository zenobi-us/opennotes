# Getting Started for Power Users: 15-Minute Onboarding

> **Looking for a simpler onboarding path?** Check out the [Getting Started: Beginner's Guide](getting-started-basics.md) if you prefer to learn basic note management without SQL first. You can always come back to this guide when you're ready for advanced features!

Welcome to Jot! This guide is designed for experienced developers who want to unlock the full power of Jot in 15 minutes.

## The Jot Advantage

Unlike basic note tools, Jot gives you:

- **SQL Querying**: Query your entire markdown collection using DuckDB's powerful SQL engine
- **Markdown Intelligence**: Extract structure, statistics, and metadata from markdown files
- **Automation Ready**: JSON output designed for piping to jq, shell scripts, and external tools
- **Developer-Friendly**: CLI-native, git-compatible, markdown-native, zero external dependencies

**Perfect for**: Developers managing large markdown collections, building personal knowledge bases, automating note processing, and creating data-driven workflows.

---

## Part 1: Import Your Existing Notes (2 minutes)

Jot works best with your existing markdown files. No migration needed—just point it at your notes directory.

### Basic Setup

```bash
# Create a notebook from your existing markdown folder
jot notebook create "My Notes" --path ~/my-notes

# Verify the import worked
jot notes list
```

### Verify Import Success

You should see your markdown files listed with titles extracted from frontmatter or filenames. For example:

```
### Notes (142)

- [Meeting Notes 2024-01] notes/meeting-notes-2024-01.md
- [Project Alpha Spec] notes/projects/alpha-spec.md
- [TODO List] notes/todo.md
```

**Pro Tip**: If you have multiple note collections (work, personal, projects), create separate notebooks:

```bash
jot notebook create "Work" --path ~/work/notes
jot notebook create "Personal" --path ~/personal/notes
jot notebook create "Projects" --path ~/projects/notes

# Switch between contexts automatically by changing directories
cd ~/work/notes && jot notes list  # Uses "Work" notebook
```

---

## Part 2: Discover SQL Power (5 minutes)

Now for the magic. Execute sophisticated queries against your entire note collection.

### Your First SQL Query

```bash
# Find all your markdown files
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md')"
```

**What this does**: Queries all markdown files in your notebook using DuckDB's SQL engine, returning clean JSON.

### Practical Examples

#### Find Notes by Content

```bash
# Search for "deadline" across all notes
jot notes search --sql \
  "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true)
   WHERE content ILIKE '%deadline%'
   LIMIT 10"
```

#### Get Statistics on Your Notes

```bash
# Find your longest notes (by word count)
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count as words
   FROM read_markdown('**/*.md', include_filepath:=true)
   ORDER BY words DESC
   LIMIT 10"
```

#### Find Actionable Items

```bash
# Find all unchecked tasks in your notes
jot notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)
   WHERE content LIKE '%[ ]%'
   ORDER BY file_path"
```

#### Analyze Your Note Patterns

```bash
# See word count distribution
jot notes search --sql \
  "SELECT
     CASE
       WHEN (md_stats(content)).word_count < 500 THEN 'short'
       WHEN (md_stats(content)).word_count < 2000 THEN 'medium'
       ELSE 'long'
     END as category,
     COUNT(*) as count
   FROM read_markdown('**/*.md')
   GROUP BY category
   ORDER BY count DESC"
```

### Learn More About SQL

All the examples above just scratch the surface. For complete documentation:

- **[SQL Query Guide](sql-guide.md)** - Complete patterns and best practices
- **[SQL Functions Reference](sql-functions-reference.md)** - Full function reference with examples
- **[DuckDB Markdown Extension](https://github.com/duckdb/duckdb-wasm/wiki/markdown-extension)** - Official docs

---

## Part 3: Automation with JSON (5 minutes)

All Jot query results are JSON—perfect for piping to tools and scripts.

### Basic JSON Output

```bash
# All SQL query results are automatically JSON
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md')"
# Output:
# [
#   { "file_path": "notes/project-ideas.md" },
#   { "file_path": "notes/meeting-notes.md" },
#   ...
# ]
```

### Using jq for Post-Processing

#### Extract Just Filenames

```bash
# Get just the file paths for piping to other commands
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md')" \
  | jq -r '.[].file_path'
```

#### Calculate Statistics

```bash
# Total word count across all notes
jot notes search --sql \
  "SELECT (md_stats(content)).word_count FROM read_markdown('**/*.md')" \
  | jq 'map(.word_count) | {
      total: add,
      count: length,
      average: (add / length | round)
    }'
```

#### Reformat and Export

```bash
# Export as CSV
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count FROM read_markdown('**/*.md')" \
  | jq -r '.[] | [.file_path, .word_count] | @csv'

# Export as tab-separated
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count FROM read_markdown('**/*.md')" \
  | jq -r '.[] | [.file_path, .word_count] | @tsv'
```

### Shell Script Integration

#### Find Large Files and Get Size

```bash
# Find markdown files and get real file size
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md')" \
  | jq -r '.[].file_path' \
  | xargs ls -lh

# Output:
# -rw-r--r-- 1 user group 42K Jan 20 12:34 notes/big-project.md
```

#### Count Lines in Notes

```bash
# Count total lines across all notes
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md')" \
  | jq -r '.[].file_path' \
  | xargs wc -l | tail -1
```

#### Find Files Modified Today

```bash
# Find notes created today and show their content
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md')" \
  | jq -r '.[].file_path' \
  | xargs find -mtime -1

# Combine with other tools
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md')" \
  | jq -r '.[].file_path' \
  | xargs ls -lh | awk '{print $9, $5}'
```

### Continuous Monitoring

Use cron + Jot for automated note processing:

```bash
#!/bin/bash
# Save as ~/bin/note-stats.sh

STATS=$(jot notes search --sql \
  "SELECT
     COUNT(*) as total_notes,
     AVG((md_stats(content)).word_count) as avg_words
   FROM read_markdown('**/*.md')" \
  | jq '.[] | "\(.total_notes) notes, \(.avg_words) avg words"')

echo "$(date): $STATS" >> ~/note-stats.log
```

Add to crontab:

```bash
# Run daily at 9am
0 9 * * * ~/bin/note-stats.sh
```

---

## Part 4: Your Workflow (3 minutes)

Now that you understand the core capabilities, here are some practical workflows.

### Workflow 1: Personal Knowledge Base

```bash
# Indexed knowledge base with search
jot notebook create "Knowledge" --path ~/knowledge

# Find related topics
jot notes search --sql \
  "SELECT DISTINCT file_path FROM read_markdown('**/*.md', include_filepath:=true)
   WHERE content ILIKE '%machine learning%' OR content ILIKE '%neural networks%'
   ORDER BY file_path"

# Generate index of all notes
jot notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count FROM read_markdown('**/*.md')
   ORDER BY file_path" \
  | jq -r '.[] | "- [\(.file_path)](\(.file_path)) (\(.word_count) words)"'
```

### Workflow 2: Project Documentation Management

```bash
# Create notebook for project docs
jot notebook create "ProjectDocs" --path ~/projects/docs

# Find all decision records
jot notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)
   WHERE file_path LIKE '%decision%' OR file_path LIKE '%adr%'
   ORDER BY file_path DESC LIMIT 20"

# Get documentation completeness
jot notes search --sql \
  "SELECT
     file_path,
     CASE WHEN (md_stats(content)).word_count > 500 THEN 'complete' ELSE 'needs-work' END
   FROM read_markdown('**/*.md')
   ORDER BY (md_stats(content)).word_count DESC"
```

### Workflow 3: Research and Reference

```bash
# Create research notebook
jot notebook create "Research" --path ~/research

# Find all references to specific topics
jot notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)
   WHERE content LIKE '%@TODO%' OR content LIKE '%[CITATION NEEDED]%'"

# Get topic frequency (markdown headings)
jot notes search --sql \
  "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true)
   LIMIT 100" | jq '.[] | select(.content | startswith("#"))'
```

### Workflow 4: Automation and Reporting

```bash
# Weekly stats email
WEEKLY_REPORT=$(jot notes search --sql \
  "SELECT
     COUNT(*) as new_notes,
     ROUND(AVG((md_stats(content)).word_count)) as avg_length
   FROM read_markdown('**/*.md')" \
  | jq '.[] | "Weekly: \(.new_notes) notes, \(.avg_length) avg words"')

echo "Subject: Weekly Notes Report" | \
  { cat; echo "$WEEKLY_REPORT"; } | \
  mail -s "Weekly Notes Report" you@example.com
```

---

## Next Steps: Advanced Topics

You now understand the Jot power-user workflow. Here's what's available for deeper dives:

### Reference Documentation

- **[SQL Functions Reference](sql-functions-reference.md)** — Complete list of available SQL functions and markdown-specific operations
- **[SQL Query Guide](sql-guide.md)** — Advanced patterns, performance tips, and security considerations
- **[JSON Output Guide](json-sql-guide.md)** — Comprehensive automation examples and tool integration patterns
- **[Notebook Discovery](notebook-discovery.md)** — Multi-notebook management and context-aware workflows

### Advanced Topics

1. **Performance Optimization** — Query large notebooks efficiently
2. **Security** — Understanding query validation and sandbox restrictions
3. **Custom Workflows** — Building shell scripts and automation
4. **Integration** — Connecting with other tools and systems

---

## Troubleshooting

### "Query returned no results"

```bash
# Verify your notebook is set up correctly
jot notes list

# Try a simple query first
jot notes search --sql "SELECT file_path FROM read_markdown('*.md') LIMIT 1"

# Check file patterns (use forward slashes, even on Windows)
# ✓ Good: '**/*.md', 'notes/*.md'
# ✗ Bad: '**\*.md', 'notes\.md'
```

### "File pattern not working"

File patterns are resolved from your notebook's root directory:

```bash
# If your notebook is at ~/my-notes/
# Pattern '**/*.md' means: ~/my-notes/**/*.md

# Create a test file to verify
mkdir -p ~/my-notes/test
echo "# Test" > ~/my-notes/test/sample.md

# Then try:
jot notes search --sql "SELECT file_path FROM read_markdown('test/*.md')"
```

### "JSON output format unexpected"

Jot always returns an array of objects:

```bash
# Single result
[{"file_path": "notes.md"}]

# Multiple results
[{"file_path": "notes1.md"}, {"file_path": "notes2.md"}]

# Parse with jq
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md')" | jq '.[].file_path'
```

### Performance Issues

For notebooks with 1000+ notes:

```bash
# Limit results
jot notes search --sql "... LIMIT 100"

# Filter early with WHERE clauses
jot notes search --sql "SELECT * FROM read_markdown('**/*.md') WHERE ..."

# See SQL optimization tips in sql-guide.md
```

---

## Key Takeaways

✅ **Jot = Markdown + SQL + Automation**

1. **Import** your existing markdown instantly—no migration needed
2. **Query** using SQL to find patterns and insights across your notes
3. **Automate** with JSON output for shell scripts and external tools
4. **Scale** efficiently from dozens to thousands of notes

✅ **Your New Superpowers**

- Find any note by content, pattern, or metadata
- Calculate statistics and generate reports
- Build automated workflows and monitoring
- Integrate with Unix tools, shell scripts, and external systems

✅ **What's Different**

Most note tools give you basic search. Jot gives you a full database query language over your markdown—a unique combination of simplicity and power.

---

## Related Learning Resources

### If You're New to Importing:

- **[Import Workflow Guide](import-workflow-guide.md)** - Comprehensive guide for importing existing markdown collections
  - Step-by-step import process for all scenarios
  - Migration from Obsidian, Bear, and generic markdown folders
  - Troubleshooting common import issues

### If You Want to Learn SQL Progressively:

- **[SQL Quick Reference](sql-quick-reference.md)** - Progressive learning path with 20+ practical examples
  - Level 1: Basic queries
  - Level 2: Content search
  - Level 3: Metadata analysis
  - Level 4: Complex queries
  - Practice exercises for each level

### For Complete Reference Documentation:

- **[SQL Query Guide](sql-guide.md)** - Detailed query documentation and patterns
- **[SQL Functions Reference](sql-functions-reference.md)** - Complete SQL function list
- **[JSON Output Guide](json-sql-guide.md)** - Automation and tool integration

### For Multi-Notebook Management:

- **[Notebook Discovery](notebook-discovery.md)** - Manage multiple notebooks and contexts

---

## Questions?

- **[Check the SQL Quick Reference](sql-quick-reference.md)** to start learning SQL progressively
- **[Read the SQL Guide](sql-guide.md)** for advanced query patterns
- **[See the Import Guide](import-workflow-guide.md)** if you're having import issues
- **[Join the Community](https://github.com/zenobi-us/jot)** on GitHub
