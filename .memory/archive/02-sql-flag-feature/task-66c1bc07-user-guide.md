# Task: Write SQL User Guide

**Epic**: [SQL Flag Feature](epic-2f3c4d5e-sql-flag-feature.md)

**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 3 - Documentation
**Priority**: MEDIUM
**Complexity**: Medium
**Estimated Time**: 90 minutes

## Objective

Create a comprehensive user guide document (`docs/sql-guide.md`) that teaches users how to write and execute SQL queries against their notebooks, with practical examples and troubleshooting.

## Context

Users need guidance on:
- Getting started with SQL flag
- Understanding available schema and functions
- Writing common query patterns
- Troubleshooting errors
- Performance tips
- Security model

## Steps to Take

1. **Create guide document**
   - File: `docs/sql-guide.md`
   - Table of contents with links
   - Clear section organization

2. **Write "Getting Started" section**
   - Basic syntax: `opennotes search --sql "SELECT ..."`
   - Running first query
   - Understanding output
   - Common errors and fixes

3. **Write "Available Functions" section**
   - Document DuckDB markdown extension functions
   - Table functions: `read_markdown()`, `read_markdown_blocks()`
   - Scalar functions: `md_extract_*`, `md_stats()`, etc.
   - Include parameter reference
   - Return type documentation

4. **Write "Schema Overview" section**
   - Available columns in `read_markdown()`
   - Available columns in `read_markdown_blocks()`
   - Data types for each column
   - Example queries for each table

5. **Write "Common Patterns" section**
   - Extract code blocks by language
   - Find notes by word count
   - Search by metadata/frontmatter
   - Extract links and images
   - Multi-table joins
   - Aggregations and statistics

6. **Write "Real-World Examples" section**
   - 5-10 practical examples:
     - Find all Python code snippets
     - Identify incomplete notes (< 200 words)
     - Track project status via frontmatter
     - Find external links
     - Calculate total word count
     - Find notes with specific tags

7. **Write "Troubleshooting" section**
   - Common error messages
   - How to debug queries
   - Performance tips
   - Timeout handling

8. **Write "Security Model" section**
   - Read-only access guarantee
   - Query validation explanation
   - Safe to run untrusted queries (mostly)
   - Limitations

9. **Add appendix with reference**
   - Complete function listing
   - Command line options
   - Links to resources

## Expected Outcomes

- [ ] Complete user guide written
- [ ] Clear examples for all features
- [ ] Troubleshooting section helpful
- [ ] Security model explained
- [ ] Guide is discoverable
- [ ] Documentation is accurate

## Acceptance Criteria

- [x] Guide covers all major features
- [x] At least 5 real-world examples
- [x] All DuckDB functions documented
- [x] Common patterns shown
- [x] Troubleshooting section helpful
- [x] Security model explained
- [x] Code examples are correct and tested
- [x] Links work (internal and external)
- [x] Markdown is well-formatted
- [x] Document is readable and complete

## Implementation Notes

### Document Structure
```markdown
# SQL Guide: Writing Queries Against Your Notes

## Table of Contents
- Getting Started
- Schema Overview
- Available Functions
- Common Patterns
- Real-World Examples
- Troubleshooting
- Security Model
- Appendix

## Getting Started

### Basic Syntax
- Command format
- Query structure
- Running queries

### Your First Query
- Simple SELECT
- Understanding output
- Common first-time issues

## Schema Overview

### read_markdown() Table
| Column | Type | Description |
| filepath | STRING | Path to markdown file |
| content | STRING | Full markdown content |

[more columns...]

### read_markdown_blocks() Table
[similar structure]

## Available Functions

### md_extract_code_blocks(content)
Returns: LIST
Description: Extract code blocks from markdown
Example: ...

[more functions...]

## Common Patterns

### Extract Code Blocks by Language
```sql
SELECT filepath, cb.language, cb.code
FROM read_markdown('**/*.md', include_filepath:=true),
LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
WHERE cb.language = 'python'
```

[more patterns...]

## Real-World Examples

### Example 1: Find All Python Snippets
[explanation + query + expected output]

### Example 2: Identify Short Notes
[explanation + query + expected output]

[more examples...]

## Troubleshooting

### Error: "only SELECT queries are allowed"
**Problem**: Tried to run UPDATE/DELETE query
**Solution**: Use SELECT instead

[more troubleshooting...]

## Security Model

### Read-Only Access
- Guaranteed by DuckDB access mode
- No data can be modified
- Safe to run untrusted queries (with caution)

### Query Validation
- SELECT and WITH only
- Dangerous keywords blocked
- 30-second timeout

## Appendix

### Complete Function Reference
[Full listing...]

### Command Options
- --sql
- --format (optional)

### Resources
- DuckDB Docs
- Markdown Extension Docs
```

### Example Queries to Include
1. Extract Python code blocks
2. Find short notes (< 300 words)
3. Extract frontmatter metadata
4. Find notes by tag
5. Count word count by note
6. Find broken links
7. Extract all images
8. Find notes modified recently
9. Statistics by folder
10. Complex multi-table query

## Dependencies

- ✅ [task-ed37261d-function-docs.md](task-ed37261d-function-docs.md) - Function reference (will reference)
- ✅ Specification document
- ✅ Research document on DuckDB

## Blockers

- None identified (can be written from spec and research)

## Time Estimate

- Structure and outline: 10 minutes
- Getting Started section: 15 minutes
- Schema Overview: 15 minutes
- Available Functions: 20 minutes
- Common Patterns: 15 minutes
- Real-World Examples: 20 minutes
- Troubleshooting: 10 minutes
- Security section: 5 minutes
- Review and polish: 10 minutes
- Total: 90 minutes

## Definition of Done

- [ ] Document created at docs/sql-guide.md
- [ ] All sections completed
- [ ] Examples are accurate and tested
- [ ] Links verified
- [ ] Markdown properly formatted
- [ ] Ready for publication
- [ ] Linked from README

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Priority**: Documentation/Story 3
**Links**:
- [Function Docs](task-ed37261d-function-docs.md)
- [SQL Spec](spec-a1b2c3d4-sql-flag.md)
