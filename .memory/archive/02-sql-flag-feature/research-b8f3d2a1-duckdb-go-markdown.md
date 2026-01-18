# Research: DuckDB Go Client and Markdown Extension

**Date**: 2026-01-09
**Purpose**: Research DuckDB Go client capabilities and markdown extension for implementing `--sql` flag in search command

## Current OpenNotes Implementation Analysis

### Database Service (internal/services/db.go)

**Current Architecture:**
```go
type DbService struct {
    db   *sql.DB
    once sync.Once
    mu   sync.Mutex
    log  zerolog.Logger
}
```

**Key Patterns:**

1. **Singleton Pattern**: Uses `sync.Once` for lazy initialization
2. **In-Memory Database**: Opens with `sql.Open("duckdb", "")`
3. **Extension Loading**: Installs markdown extension on first connection
4. **Query Helper**: Has `Query()` method that returns `[]map[string]interface{}`
5. **Row Mapping**: Already implements `rowsToMaps()` for generic queries

**Initialization Flow:**
```go
func (d *DbService) GetDB(ctx context.Context) (*sql.DB, error) {
    d.once.Do(func() {
        db, err := sql.Open("duckdb", "")
        db.ExecContext(ctx, "INSTALL markdown FROM community")
        db.ExecContext(ctx, "LOAD markdown")
        d.db = db
    })
    return d.db, nil
}
```

**Important Finding**: DbService already has the infrastructure we need!
- ✅ Already has `Query()` method returning maps
- ✅ Already has `rowsToMaps()` helper
- ✅ Already initializes markdown extension
- ✅ Already handles context properly

### Note Service (internal/services/note.go)

**Current Query Pattern:**
```go
func (s *NoteService) SearchNotes(ctx context.Context, query string) ([]Note, error) {
    db, err := s.dbService.GetDB(ctx)
    glob := filepath.Join(s.notebookPath, "**", "*.md")
    
    // Uses read_markdown with include_filepath parameter
    sqlQuery := `SELECT * FROM read_markdown(?, include_filepath:=true)`
    rows, err := db.QueryContext(ctx, sqlQuery, glob)
    
    // Maps to Note struct
    // Filters by query string in code (not SQL)
}
```

**Current Features:**
- Uses parameterized queries (safe)
- Includes filepath in results
- Handles metadata automatically
- Filters in Go code (not SQL WHERE clause)
- Has `Query()` method that calls `dbService.Query()`

**Important Finding**: NoteService already has `Query()` method!
```go
func (s *NoteService) Query(ctx context.Context, sql string) ([]map[string]any, error) {
    return s.dbService.Query(ctx, sql)
}
```

This means raw SQL execution is already possible via `NoteService.Query()`.

## DuckDB Go Client

### Installation & Import

```go
// Install (already in go.mod)
go get github.com/duckdb/duckdb-go/v2

// Import (already in db.go)
import (
    "database/sql"
    _ "github.com/duckdb/duckdb-go/v2"
)
```

### Key Findings

1. **Official Primary Client**: Maintained by DuckDB team, primary support tier
2. **Current Version**: v1.4.3 (DuckDB), v2.4.2+ (go-duckdb)
3. **Standard Interface**: Uses Go's `database/sql` interface
4. **Repository**: `github.com/duckdb/duckdb-go` (as used in OpenNotes)

### Connection Patterns (Relevant to OpenNotes)

```go
// In-memory database (current OpenNotes usage)
db, err := sql.Open("duckdb", "")

// Read-only mode (for --sql flag)
db, err := sql.Open("duckdb", "?access_mode=READ_ONLY")

// With configuration options
db, err := sql.Open("duckdb", "?access_mode=READ_ONLY&threads=4")
```

**Important**: DSN format supports query parameters for configuration.

### Query Execution (Already in Use)

OpenNotes already uses these patterns:
```go
// Execute query with parameters
rows, err := db.QueryContext(ctx, sqlQuery, glob)

// Scan results
var values []interface{}
rows.Scan(valuePtrs...)

// Convert to maps (already implemented)
rowsToMaps(rows)
```

## DuckDB Markdown Extension

### Current OpenNotes Usage

Already installed and loaded in DbService initialization:
```go
db.ExecContext(ctx, "INSTALL markdown FROM community")
db.ExecContext(ctx, "LOAD markdown")
```

### Functions Used by OpenNotes

**Current:**
```sql
SELECT * FROM read_markdown(?, include_filepath:=true)
```

**Available for --sql flag:**
- `read_markdown(path_pattern, ...)` - Already used ✅
- `read_markdown_blocks(path_pattern)` - Block-level parsing
- `read_markdown_sections(path_pattern)` - Section-level parsing
- `md_extract_code_blocks(content)` - Extract code blocks
- `md_extract_links(content)` - Extract links
- `md_extract_images(content)` - Extract images
- `md_extract_metadata(content)` - Extract frontmatter
- `md_stats(content)` - Get statistics
- `md_to_text(content)` - Convert to plain text
- `md_to_html(content)` - Convert to HTML

### Performance

- Processes 287 files (2,699 sections) in ~603ms
- ~4,000 sections/second processing rate
- Already loaded and ready to use

## Implementation Simplifications

### What We DON'T Need to Build

❌ ~~New DB connection initialization~~ - Already exists
❌ ~~Markdown extension loading~~ - Already happens
❌ ~~Query execution method~~ - Already exists (`NoteService.Query()`)
❌ ~~Row to map conversion~~ - Already exists (`rowsToMaps()`)

### What We DO Need to Build

✅ Query validation (SELECT only, keyword blocking)
✅ Read-only connection option
✅ Result display formatting (table output)
✅ CLI flag and command integration
✅ Error handling and user messages
✅ Documentation

### Simplified Architecture

```
cmd/search.go
    │
    ├─ Parse --sql flag
    │
    └─ Call NoteService.ExecuteSQLSafe(ctx, query)
            │
            ├─ Validate query ✅ NEW
            ├─ Call s.Query(ctx, query) ✅ EXISTS
            └─ Return []map[string]interface{}
```

Then in DisplayService:
```
DisplayService.RenderSQLResults(results) ✅ NEW
    │
    └─ Format as table and print
```

## Revised Implementation Strategy

### Approach: Leverage Existing Infrastructure

Since `NoteService.Query()` already exists and returns generic maps,
we only need:

1. **Validation Wrapper** in NoteService
2. **Table Display** in DisplayService
3. **CLI Integration** in cmd/search.go

### Read-Only Mode Challenge

**Problem**: Current DbService is singleton with shared connection.
Can't make it read-only after initialization.

**Solutions:**

**Option A: Add Read-Only Flag to Existing Connection**
```go
// In DbService.GetDB(), check if we need read-only
db, err := sql.Open("duckdb", "?access_mode=READ_ONLY")
```
❌ Breaks existing writes (if any)

**Option B: Create Separate Read-Only Connection**
```go
func (d *DbService) GetReadOnlyDB(ctx context.Context) (*sql.DB, error) {
    // Create new connection in read-only mode
    db, err := sql.Open("duckdb", "?access_mode=READ_ONLY")
    // Still load markdown extension
    return db, nil
}
```
✅ Clean, doesn't affect existing functionality
✅ Can still install/load extension in read-only mode

**Option C: Validate Query Instead**
```go
// Don't use read-only mode, just validate query
ValidateQuery(query) // Block INSERT, UPDATE, DELETE, etc.
```
✅ Simpler, no new connection needed
⚠️ Less secure (validation could be bypassed)

**Recommendation**: Option B - Separate read-only connection
- Better security through defense in depth
- Clean separation of concerns
- No impact on existing code

### Code Examples for Implementation

#### 1. Add Read-Only Connection to DbService

```go
// GetReadOnlyDB returns a read-only database connection.
func (d *DbService) GetReadOnlyDB(ctx context.Context) (*sql.DB, error) {
    d.log.Debug().Msg("creating read-only database connection")
    
    // Open in read-only mode
    db, err := sql.Open("duckdb", "?access_mode=READ_ONLY")
    if err != nil {
        return nil, fmt.Errorf("failed to open read-only database: %w", err)
    }
    
    // Load markdown extension (read-only still allows loading extensions)
    if _, err := db.ExecContext(ctx, "LOAD markdown"); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to load markdown extension: %w", err)
    }
    
    return db, nil
}
```

#### 2. Add Safe SQL Execution to NoteService

```go
// ExecuteSQLSafe executes a validated SQL query.
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
    
    // Use existing rowsToMaps (from DbService)
    // We can move this to a shared location or duplicate it
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
```

#### 3. Add Table Display to DisplayService

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
    sort.Strings(columns) // Consistent order
    
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

#### 4. CLI Integration (cmd/search.go)

```go
sqlQuery := cmd.Flags().String("sql", "", "Execute custom SQL query")

// In command handler
if *sqlQuery != "" {
    results, err := noteService.ExecuteSQLSafe(ctx, *sqlQuery)
    if err != nil {
        return fmt.Errorf("SQL query failed: %w", err)
    }
    return displayService.RenderSQLResults(results)
}
```

## Schema Available to Users

Users can query with any SQL against these functions:

```sql
-- Read all markdown files
SELECT * FROM read_markdown('**/*.md', include_filepath:=true);

-- Block-level parsing
SELECT * FROM read_markdown_blocks('**/*.md');

-- Extract code blocks
SELECT 
    filename,
    UNNEST(md_extract_code_blocks(content)) as cb
FROM read_markdown('**/*.md');

-- Get statistics
SELECT 
    filename,
    md_stats(content)
FROM read_markdown('**/*.md');

-- Extract metadata
SELECT 
    filename,
    md_extract_metadata(content)
FROM read_markdown('**/*.md');
```

## Testing Considerations

### Existing Test Infrastructure

Looking at `internal/services/db_test.go`:
```go
// Already tests markdown extension loading
rows, err := db.QueryContext(ctx, "SELECT extension_name FROM duckdb_extensions() WHERE extension_name = 'markdown' AND loaded = true")
```

### New Tests Needed

**DbService.GetReadOnlyDB():**
- ✓ Returns valid connection
- ✓ Markdown extension is loaded
- ✓ Write operations fail (INSERT/UPDATE/DELETE)

**NoteService.ExecuteSQLSafe():**
- ✓ Valid SELECT query succeeds
- ✓ Invalid query (INSERT) blocked
- ✓ Dangerous keywords blocked
- ✓ Timeout works
- ✓ Returns correct results

**DisplayService.RenderSQLResults():**
- ✓ Empty results
- ✓ Single row
- ✓ Multiple rows
- ✓ Different column types
- ✓ Long content truncation

## Security Analysis

### Current Security Posture

**Existing NoteService.Query():**
- ❌ No validation
- ❌ No read-only mode
- ❌ No timeout
- ⚠️ Exposed but not used by CLI

**Proposed ExecuteSQLSafe():**
- ✅ Query validation (SELECT only)
- ✅ Keyword blocking
- ✅ Read-only connection
- ✅ 30-second timeout
- ✅ Proper error handling

### Defense in Depth

1. **Query Validation**: First line of defense
2. **Read-Only Mode**: Database-level protection
3. **Timeout**: Resource protection
4. **Error Sanitization**: Don't leak internal details

### Risk Assessment

**Threat**: SQL injection via user input
**Mitigation**: 
- Query validation blocks dangerous keywords
- Read-only mode prevents data modification
- No parameterization needed (user provides full query)
**Risk**: LOW (local data, user-owned)

**Threat**: Denial of service (long query)
**Mitigation**: 30-second timeout
**Risk**: LOW (local only, affects user's own session)

**Threat**: Data exfiltration
**Mitigation**: None needed (user's own local data)
**Risk**: NONE

## Documentation Needs

### User-Facing

1. **Help Text**: Add --sql flag documentation to search command
2. **Examples**: Provide 5-10 common query patterns
3. **Schema Reference**: Document available functions and columns
4. **Security Warning**: Explain this is a power-user feature

### Developer-Facing

1. **Architecture Decision**: Why separate read-only connection
2. **Testing Guide**: How to test SQL features
3. **Extension Guide**: How to add new SQL capabilities

## References

- Current Implementation: `internal/services/db.go`, `internal/services/note.go`
- DuckDB Go Client: https://duckdb.org/docs/stable/clients/go
- Markdown Extension: https://duckdb.org/community_extensions/extensions/markdown
- DuckDB Configuration: https://duckdb.org/docs/stable/configuration/overview
