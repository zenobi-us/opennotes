# Advanced Automation Recipes

Master automation patterns to transform Jot into a powerful knowledge management and reporting system. This guide provides production-ready scripts and patterns for real-world workflows.

---

## Overview: Why Automation?

Jot excels when integrated into your existing workflows. Combine the CLI with shell scripts, cron jobs, and external tools to:

- **Automate insights generation** - Daily/weekly statistics without manual queries
- **Build dashboards** - Generate markdown reports for teams or personal tracking
- **Integrate with systems** - Feed note data into git, cloud storage, or analysis tools
- **Reduce friction** - One-command operations instead of multiple steps
- **Scale knowledge work** - Process thousands of notes with consistent patterns

---

## Personal Knowledge Base Automation

Use Jot to build automated reporting on your personal notes collection.

### Daily Note Statistics

Collect statistics about your note-taking patterns daily:

```bash
#!/bin/bash
# daily-note-stats.sh - Generate daily statistics report
# Usage: ./daily-note-stats.sh [notebook_name]
# Default: Uses configured notebook

NOTEBOOK="${1:- }"
DATE=$(date +"%Y-%m-%d")

echo "# Note Statistics - $DATE"
echo ""

# Total notes in collection
TOTAL=$(jot notes search --sql "SELECT COUNT(*) as count FROM read_markdown() WHERE file_path IS NOT NULL" 2>/dev/null | tail -1)
echo "**Total Notes**: $TOTAL"
echo ""

# Notes by folder
echo "## Notes by Folder"
echo ""
jot notes search --sql "
SELECT
  SUBSTRING(file_path FROM 1 FOR GREATEST(POSITION('/' IN REVERSE(file_path)) - 1, 1)) as folder,
  COUNT(*) as count
FROM read_markdown()
WHERE file_path IS NOT NULL
GROUP BY folder
ORDER BY count DESC
LIMIT 10
" 2>/dev/null | grep -v "^+" | tail -n +2
echo ""

# Average metrics
echo "## Writing Metrics"
echo ""
jot notes search --sql "
SELECT
  ROUND(AVG(md_stats(content).words), 0) as avg_words,
  ROUND(AVG(md_stats(content).lines), 0) as avg_lines,
  MAX(md_stats(content).words) as max_words
FROM read_markdown()
WHERE content IS NOT NULL
" 2>/dev/null | grep -v "^+" | tail -n +2
```

**Usage**:

```bash
chmod +x daily-note-stats.sh
./daily-note-stats.sh >> stats-log.md

# Add to cron for daily 9am reports:
# 0 9 * * * /path/to/daily-note-stats.sh >> /home/user/notes/stats-log.md
```

### Weekly Summary Generation

Generate automated weekly digests:

```bash
#!/bin/bash
# weekly-note-summary.sh - Generate weekly summary of note activity
# Usage: ./weekly-note-summary.sh [notebook_name] [output_file]

NOTEBOOK="${1:- }"
OUTPUT="${2:-weekly-summary.md}"
WEEK=$(date +"%Y-W%V")

cat > "$OUTPUT" << 'EOF'
# Weekly Note Summary

EOF

# Notes created this week (requires date tracking in metadata)
echo "## New Notes This Week" >> "$OUTPUT"
jot notes search --sql "
SELECT
  title,
  file_path,
  created_date
FROM read_markdown()
WHERE file_path IS NOT NULL
ORDER BY file_path DESC
LIMIT 5
" 2>/dev/null | grep -v "^+" | tail -n +2 >> "$OUTPUT"

# Most referenced topics (word frequency analysis)
echo "" >> "$OUTPUT"
echo "## Top Topics" >> "$OUTPUT"
jot notes search --sql "
SELECT
  SUBSTRING(content FROM 1 FOR 50) as snippet,
  COUNT(*) as frequency
FROM read_markdown()
WHERE content LIKE '%#%'
GROUP BY SUBSTRING(content FROM 1 FOR 50)
ORDER BY frequency DESC
LIMIT 10
" 2>/dev/null | grep -v "^+" | tail -n +2 >> "$OUTPUT"

echo "✅ Weekly summary generated: $OUTPUT"
```

**Usage**:

```bash
chmod +x weekly-note-summary.sh
./weekly-note-summary.sh > weekly-report.md

# Or add to cron for Mondays at 8am:
# 0 8 * * 1 /path/to/weekly-note-summary.sh > /home/user/notes/weekly-$(date +\%Y-\%V).md
```

### Automated Backup with Git

Create versioned backups automatically:

```bash
#!/bin/bash
# note-backup.sh - Backup notes with git, optionally to remote
# Usage: ./note-backup.sh [notebook_path] [remote_url]
# Example: ./note-backup.sh ~/my-notes origin

NOTEBOOK_PATH="${1:-$HOME/notes}"
REMOTE="${2:-}"

if [ ! -d "$NOTEBOOK_PATH" ]; then
  echo "❌ Error: Notebook path not found: $NOTEBOOK_PATH"
  exit 1
fi

cd "$NOTEBOOK_PATH"

# Initialize git if needed
if [ ! -d ".git" ]; then
  git init
  git config user.email "notes-backup@localhost"
  git config user.name "Notes Backup"
  echo "*.swp" > .gitignore
  echo ".DS_Store" >> .gitignore
fi

# Stage and commit
git add -A
CHANGES=$(git diff --cached --quiet || echo "yes")

if [ ! -z "$CHANGES" ]; then
  TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
  COUNT=$(git diff --cached --numstat | wc -l)
  git commit -m "backup: $COUNT files changed at $TIMESTAMP"
  echo "✅ Committed $COUNT file changes"
else
  echo "✅ No changes to commit"
fi

# Push to remote if configured
if [ ! -z "$REMOTE" ]; then
  git push "$REMOTE" main 2>/dev/null || git push "$REMOTE" master
  echo "✅ Pushed to remote: $REMOTE"
fi
```

**Usage**:

```bash
chmod +x note-backup.sh

# Manual backup
./note-backup.sh ~/my-notes

# Automatic daily backup (cron)
# 0 23 * * * /path/to/note-backup.sh /home/user/notes origin

# Or use git hooks for on-every-save backup:
# Create .git/hooks/post-merge: /path/to/note-backup.sh
```

---

## Project Documentation Workflows

Automate documentation tasks for teams or large projects.

### Auto-Generate Documentation Index

Build table of contents from your documentation:

```bash
#!/bin/bash
# doc-index-gen.sh - Generate documentation index with search
# Usage: ./doc-index-gen.sh [notebook_path] [output_file]

NOTEBOOK="${1:-.}"
OUTPUT="${2:-INDEX.md}"

cat > "$OUTPUT" << 'EOF'
# Documentation Index

Auto-generated documentation table of contents.

EOF

# By folder
echo "## By Folder" >> "$OUTPUT"
echo "" >> "$OUTPUT"
jot notes search --sql "
SELECT
  SUBSTRING(file_path FROM 1 FOR POSITION('/' IN file_path) - 1) as folder,
  COUNT(*) as doc_count
FROM read_markdown()
WHERE file_path LIKE '%.md'
GROUP BY folder
ORDER BY folder
" 2>/dev/null | grep -v "^+" | tail -n +2 >> "$OUTPUT"

# By type (detected from frontmatter or filename patterns)
echo "" >> "$OUTPUT"
echo "## By Type" >> "$OUTPUT"
echo "" >> "$OUTPUT"
echo "- [Architecture](docs/architecture/)" >> "$OUTPUT"
echo "- [API Reference](docs/api/)" >> "$OUTPUT"
echo "- [Guides](docs/guides/)" >> "$OUTPUT"
echo "- [FAQ](docs/faq/)" >> "$OUTPUT"

# Search index
echo "" >> "$OUTPUT"
echo "## Quick Search" >> "$OUTPUT"
echo "" >> "$OUTPUT"
echo "Use \`jot notes search\` to find documentation:" >> "$OUTPUT"
echo "" >> "$OUTPUT"
echo "\`\`\`bash" >> "$OUTPUT"
echo "# Find all architecture decisions" >> "$OUTPUT"
echo "jot notes search -i architecture" >> "$OUTPUT"
echo "" >> "$OUTPUT"
echo "# Find SQL examples" >> "$OUTPUT"
echo "jot notes search -i 'select\\|from'" >> "$OUTPUT"
echo "\`\`\`" >> "$OUTPUT"

echo "✅ Documentation index generated: $OUTPUT"
```

**Usage**:

```bash
chmod +x doc-index-gen.sh
./doc-index-gen.sh ~/my-project/docs > docs/INDEX.md

# Regenerate before commits:
# Add to git pre-commit hook
```

### Track Documentation Completeness

Find missing or incomplete documentation:

```bash
#!/bin/bash
# doc-completeness.sh - Report on documentation coverage
# Usage: ./doc-completeness.sh [notebook_path]

NOTEBOOK="${1:-.}"

echo "# Documentation Completeness Report"
echo ""
echo "Generated: $(date)"
echo ""

# Find TODO markers
echo "## Items Needing Work"
echo ""
jot notes search --sql "
SELECT file_path, COUNT(*) as todo_count
FROM read_markdown()
WHERE content LIKE '%TODO%'
   OR content LIKE '%FIXME%'
   OR content LIKE '%XXX%'
   OR content LIKE '%HACK%'
GROUP BY file_path
ORDER BY todo_count DESC
" 2>/dev/null | grep -v "^+" | tail -n +2 || echo "✅ No TODOs found!"

# Find empty sections
echo "" >>
echo "## Sections Without Content"
echo ""
jot notes search --sql "
SELECT file_path
FROM read_markdown()
WHERE content LIKE '%^## %'
   AND NOT content LIKE '%## .*[a-zA-Z0-9]%'
ORDER BY file_path
" 2>/dev/null | grep -v "^+" | tail -n +2 || echo "✅ All sections have content!"

echo ""
echo "---"
echo "Last checked: $(date '+%Y-%m-%d %H:%M')"
```

**Usage**:

```bash
chmod +x doc-completeness.sh
./doc-completeness.sh ~/my-project/docs
```

---

## Research & Analysis Automation

Transform raw research into organized knowledge.

### Collect and Deduplicate Sources

Manage research sources with automatic deduplication:

```bash
#!/bin/bash
# research-deduplicate.sh - Find duplicate sources in research
# Usage: ./research-deduplicate.sh [notebook_path]

NOTEBOOK="${1:-.}"

echo "# Research Deduplication Report"
echo ""

# Find duplicate URLs
echo "## Duplicate URLs"
echo ""
jot notes search --sql "
SELECT
  url,
  COUNT(*) as occurrences,
  COUNT(DISTINCT file_path) as in_files
FROM (
  SELECT
    SUBSTRING(content FROM POSITION('http' IN content) FOR 100) as url,
    file_path
  FROM read_markdown()
  WHERE content LIKE '%http%'
)
GROUP BY url
HAVING COUNT(*) > 1
ORDER BY occurrences DESC
LIMIT 20
" 2>/dev/null | grep -v "^+" | tail -n +2

# Find duplicate citations (by author or title)
echo ""
echo "## Possible Duplicate Citations"
echo ""
jot notes search --sql "
SELECT
  SUBSTRING(content FROM 1 FOR 80) as citation_snippet,
  COUNT(*) as occurrences
FROM read_markdown()
WHERE content LIKE '%[^]%:%'
GROUP BY SUBSTRING(content FROM 1 FOR 80)
HAVING COUNT(*) > 1
ORDER BY occurrences DESC
LIMIT 10
" 2>/dev/null | grep -v "^+" | tail -n +2

echo ""
echo "✅ Report generated"
```

**Usage**:

```bash
chmod +x research-deduplicate.sh
./research-deduplicate.sh ~/research/papers
```

---

## Shell Script Integration

Ready-to-use scripts for common tasks. All scripts include error handling and helpful output.

### note-stats.sh - Generate Weekly Statistics

Complete statistics collection script:

```bash
#!/bin/bash
# note-stats.sh - Comprehensive note statistics
# Usage: ./note-stats.sh [notebook] [--json] [--save]

NOTEBOOK="${1:- }"
JSON_FORMAT="${2:---json}"
SAVE_RESULTS="${3:---save}"

# Generate report
REPORT=$(mktemp)

{
  echo "# Note Statistics Report"
  echo "Generated: $(date '+%Y-%m-%d %H:%M:%S')"
  echo ""

  echo "## Summary"
  jot notes search --sql "
  SELECT
    COUNT(*) as total_notes,
    SUM(md_stats(content).lines) as total_lines,
    SUM(md_stats(content).words) as total_words,
    ROUND(AVG(md_stats(content).words), 0) as avg_words_per_note
  FROM read_markdown()
  WHERE content IS NOT NULL
  " 2>/dev/null | tail -1

} > "$REPORT"

if [ "$SAVE_RESULTS" = "--save" ]; then
  cp "$REPORT" "stats-$(date +%Y-%m-%d).md"
  echo "✅ Saved: stats-$(date +%Y-%m-%d).md"
fi

cat "$REPORT"
rm "$REPORT"
```

### note-search.sh - Enhanced Search with Formatting

Better search results formatting:

```bash
#!/bin/bash
# note-search.sh - Search with formatted output
# Usage: ./note-search.sh [pattern] [--all] [--count]

PATTERN="${1:- }"
SHOW_ALL="${2:---all}"
SHOW_COUNT="${3:---count}"

if [ -z "$PATTERN" ]; then
  echo "Usage: note-search.sh [pattern] [--all] [--count]"
  echo ""
  echo "Examples:"
  echo "  note-search.sh todo             # Find TODOs"
  echo "  note-search.sh 'sql|query' --all # Find SQL references"
  exit 1
fi

echo "🔍 Searching for: $PATTERN"
echo ""

RESULTS=$(jot notes search "$PATTERN" 2>/dev/null)
COUNT=$(echo "$RESULTS" | grep -c "^-")

if [ "$SHOW_COUNT" = "--count" ]; then
  echo "Found $COUNT results"
else
  echo "$RESULTS"
fi
```

### note-export.sh - Export Notes to Various Formats

Convert notes to different formats:

```bash
#!/bin/bash
# note-export.sh - Export notes to various formats
# Usage: ./note-export.sh [format] [notebook] [output_dir]
# Formats: json, csv, html

FORMAT="${1:-json}"
NOTEBOOK="${2:- }"
OUTPUT_DIR="${3:-./export}"

mkdir -p "$OUTPUT_DIR"

case "$FORMAT" in
  json)
    echo "📤 Exporting to JSON..."
    jot notes search --sql "
    SELECT
      file_path,
      md_stats(content).words as words,
      md_stats(content).lines as lines,
      SUBSTRING(content FROM 1 FOR 200) as preview
    FROM read_markdown()
    " 2>/dev/null | tail -1 > "$OUTPUT_DIR/export.json"
    echo "✅ Exported to: $OUTPUT_DIR/export.json"
    ;;
  csv)
    echo "📤 Exporting to CSV..."
    jot notes search --sql "
    SELECT file_path, words, lines
    FROM (
      SELECT
        file_path,
        md_stats(content).words as words,
        md_stats(content).lines as lines
      FROM read_markdown()
    )
    " 2>/dev/null | tail -1 > "$OUTPUT_DIR/export.csv"
    echo "✅ Exported to: $OUTPUT_DIR/export.csv"
    ;;
  *)
    echo "❌ Unknown format: $FORMAT"
    echo "Supported: json, csv, html"
    exit 1
    ;;
esac
```

### note-report.sh - Generate Markdown Reports

Create markdown reports from queries:

```bash
#!/bin/bash
# note-report.sh - Generate markdown reports
# Usage: ./note-report.sh [report_type] [output_file]
# Types: daily, weekly, monthly

REPORT_TYPE="${1:-daily}"
OUTPUT="${2:-report-$(date +%Y-%m-%d).md}"

generate_daily_report() {
  cat << 'EOF'
# Daily Report

## Overview
EOF

  jot notes search --sql "
  SELECT 'Total Notes: ' || COUNT(*) as stat
  FROM read_markdown()
  " 2>/dev/null | tail -1 >> "$OUTPUT"
}

generate_weekly_report() {
  cat << 'EOF'
# Weekly Report

## Overview
EOF

  jot notes search --sql "
  SELECT 'Notes Modified: ' || COUNT(*) as stat
  FROM read_markdown()
  WHERE file_path LIKE '%.md'
  " 2>/dev/null | tail -1 >> "$OUTPUT"
}

generate_monthly_report() {
  cat << 'EOF'
# Monthly Report

## Statistics
EOF

  jot notes search --sql "
  SELECT 'Total Notes: ' || COUNT(*) as stat
  FROM read_markdown()
  " 2>/dev/null | tail -1 >> "$OUTPUT"
}

case "$REPORT_TYPE" in
  daily) generate_daily_report ;;
  weekly) generate_weekly_report ;;
  monthly) generate_monthly_report ;;
  *) echo "Unknown report type: $REPORT_TYPE" && exit 1 ;;
esac

echo "✅ Report generated: $OUTPUT"
```

---

## Cron Integration

Automate recurring tasks with cron jobs.

### Cron Setup Examples

Add these to your crontab with `crontab -e`:

```bash
# Daily statistics at 9 AM
0 9 * * * cd /home/user/notes && ./daily-note-stats.sh >> stats-log.md

# Weekly summary every Monday at 8 AM
0 8 * * 1 /home/user/scripts/weekly-note-summary.sh > /home/user/notes/weekly-$(date +\%Y-\%V).md

# Automatic backup every day at 11 PM
0 23 * * * /home/user/scripts/note-backup.sh /home/user/notes origin

# Documentation index regeneration every Friday at 5 PM
0 17 * * 5 cd /home/user/projects/docs && /home/user/scripts/doc-index-gen.sh . INDEX.md

# Completeness check daily at 8 AM
0 8 * * * /home/user/scripts/doc-completeness.sh /home/user/projects/docs > /tmp/doc-status.txt 2>&1
```

### Important: Environment Variables in Cron

When using cron, you may need to set environment variables:

```bash
# Add to beginning of crontab
SHELL=/bin/bash
PATH=/usr/local/bin:/usr/bin:/bin
HOME=/home/user

# Now your jobs will have access to jot in PATH
0 9 * * * cd /home/user/notes && ./daily-note-stats.sh >> stats-log.md
```

---

## Tool Integration Examples

### jq Pipelines for Data Transformation

Transform Jot output with jq:

```bash
# Export SQL results to JSON and process with jq
jot notes search --sql "
SELECT file_path, md_stats(content).words as words
FROM read_markdown()
ORDER BY words DESC
LIMIT 10
" | jq -r '.[] | "\(.file_path): \(.words) words"'

# Count notes by folder depth
jot notes search --sql "
SELECT file_path FROM read_markdown()
" | jq -r '.[] | .file_path | split("/") | length' | sort | uniq -c

# Find notes with specific metadata
jot notes search --sql "
SELECT file_path, content FROM read_markdown()
WHERE content LIKE '%author:%'
" | jq -r '.[] | select(.content | contains("author:")) | .file_path'
```

### Git Workflow Integration

Combine Jot with git operations:

```bash
#!/bin/bash
# sync-and-report.sh - Sync notes and generate report
# Usage: ./sync-and-report.sh

cd ~/my-notes

# Pull latest changes
git pull origin main

# Generate statistics
./daily-note-stats.sh >> stats.md

# Commit changes
git add -A
git commit -m "docs: daily statistics update"

# Push back
git push origin main

echo "✅ Notes synced and stats generated"
```

### Obsidian Compatibility

Use Jot to analyze Obsidian vaults:

```bash
#!/bin/bash
# obsidian-analyzer.sh - Analyze Obsidian vault with Jot
# Usage: ./obsidian-analyzer.sh [vault_path]

VAULT="${1:-.}"

echo "# Obsidian Vault Analysis"
echo ""

# Count notes with specific tags
echo "## Notes by Tag"
jot notes search --sql "
SELECT
  SUBSTRING(content FROM POSITION('#' IN content) FOR 20) as tag,
  COUNT(*) as count
FROM read_markdown()
WHERE content LIKE '%#%'
GROUP BY tag
ORDER BY count DESC
LIMIT 20
"

# Find broken links
echo ""
echo "## Potential Broken Links"
jot notes search --sql "
SELECT file_path,
  SUBSTRING(content FROM POSITION('[[' IN content) FOR 50) as potential_link
FROM read_markdown()
WHERE content LIKE '%[[%'
LIMIT 10
"
```

### Bear Notes Bridge

Migrate from Bear and keep analysis consistent:

```bash
#!/bin/bash
# bear-to-jot.sh - Process Bear export
# Usage: ./bear-to-jot.sh [bear_export_dir] [notebook_name]

BEAR_DIR="${1:-.}"
NOTEBOOK="${2:-Bear-Notes}"

# Bear exports as individual files - create notebook
jot notebook create "$NOTEBOOK" --path "$BEAR_DIR"

# Analyze converted structure
echo "✅ Imported from Bear"
jot notes list

# Export statistics
jot notes search --sql "
SELECT
  COUNT(*) as total,
  SUM(md_stats(content).words) as total_words,
  AVG(md_stats(content).words) as avg_words
FROM read_markdown()
"
```

---

## Performance Considerations

When automating queries across large note collections:

### Optimize Query Performance

```bash
# ❌ Slow: Processes entire content
jot notes search --sql "SELECT * FROM read_markdown()"

# ✅ Fast: Uses metadata only when possible
jot notes search --sql "
SELECT file_path, md_stats(content).words
FROM read_markdown()
WHERE file_path LIKE '%.md'
LIMIT 100
"

# ❌ Slow: Full content scan for pattern
jot notes search --sql "
SELECT * FROM read_markdown()
WHERE content LIKE '%pattern%'
"

# ✅ Better: Combine with file pattern
jot notes search --sql "
SELECT file_path FROM read_markdown()
WHERE file_path LIKE '%docs/%'
AND content LIKE '%pattern%'
LIMIT 50
"
```

### Memory and Timeout Management

```bash
# For very large collections, limit results:
jot notes search --sql "
SELECT file_path FROM read_markdown()
LIMIT 1000
"

# Use pagination pattern in scripts:
LIMIT=100
OFFSET=0
while [ $OFFSET -lt 10000 ]; do
  jot notes search --sql "
  SELECT file_path FROM read_markdown()
  LIMIT $LIMIT OFFSET $OFFSET
  "
  OFFSET=$((OFFSET + LIMIT))
done
```

### Cron Resource Limits

Prevent cron jobs from overwhelming system:

```bash
# In crontab: Use nice/ionice for background priority
0 23 * * * nice -n 19 ionice -c3 /home/user/scripts/heavy-backup.sh

# Redirect logs to prevent filling disk
0 */6 * * * /home/user/scripts/report.sh >> /tmp/report.log 2>&1

# Remove old reports to save space
0 0 * * 0 find /home/user/reports -name "*.md" -mtime +30 -delete
```

---

## Troubleshooting Automation

### Script Not Finding jot

Make sure `jot` is in your PATH:

```bash
# Check if jot is available
which jot

# If not found, add to your script:
export PATH="/home/user/.local/bin:$PATH"

# Or use full path:
/home/user/.local/bin/jot notes search ...
```

### Cron Jobs Not Running

Debug cron issues:

```bash
# Check cron log (Linux)
grep CRON /var/log/syslog | tail -20

# Check if cron is running
ps aux | grep cron

# Test your script manually with cron environment
env -i /bin/sh -c 'cd /tmp && /path/to/your/script.sh'

# Make scripts executable
chmod +x /home/user/scripts/*.sh
```

### Permission Denied Errors

Ensure proper permissions:

```bash
# Scripts must be executable
chmod +x *.sh

# Notebooks must be readable
chmod -R 755 ~/my-notes

# For cron jobs, use same user who created notebook:
# Don't run from root crontab if user created notebook
```

---

## Next Steps

Once you've mastered these recipes:

1. **Extend for your workflows** - Adapt scripts to your specific needs
2. **Create custom reports** - Build domain-specific analysis queries
3. **Integrate with services** - Send reports to Slack, email, or cloud storage
4. **Build dashboards** - Combine multiple reports into living documentation

See [SQL Quick Reference](sql-quick-reference.md) for query examples and [JSON-SQL Guide](json-sql-guide.md) for advanced data transformation patterns.
