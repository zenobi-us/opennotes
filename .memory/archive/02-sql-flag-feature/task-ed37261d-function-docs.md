# Task: Document Available SQL Functions

**Epic**: [SQL Flag Feature](epic-2f3c4d5e-sql-flag-feature.md)

**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 3 - Documentation
**Priority**: MEDIUM
**Complexity**: Low
**Estimated Time**: 45 minutes

## Objective

Create a reference document (`docs/sql-functions-reference.md`) that comprehensively documents all DuckDB markdown extension functions available to users, with parameter descriptions and usage examples.

## Context

Users need a quick reference to understand what functions are available, their parameters, return types, and basic usage. This is different from the user guide (which teaches concepts) - this is a reference for looking things up.

## Steps to Take

1. **Create reference document**
   - File: `docs/sql-functions-reference.md`
   - Organization: by function type (table, scalar)
   - Quick lookup structure

2. **Document table functions**
   - `read_markdown()`
   - `read_markdown_blocks()`
   - `read_markdown_sections()`
   - For each: parameters, return columns, example

3. **Document scalar functions**
   - `md_extract_code_blocks()`
   - `md_extract_links()`
   - `md_extract_images()`
   - `md_extract_metadata()`
   - `md_stats()`
   - `md_to_text()`
   - `md_to_html()`
   - For each: parameters, return type, description, example

4. **Document utility functions**
   - Standard DuckDB functions that work with markdown
   - UNNEST for expanding arrays
   - List operations
   - String functions

5. **Add parameter reference**
   - Common parameters across functions
   - `include_filepath` parameter
   - Path patterns (glob syntax)
   - Data types

6. **Add type reference**
   - Return types: STRING, LIST, MAP, STRUCT
   - Column types in tables
   - How to work with complex types

7. **Add examples for each function**
   - Minimal working example
   - Show expected output
   - Link to user guide for more context

## Expected Outcomes

- [ ] All markdown functions documented
- [ ] Parameters clearly described
- [ ] Return types explained
- [ ] Examples provided
- [ ] Quick reference usable
- [ ] Linked from main docs

## Acceptance Criteria

- [x] All table functions documented
- [x] All scalar functions documented
- [x] Parameters described clearly
- [x] Return types specified
- [x] At least one example per function
- [x] Examples are correct
- [x] Document is well-organized
- [x] Easy to search/navigate
- [x] Linked from other docs

## Implementation Notes

### Reference Format
```markdown
# SQL Functions Reference

## Table Functions

### read_markdown(pattern, include_filepath := false)

**Description**: Read and parse markdown files matching glob pattern

**Parameters**:
- `pattern` (VARCHAR): Glob pattern for file matching (e.g., '**/*.md')
- `include_filepath` (BOOLEAN, optional): Add filepath column to results

**Returns**: TABLE with columns:
- `filepath` (VARCHAR, if include_filepath=true)
- `content` (VARCHAR)
- `line_number` (INTEGER)
- ... other columns

**Example**:
```sql
SELECT filepath, LENGTH(content) as bytes
FROM read_markdown('**/*.md', include_filepath:=true)
ORDER BY bytes DESC LIMIT 10
```

Result:
```
filepath                          bytes
--------------------------------  -----
my_large_note.md                  5432
another_note.md                   3210
...
```

[more functions...]
```

### Document Sections

1. **Quick Reference** - One-line descriptions
2. **Table Functions** - Detailed reference for each
3. **Scalar Functions** - Detailed reference for each
4. **Types Reference** - How to work with complex types
5. **Parameters** - Common parameters
6. **Examples** - Example queries using each function

### Functions to Document

**Table Functions:**
- read_markdown()
- read_markdown_blocks()
- read_markdown_sections()

**Scalar Functions:**
- md_extract_code_blocks()
- md_extract_links()
- md_extract_images()
- md_extract_metadata()
- md_stats()
- md_to_text()
- md_to_html()
- md_to_markdown() (if available)

**Utility:**
- UNNEST() - for expanding lists
- LATERAL - for joining with unnested values
- Standard string/array functions

## Dependencies

- ✅ [research-b8f3d2a1-duckdb-go-markdown.md](.memory/research-b8f3d2a1-duckdb-go-markdown.md) - Function research
- ✅ Specification document
- ✅ DuckDB markdown extension documentation

## Blockers

- Need access to exact function signatures (check research doc)

## Time Estimate

- Structure: 5 minutes
- Table functions: 10 minutes
- Scalar functions: 15 minutes
- Types reference: 5 minutes
- Examples: 10 minutes
- Review: 5 minutes
- Total: 45 minutes

## Definition of Done

- [ ] All functions documented
- [ ] Examples verified
- [ ] Document is complete
- [ ] Properly formatted
- [ ] Linked from main docs
- [ ] Ready for publication

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Priority**: Documentation/Story 3
**Links**:
- [Research: DuckDB Markdown](../research-b8f3d2a1-duckdb-go-markdown.md)
- [User Guide](task-66c1bc07-user-guide.md)
