# Task: Add DisplayService.RenderSQLResults() Method

**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 1 - Core Functionality (MVP)
**Priority**: HIGH
**Complexity**: Medium
**Estimated Time**: 45 minutes

## Objective

Create a `RenderSQLResults()` method in DisplayService that formats query results as a readable ASCII table with aligned columns and headers.

## Context

User SQL queries return `[]map[string]interface{}`. This method transforms that into a nicely formatted table output suitable for terminal display. The table includes:
- Column headers (sorted alphabetically)
- Separator line
- Data rows with aligned columns
- Row count summary

## Steps to Take

1. **Add RenderSQLResults() method to DisplayService**
   - File: `internal/services/display.go`
   - Signature: `func (s *DisplayService) RenderSQLResults(results []map[string]interface{}) error`

2. **Handle empty results**
   - If `len(results) == 0`, print "No results" and return nil

3. **Extract column names**
   - Get columns from first result map
   - Sort alphabetically for consistent ordering
   - Store in slice

4. **Calculate column widths**
   - Initialize widths map with column name lengths
   - Iterate through all rows
   - Update width if value is wider than current width
   - Use `fmt.Sprintf("%v", value)` for consistent formatting

5. **Print table**
   - Print header row with column names, padded to width
   - Print separator row with dashes
   - Print each data row with values padded to column width
   - Print blank line and row count summary

6. **Use proper formatting**
   - Format string: `fmt.Printf("%-*s  ", width, value)`
   - Two spaces between columns for readability
   - Left-aligned text

## Expected Outcomes

- [ ] RenderSQLResults() method implemented
- [ ] Handles empty result sets
- [ ] Displays column headers
- [ ] Displays separator line
- [ ] Displays all rows
- [ ] Columns properly aligned
- [ ] Column widths calculated correctly
- [ ] Output is readable and professional-looking

## Acceptance Criteria

- [x] Method accepts `[]map[string]interface{}`
- [x] Returns error (or nil on success)
- [x] Empty results show "No results"
- [x] Extracts columns from first result
- [x] Sorts columns alphabetically
- [x] Calculates column widths correctly
- [x] Handles nil and empty values
- [x] Handles different data types (int, string, float, etc.)
- [x] Aligns columns properly with padding
- [x] Shows row count summary
- [x] Uses consistent spacing between columns
- [x] Output is clean and readable
- [x] Code compiles and lints

## Implementation Notes

### Table Format
```
name        email              age
----------  ----------------  ---
John Doe    john@example.com   30
Jane Smith  jane@example.com   28

2 rows
```

### Width Calculation Pattern
```go
widths := make(map[string]int)

// Initialize with header widths
for _, col := range columns {
    widths[col] = len(col)
}

// Update with data widths
for _, row := range results {
    for _, col := range columns {
        val := fmt.Sprintf("%v", row[col])
        if len(val) > widths[col] {
            widths[col] = len(val)
        }
    }
}
```

### Printing Pattern
```go
// Header
for _, col := range columns {
    fmt.Printf("%-*s  ", widths[col], col)
}
fmt.Println()

// Separator
for _, col := range columns {
    fmt.Print(strings.Repeat("-", widths[col]), "  ")
}
fmt.Println()

// Rows
for _, row := range results {
    for _, col := range columns {
        val := fmt.Sprintf("%v", row[col])
        fmt.Printf("%-*s  ", widths[col], val)
    }
    fmt.Println()
}

// Summary
fmt.Printf("\n%d rows\n", len(results))
```

### Edge Cases

```
- Empty results → "No results"
- Single column → works correctly
- Single row → works correctly
- Wide content → gets full width but still readable
- nil values → formatted as empty or "null"
- Mixed types → all converted to string via %v
```

## Dependencies

- ✅ DisplayService already exists
- ✅ Standard Go libraries (fmt, strings, sort)
- ✅ [task-bef53880-execute-sql-safe.md](task-bef53880-execute-sql-safe.md) - produces results

## Blockers

- None identified

## Time Estimate

- Implementation: 30 minutes
- Testing: 10 minutes
- Review: 5 minutes
- Total: 45 minutes

## Definition of Done

- [ ] Method implemented and compiles
- [ ] Handles all edge cases
- [ ] Output is properly formatted
- [ ] Passes linting and formatting
- [ ] Ready for integration test
- [ ] Linked to CLI integration task

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Links**: 
- [Execute SQL Safe](task-bef53880-execute-sql-safe.md)
- [SQL Flag CLI](task-710bd5bd-sql-flag-cli.md)
