# Task: Add Format Flag Support

**Epic**: [SQL Flag Feature](epic-2f3c4d5e-sql-flag-feature.md)

**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 2 - Enhanced Display (Optional)
**Priority**: MEDIUM
**Complexity**: Medium
**Estimated Time**: 60 minutes

## Objective

Add `--format` flag to search command supporting multiple output formats: table (default), JSON, CSV, and TSV for flexibility in data export and programmatic consumption.

## Context

Different users need different output formats:
- **table**: Human-readable (default)
- **json**: Programmatic, easy to parse
- **csv**: Spreadsheet import, data analysis
- **tsv**: Tab-separated variant of CSV

The `--format` flag gives users choice while maintaining backward compatibility.

## Steps to Take

1. **Add --format flag to search command**
   - File: `cmd/search.go`
   - Options: "table" (default), "json", "csv", "tsv"
   - Validate format value
   - Only applies to SQL queries

2. **Add format method to DisplayService**
   - `func (s *DisplayService) RenderResults(results []map[string]any, format string) error`
   - Route to appropriate renderer based on format

3. **Implement JSON formatter**
   - Marshal results to JSON
   - Pretty-print for readability
   - Use standard `encoding/json`

4. **Implement CSV formatter**
   - Use `encoding/csv` package
   - First row = headers
   - Subsequent rows = data
   - Proper quoting/escaping

5. **Implement TSV formatter**
   - Similar to CSV but tab-separated
   - Use same CSV logic with tab delimiter

6. **Add format selection logic**
   - In SQL path, call `displayService.RenderResults(results, format)`
   - Pass format from flag

## Expected Outcomes

- [ ] `--format` flag added
- [ ] Default to "table" format
- [ ] JSON output works
- [ ] CSV output works
- [ ] TSV output works
- [ ] Format validation working
- [ ] Backward compatible

## Acceptance Criteria

- [x] Flag accepts valid formats: table, json, csv, tsv
- [x] Rejects invalid formats with clear error
- [x] Default is "table"
- [x] JSON output is valid JSON
- [x] CSV output is valid CSV (can import to Excel)
- [x] TSV output is valid TSV
- [x] Headers included in output
- [x] All data types properly serialized
- [x] nil values handled (null in JSON, empty in CSV)
- [x] Only applies to --sql queries
- [x] Works with other SQL flags

## Implementation Notes

### Flag Definition
```go
format := cmd.Flags().String("format", "table", "Output format: table, json, csv, tsv")

// Validate
validFormats := []string{"table", "json", "csv", "tsv"}
if !contains(validFormats, *format) {
    return fmt.Errorf("invalid format: %s", *format)
}
```

### JSON Formatter
```go
func (s *DisplayService) renderJSON(results []map[string]any) error {
    data, err := json.MarshalIndent(results, "", "  ")
    if err != nil {
        return err
    }
    fmt.Println(string(data))
    return nil
}
```

### CSV Formatter
```go
import "encoding/csv"

func (s *DisplayService) renderCSV(results []map[string]any) error {
    if len(results) == 0 {
        fmt.Println("No results")
        return nil
    }
    
    w := csv.NewWriter(os.Stdout)
    defer w.Flush()
    
    // Get headers
    headers := getHeaders(results[0])
    w.Write(headers)
    
    // Write rows
    for _, row := range results {
        record := make([]string, len(headers))
        for i, h := range headers {
            record[i] = fmt.Sprintf("%v", row[h])
        }
        w.Write(record)
    }
    
    return nil
}
```

### TSV Formatter
```go
func (s *DisplayService) renderTSV(results []map[string]any) error {
    w := csv.NewWriter(os.Stdout)
    w.Comma = '\t'
    defer w.Flush()
    
    // Same logic as CSV but with tab separator
}
```

## Dependencies

- ✅ [task-c7fc4f57-render-sql-results.md](task-c7fc4f57-render-sql-results.md) - Base display
- ✅ [task-710bd5bd-sql-flag-cli.md](task-710bd5bd-sql-flag-cli.md) - CLI integration
- ✅ Standard Go libraries (encoding/csv, encoding/json)

## Blockers

- None identified

## Time Estimate

- Flag implementation: 15 minutes
- JSON formatter: 15 minutes
- CSV formatter: 15 minutes
- TSV formatter: 10 minutes
- Testing: 15 minutes
- Review: 5 minutes
- Total: 60 minutes

## Definition of Done

- [ ] All formats implemented and working
- [ ] Flag added to search command
- [ ] Format validation working
- [ ] Output verified as valid (JSON/CSV/TSV)
- [ ] Tests added for each format
- [ ] Documentation updated
- [ ] Backward compatible

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Priority**: Optional/Story 2
**Links**:
- [Render SQL Results](task-c7fc4f57-render-sql-results.md)
- [SQL Flag CLI](task-710bd5bd-sql-flag-cli.md)
