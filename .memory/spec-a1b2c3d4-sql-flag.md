# Specification: SQL Flag for Search Command

**Spec ID**: a1b2c3d4
**Epic**: [TBD - No epic defined yet]
**Status**: Planning
**Priority**: Medium
**Created**: 2026-01-09
**Updated**: 2026-01-17 (Reorganized as spec file with stories)

## Overview

Add a `--sql` flag to the `search` command that allows users to write custom SQL queries against their notebook's markdown files using DuckDB's markdown extension capabilities.

**Key Finding**: Much of the infrastructure already exists! 
- ✅ DbService has `Query()` method returning `[]map[string]interface{}`
- ✅ NoteService has `Query()` method wrapping DbService
- ✅ Markdown extension already loaded
- ✅ `rowsToMaps()` helper already implemented

We only need to add:
1. Query validation wrapper
2. Read-only connection option
3. Table display formatter
4. CLI integration

## Motivation

**Current State**:
- Users are limited to predefined search filters (`--query`, `--tag`, `--path`)
- No way to leverage DuckDB's powerful SQL capabilities
- Cannot use markdown extension's advanced features (code block extraction, metadata parsing, etc.)
- `NoteService.Query()` exists but isn't exposed to CLI

**Desired State**:
- Power users can write custom SQL queries via CLI
- Access to full markdown extension functionality
- Complex queries combining multiple conditions
- Ability to extract structured data from markdown

**User Value**:
- Flexibility for complex searches
- Access to DuckDB markdown extension features
- Better data extraction and analysis capabilities
- No need to export data for analysis

## User Stories

### Story 1: Extract Code Blocks
**As a** developer maintaining documentation
**I want to** extract all Python code blocks from my notes
**So that** I can validate code examples are up-to-date

```bash
opennotes search --sql "
  SELECT filepath, cb.language, cb.code 
  FROM read_markdown('**/*.md', include_filepath:=true) as md,
  LATERAL UNNEST(md_extract_code_blocks(md.content)) as cb
  WHERE cb.language = 'python'
"
```

### Story 2: Find Notes by Word Count
**As a** content writer
**I want to** find notes by word count
**So that** I can identify which articles need expansion

```bash
opennotes search --sql "
  SELECT filepath, (md_stats(content)).word_count as words
  FROM read_markdown('**/*.md', include_filepath:=true)
  WHERE (md_stats(content)).word_count < 300
  ORDER BY words ASC
"
```

### Story 3: Analyze Metadata
**As a** project manager
**I want to** query notes by frontmatter metadata
**So that** I can track project status across notes

```bash
opennotes search --sql "
  SELECT filepath, 
         md_extract_metadata(content)['status'] as status,
         md_extract_metadata(content)['priority'] as priority
  FROM read_markdown('**/*.md', include_filepath:=true)
  WHERE md_extract_metadata(content)['status'] = 'in-progress'
  ORDER BY priority DESC
"
```

### Story 4: Full-Text Search with Context
**As a** researcher
**I want to** search text with surrounding context
**So that** I can see how terms are used

```bash
opennotes search --sql "
  SELECT filepath, element_type, content
  FROM read_markdown_blocks('**/*.md')
  WHERE md_to_text(content) LIKE '%important concept%'
"
```

## Requirements

### Functional Requirements

**FR-1**: Add `--sql` flag to search command
- **Priority**: MUST
- **Description**: Command line flag that accepts SQL query string
- **Acceptance**: `opennotes search --sql "SELECT ..."`

**FR-2**: Validate user SQL query
- **Priority**: MUST
- **Description**: Validate query is safe before execution
- **Acceptance**: Only SELECT/WITH queries allowed, dangerous keywords blocked

**FR-3**: Execute SQL in read-only mode
- **Priority**: MUST
- **Description**: SQL queries cannot modify data
- **Acceptance**: New read-only connection created for query

**FR-4**: Display query results in table format
- **Priority**: MUST
- **Description**: Format results as readable table in terminal
- **Acceptance**: Results displayed in aligned columns with headers

**FR-5**: Query timeout
- **Priority**: MUST
- **Description**: Prevent runaway queries
- **Acceptance**: Queries timeout after 30 seconds

**FR-6**: Error handling
- **Priority**: MUST
- **Description**: Clear error messages for SQL errors
- **Acceptance**: DuckDB errors translated to user-friendly messages

**FR-7**: Help documentation
- **Priority**: MUST
- **Description**: Document `--sql` flag usage and examples
- **Acceptance**: `opennotes search --help` shows SQL flag with examples

**FR-8**: Schema documentation
- **Priority**: SHOULD
- **Description**: Document available functions for queries
- **Acceptance**: User guide includes function reference

**FR-9**: Example queries
- **Priority**: SHOULD
- **Description**: Provide example queries in documentation
- **Acceptance**: README has 5+ example queries

### Non-Functional Requirements

**NFR-1**: Performance
- **Description**: SQL queries should execute in reasonable time
- **Target**: < 1 second for typical notebook (< 1000 files)

**NFR-2**: Security
- **Description**: Prevent SQL injection and data modification
- **Target**: Zero SQL injection vulnerabilities

**NFR-3**: Usability
- **Description**: Error messages should be helpful
- **Target**: Users can understand and fix query errors

**NFR-4**: Compatibility
- **Description**: Works across all platforms
- **Target**: All platforms supported by current OpenNotes

## Technical Design

### Architecture

```
┌─────────────────────┐
│  cmd/search.go      │
│  --sql flag         │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────────────┐
│ NoteService                 │
│ .ExecuteSQLSafe() [NEW]     │
│   ├─ validateSQLQuery()     │
│   ├─ GetReadOnlyDB()        │
│   └─ Execute + timeout      │
└──────────┬──────────────────┘
           │
           ▼
┌─────────────────────────────┐
│ DbService                   │
│ .GetReadOnlyDB() [NEW]      │
│ .Query() [EXISTS] ✅        │
└──────────┬──────────────────┘
           │
           ▼
┌─────────────────────┐
│ DuckDB + markdown   │
│ extension           │
└─────────────────────┘
```

### Component Changes

**1. internal/services/db.go** (NEW METHOD)
```go
// GetReadOnlyDB returns a read-only database connection.
// This creates a new connection separate from the singleton.
func (d *DbService) GetReadOnlyDB(ctx context.Context) (*sql.DB, error) {
    d.log.Debug().Msg("creating read-only database connection")
    
    // Open in read-only mode
    db, err := sql.Open("duckdb", "?access_mode=READ_ONLY")
    if err != nil {
        return nil, fmt.Errorf("failed to open read-only database: %w", err)
    }
    
    // Load markdown extension (INSTALL not needed, use cached)
    if _, err := db.ExecContext(ctx, "LOAD markdown"); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to load markdown extension: %w", err)
    }
    
    return db, nil
}
```

**2. internal/services/note.go** (NEW METHOD)
```go
// ExecuteSQLSafe executes a validated SQL query in read-only mode.
func (s *NoteService) ExecuteSQLSafe(ctx context.Context, query string) ([]map[string]any, error) {
    // Validate query
    if err := validateSQLQuery(query); err != nil {
        return nil, err
    }
    
    // Get read-only connection
    db, err := s.dbService.GetReadOnlyDB(ctx)
    if err != nil {
        return nil, err
    }
    defer db.Close()
    
    // Set timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Execute query
    rows, err := db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()
    
    // Reuse existing rowsToMaps logic
    return rowsToMaps(rows)
}

func validateSQLQuery(query string) error {
    q := strings.TrimSpace(strings.ToUpper(query))
    
    if !strings.HasPrefix(q, "SELECT") && !strings.HasPrefix(q, "WITH") {
        return fmt.Errorf("only SELECT queries are allowed")
    }
    
    // Block dangerous keywords
    dangerous := []string{
        "DROP", "DELETE", "UPDATE", "INSERT", 
        "ALTER", "CREATE", "TRUNCATE", "REPLACE",
        "ATTACH", "DETACH", "PRAGMA",
    }
    
    for _, keyword := range dangerous {
        if strings.Contains(q, keyword) {
            return fmt.Errorf("keyword %s is not allowed", keyword)
        }
    }
    
    return nil
}

// Helper to convert rows to maps (can extract from DbService)
func rowsToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
    columns, err := rows.Columns()
    if err != nil {
        return nil, err
    }

    var results []map[string]interface{}
    for rows.Next() {
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        for i := range values {
            valuePtrs[i] = &values[i]
        }

        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, err
        }

        row := make(map[string]interface{})
        for i, col := range columns {
            row[col] = values[i]
        }
        results = append(results, row)
    }

    return results, rows.Err()
}
```

**3. internal/services/display.go** (NEW METHOD)
```go
// RenderSQLResults renders SQL query results as a table.
func (s *DisplayService) RenderSQLResults(results []map[string]interface{}) error {
    if len(results) == 0 {
        fmt.Println("No results")
        return nil
    }
    
    // Get columns from first result
    var columns []string
    for col := range results[0] {
        columns = append(columns, col)
    }
    sort.Strings(columns)
    
    // Calculate column widths
    widths := make(map[string]int)
    for _, col := range columns {
        widths[col] = len(col)
    }
    
    for _, row := range results {
        for _, col := range columns {
            val := fmt.Sprintf("%v", row[col])
            if len(val) > widths[col] {
                widths[col] = len(val)
            }
        }
    }
    
    // Print header
    for _, col := range columns {
        fmt.Printf("%-*s  ", widths[col], col)
    }
    fmt.Println()
    
    // Print separator
    for _, col := range columns {
        fmt.Print(strings.Repeat("-", widths[col]), "  ")
    }
    fmt.Println()
    
    // Print rows
    for _, row := range results {
        for _, col := range columns {
            val := fmt.Sprintf("%v", row[col])
            fmt.Printf("%-*s  ", widths[col], val)
        }
        fmt.Println()
    }
    
    fmt.Printf("\n%d rows\n", len(results))
    return nil
}
```

**4. cmd/search.go** (MODIFIED)
```go
// Add flag
sqlQuery := cmd.Flags().String("sql", "", "Execute custom SQL query")

// In command handler (early in function)
if *sqlQuery != "" {
    results, err := noteService.ExecuteSQLSafe(ctx, *sqlQuery)
    if err != nil {
        return fmt.Errorf("SQL query failed: %w", err)
    }
    return displayService.RenderSQLResults(results)
}
```

### Data Flow

1. User runs: `opennotes search --sql "SELECT ..."`
2. Search command parses `--sql` flag
3. If `--sql` present, bypass normal search
4. Call `NoteService.ExecuteSQLSafe(ctx, query)`
   - Validate query (SELECT/WITH only, no dangerous keywords)
   - Get read-only connection from DbService
   - Execute with 30s timeout
   - Convert rows to maps (using existing logic)
5. Call `DisplayService.RenderSQLResults(results)`
   - Calculate column widths
   - Print formatted table
6. Exit (skip normal search logic)

### Query Validation Logic

```go
func validateSQLQuery(query string) error {
    q := strings.TrimSpace(strings.ToUpper(query))
    
    // Allow SELECT and WITH (for CTEs)
    if !strings.HasPrefix(q, "SELECT") && !strings.HasPrefix(q, "WITH") {
        return fmt.Errorf("only SELECT queries are allowed")
    }
    
    // Block dangerous keywords
    dangerous := []string{
        "DROP", "DELETE", "UPDATE", "INSERT", 
        "ALTER", "CREATE", "TRUNCATE", "REPLACE",
        "ATTACH", "DETACH", "PRAGMA",
    }
    
    for _, keyword := range dangerous {
        if strings.Contains(q, keyword) {
            return fmt.Errorf("keyword %s is not allowed", keyword)
        }
    }
    
    return nil
}
```

### Schema Available to Users

Users can use any DuckDB SQL with these markdown extension functions:

**Table Functions:**
```sql
read_markdown(path_pattern, include_filepath:=true)
read_markdown_blocks(path_pattern)
read_markdown_sections(path_pattern)
```

**Scalar Functions:**
```sql
md_extract_code_blocks(content) -> LIST
md_extract_links(content) -> LIST
md_extract_images(content) -> LIST
md_extract_metadata(content) -> MAP
md_stats(content) -> STRUCT
md_to_text(content) -> VARCHAR
md_to_html(content) -> VARCHAR
```

**Example Query Pattern:**
```sql
SELECT * FROM read_markdown('**/*.md', include_filepath:=true)
WHERE md_to_text(content) LIKE '%search term%'
```

## Tasks

This phase contains the following tasks (to be created as separate task files):

### Story 1: Core Functionality (MVP)
- [ ] `task-xxxxxxxx-add-readonly-db.md` - Add `DbService.GetReadOnlyDB()` method
- [ ] `task-xxxxxxxx-sql-validation.md` - Add `NoteService.validateSQLQuery()` helper
- [ ] `task-xxxxxxxx-execute-sql-safe.md` - Add `NoteService.ExecuteSQLSafe()` method
- [ ] `task-xxxxxxxx-render-sql-results.md` - Add `DisplayService.RenderSQLResults()` method
- [ ] `task-xxxxxxxx-sql-flag-cli.md` - Add `--sql` flag to `cmd/search.go`
- [ ] `task-xxxxxxxx-sql-unit-tests.md` - Write unit tests (>80% coverage)

**Deliverables:**
- Working `--sql` flag with validated query execution
- Simple table output
- Test coverage > 80%

**Estimated Effort**: 3-4 hours (reduced from 4-6 due to existing infrastructure)

### Story 2: Enhanced Display (Optional)
- [ ] `task-xxxxxxxx-table-formatting.md` - Improve table formatting (colors, alignment)
- [ ] `task-xxxxxxxx-content-truncation.md` - Handle long content (truncation with ...)
- [ ] `task-xxxxxxxx-format-flag.md` - Support `--format json|csv|table` flag

**Estimated Effort**: 2-3 hours

### Story 3: Documentation
- [ ] `task-xxxxxxxx-cli-help.md` - Update `opennotes search --help` text
- [ ] `task-xxxxxxxx-user-guide.md` - Write user guide with examples
- [ ] `task-xxxxxxxx-function-docs.md` - Document available SQL functions

**Estimated Effort**: 2-3 hours

### Story 4: Advanced Features (Future)
Ideas for future consideration:
- Query plan visualization (`--explain` flag)
- Query templates/macros
- Interactive SQL shell mode
- Query history in config
- Saved queries per notebook
- Schema introspection command (`opennotes schema`)

## Next Steps

1. Create individual task files for Story 1 (Core Functionality)
2. Begin implementation of MVP tasks
3. Complete testing and validation
4. Move to Story 2 and 3 as needed

## Testing Strategy

### Unit Tests

**DbService.GetReadOnlyDB()**
```go
func TestGetReadOnlyDB(t *testing.T) {
    // ✓ Returns valid connection
    // ✓ Markdown extension is loaded
    // ✓ Write operations fail
}
```

**NoteService.validateSQLQuery()**
```go
func TestValidateSQLQuery(t *testing.T) {
    // ✓ SELECT allowed
    // ✓ WITH (CTE) allowed
    // ✓ INSERT blocked
    // ✓ UPDATE blocked
    // ✓ DROP blocked
    // ✓ DELETE blocked
    // ✓ PRAGMA blocked
    // ✓ Case insensitive
}
```

**NoteService.ExecuteSQLSafe()**
```go
func TestExecuteSQLSafe(t *testing.T) {
    // ✓ Valid SELECT succeeds
    // ✓ Invalid query blocked
    // ✓ Timeout works
    // ✓ Returns correct structure
    // ✓ Handles errors properly
}
```

**DisplayService.RenderSQLResults()**
```go
func TestRenderSQLResults(t *testing.T) {
    // ✓ Empty results
    // ✓ Single row
    // ✓ Multiple rows
    // ✓ Different column types
    // ✓ Wide columns
}
```

### Integration Tests

**End-to-End**
```bash
# Create test notebook with markdown files
# Run SQL queries
# Verify output
```

Test cases:
- ✓ Simple SELECT query
- ✓ Query with markdown functions
- ✓ Query with WHERE clause
- ✓ Error handling
- ✓ Timeout handling

### Manual Testing

- [ ] Extract code blocks from real notes
- [ ] Find notes by word count
- [ ] Search metadata fields
- [ ] Complex queries with JOINs
- [ ] Error messages are clear
- [ ] Performance is acceptable

## Security Considerations

### Threat Model

**T-1**: SQL Injection
- **Mitigation**: Query validation, read-only mode
- **Risk**: LOW (local data, user-owned)

**T-2**: Denial of Service
- **Mitigation**: 30s timeout
- **Risk**: LOW (local only)

**T-3**: Data Modification
- **Mitigation**: Read-only connection
- **Risk**: NONE (prevented by database)

### Security Controls

1. **Query Validation**: Only SELECT/WITH allowed
2. **Keyword Blacklist**: Block dangerous SQL keywords
3. **Read-Only Connection**: New connection with `access_mode=READ_ONLY`
4. **Timeout**: 30 second execution limit
5. **No Interpolation**: User provides full query (no parameter substitution needed)

### Security Documentation

```
⚠️  SECURITY WARNING

The --sql flag executes SQL queries directly against DuckDB.

Security measures:
- Read-only mode enforced
- Only SELECT queries allowed
- Dangerous keywords blocked
- 30-second timeout

However:
- Query validation is not foolproof
- This is a power user feature
- Use caution with untrusted queries

Your local data only - no network access.
```

## Documentation

### Command Help Text

```
Flags:
  --sql string     Execute custom SQL query against notes
  
Examples:
  # Simple SELECT
  opennotes search --sql "SELECT * FROM read_markdown('**/*.md', include_filepath:=true)"
  
  # Extract code blocks
  opennotes search --sql "
    SELECT filepath, cb.language, cb.code 
    FROM read_markdown('**/*.md', include_filepath:=true),
    LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
    WHERE cb.language = 'python'
  "
  
  # Query metadata
  opennotes search --sql "
    SELECT filepath, md_extract_metadata(content)['status']
    FROM read_markdown('**/*.md', include_filepath:=true)
  "

Available Functions:
  read_markdown(), md_extract_code_blocks(), md_extract_metadata(),
  md_stats(), md_to_text(), md_to_html(), and more.
  
See: opennotes help sql (or check docs/sql-guide.md)
```

### README Section

```markdown
## Advanced: SQL Queries

For power users, OpenNotes supports custom SQL queries:

\`\`\`bash
opennotes search --sql "SELECT * FROM read_markdown('**/*.md', include_filepath:=true)"
\`\`\`

Access DuckDB's markdown extension:
- Extract code blocks by language
- Query frontmatter metadata
- Calculate statistics (word count, reading time)
- Complex filtering and joins

See [SQL Guide](docs/sql-guide.md) for examples.

⚠️ **Note**: Read-only access. Only SELECT queries allowed.
```

## Success Metrics

### Launch Criteria

- [ ] All FR MUST requirements implemented
- [ ] Test coverage > 80%
- [ ] Documentation complete
- [ ] Reviewed and approved
- [ ] Tested on Linux (primary platform)

### Success Indicators

**Adoption**: 10% try within first month
**Quality**: < 5 bugs reported
**Performance**: 95% queries < 1 second

## Open Questions

1. **Q**: Extract `rowsToMaps()` to shared utility?
   **A**: Yes - create `internal/core/sql_utils.go`

2. **Q**: Should validation be more lenient (allow EXPLAIN)?
   **A**: Phase 2 - add `--explain` flag explicitly

3. **Q**: Limit result set size?
   **A**: Yes - add LIMIT 1000 if user doesn't specify

4. **Q**: Should we scope queries to current notebook?
   **A**: Let user control via glob pattern in query

## References

- Research: [research-b8f3d2a1-duckdb-go-markdown.md](.memory/research-b8f3d2a1-duckdb-go-markdown.md)
- Current DB: `internal/services/db.go`
- Current Note: `internal/services/note.go`
- DuckDB Docs: https://duckdb.org/docs/stable/clients/go
- Markdown Ext: https://duckdb.org/community_extensions/extensions/markdown

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-01-09 | Agent | Initial specification |
| 2026-01-09 | Agent | Aligned with current implementation, simplified plan |
