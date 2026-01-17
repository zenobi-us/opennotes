# Task: Improve Table Formatting

**Epic**: [SQL Flag Feature](epic-2f3c4d5e-sql-flag-feature.md)

**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 2 - Enhanced Display (Optional)
**Priority**: MEDIUM
**Complexity**: Medium
**Estimated Time**: 60 minutes

## Objective

Enhance the basic table display with improved formatting: colors, borders, column alignment options, and visual polish for better readability.

## Context

Story 1 delivers basic table display. Story 2 adds refinements:
- Color support for headers and alternating rows
- Optional table borders (ASCII box drawing)
- Type-based alignment (right-align numbers, left-align text)
- Better visual hierarchy

This depends on Story 1's core display functionality.

## Steps to Take

1. **Add color constants/support**
   - File: `internal/services/display.go`
   - Add color codes for headers, alternating rows
   - Use ANSI color codes or existing color library in project
   - Make colors optional (graceful degradation if terminal doesn't support)

2. **Implement type-based alignment**
   - Detect column data types from sample values
   - Numbers: right-aligned
   - Strings: left-aligned
   - Booleans: center-aligned
   - Dates: center-aligned

3. **Add optional border mode**
   - Add parameter or detect terminal capability
   - Use ASCII box drawing characters (if UTF-8 supported)
   - Box corners: ┌ ┐ └ ┘
   - Lines: ─ │ ├ ┤ ┬ ┴ ┼
   - Fallback to `+`, `-`, `|` if ASCII-only

4. **Implement alternating row colors**
   - Alternate background colors for readability
   - Subtle color (light gray) for even rows
   - Normal for odd rows

5. **Add truncation with ellipsis** (optional)
   - For very wide columns, truncate with "..."
   - Configurable width limit
   - Show full content in separate line if needed

## Expected Outcomes

- [ ] Colors applied to headers and rows
- [ ] Numbers right-aligned
- [ ] Strings left-aligned
- [ ] Optional borders (ASCII or Unicode)
- [ ] Alternating row colors
- [ ] Professional appearance
- [ ] Works on different terminals

## Acceptance Criteria

- [x] Header row has distinct color/styling
- [x] Numbers are right-aligned
- [x] Strings are left-aligned
- [x] Borders work in ASCII mode
- [x] Borders work in Unicode mode
- [x] Alternating rows have distinct background
- [x] Colors don't break on terminals without support
- [x] Output is readable and professional
- [x] Configuration options available
- [x] Backward compatible with Story 1 output

## Implementation Notes

### Color Support Pattern

```go
const (
    colorReset = "\033[0m"
    colorBold = "\033[1m"
    colorDim = "\033[2m"
    colorGray = "\033[90m"
    colorBgGray = "\033[47m"
)

// Use conditionally based on terminal support
if supportsColor {
    fmt.Print(colorBold)
}
fmt.Print(header)
if supportsColor {
    fmt.Print(colorReset)
}
```

### Type Detection Pattern

```go
func detectType(value interface{}) string {
    switch value.(type) {
    case int, int32, int64, float32, float64:
        return "number"
    case bool:
        return "bool"
    case string:
        // Check if looks like date
        if isDate(value.(string)) {
            return "date"
        }
        return "string"
    default:
        return "string"
    }
}
```

### Alignment Pattern

```go
func alignValue(value string, align string, width int) string {
    switch align {
    case "right":
        return fmt.Sprintf("%*s", width, value)
    case "center":
        return center(value, width)
    default: // "left"
        return fmt.Sprintf("%-*s", width, value)
    }
}
```

## Dependencies

- ✅ [task-c7fc4f57-render-sql-results.md](task-c7fc4f57-render-sql-results.md) - Base display
- ⏳ Color library (check if available in project)

## Blockers

- Need to determine if project uses a color library or relies on ANSI codes

## Time Estimate

- Color implementation: 20 minutes
- Type detection: 15 minutes
- Alignment: 15 minutes
- Border support: 15 minutes
- Testing: 10 minutes
- Review: 5 minutes
- Total: 60 minutes

## Definition of Done

- [ ] Colors applied to headers and rows
- [ ] Type-based alignment working
- [ ] Borders render correctly
- [ ] Alternating rows visible
- [ ] Tests added for formatting
- [ ] Works cross-platform
- [ ] Documentation updated

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Priority**: Optional/Story 2
**Links**:
- [Render SQL Results](task-c7fc4f57-render-sql-results.md)
- [Format Flag](task-35b138e9-format-flag.md)
