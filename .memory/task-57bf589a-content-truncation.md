# Task: Handle Long Content Truncation

**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 2 - Enhanced Display (Optional)
**Priority**: MEDIUM
**Complexity**: Low
**Estimated Time**: 30 minutes

## Objective

Add intelligent truncation for long cell content to maintain readable table display without losing information, using ellipsis indicators and line wrapping options.

## Context

Long content (e.g., markdown excerpts, JSON strings) can create very wide columns that break terminal layout. This task implements:
- Truncation with "..." indicator
- Configurable max width per column
- Option to expand to see full content
- Alternative: multi-line cell display

## Steps to Take

1. **Add truncation logic**
   - File: `internal/services/display.go`
   - Function: `truncateValue(value string, maxWidth int) string`
   - If value > maxWidth: truncate to maxWidth-3 and add "..."
   - If value == maxWidth or less: return as-is

2. **Set max column width**
   - Default max width: 50 characters (configurable)
   - Can be overridden via flag (future)
   - Calculate width considering terminal width

3. **Add preview mode**
   - Show truncated content in table
   - Provide option to view full content
   - Example: "Row 1: [press 'v' to view full content]"

4. **Handle special content types**
   - JSON: pretty-print (optional)
   - Markdown: convert to plaintext for display
   - Binary: show as "[binary data]"
   - Very long text: show preview with line count

## Expected Outcomes

- [ ] Long content truncated cleanly
- [ ] Ellipsis clearly indicates truncation
- [ ] Tables remain readable
- [ ] Full content still accessible
- [ ] Special content handled gracefully

## Acceptance Criteria

- [x] Values longer than maxWidth get "..."
- [x] Ellipsis counts toward width
- [x] Readable table output maintained
- [x] No information loss (full content retrievable)
- [x] Works with all data types
- [x] Handles nil/empty values
- [x] Terminal width detection (if possible)
- [x] Graceful degradation

## Implementation Notes

### Truncation Pattern
```go
const defaultMaxWidth = 50

func truncateValue(value string, maxWidth int) string {
    if len(value) <= maxWidth {
        return value
    }
    
    if maxWidth < 3 {
        return "..."
    }
    
    return value[:maxWidth-3] + "..."
}
```

### Terminal Width Detection
```go
func getTerminalWidth() int {
    if width, _, err := term.GetSize(0); err == nil {
        return width
    }
    return 80 // Default fallback
}
```

### Multi-line Cell Option
```go
// Future enhancement: split long cells across lines
func formatCell(value string, maxWidth int, multiline bool) string {
    if multiline && len(value) > maxWidth {
        // Split into lines
        return splitIntoLines(value, maxWidth)
    }
    return truncateValue(value, maxWidth)
}
```

## Dependencies

- ✅ [task-90e473c7-table-formatting.md](task-90e473c7-table-formatting.md) - Formatting
- ✅ Standard Go libraries
- ⏳ Terminal size detection (golang.org/x/term)

## Blockers

- Need to check if terminal size library is already available

## Time Estimate

- Truncation implementation: 15 minutes
- Terminal width detection: 10 minutes
- Testing: 5 minutes
- Total: 30 minutes

## Definition of Done

- [ ] Truncation working correctly
- [ ] Tables remain readable with long content
- [ ] Tests added
- [ ] Terminal width detected automatically
- [ ] No breaking changes to Story 1

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Priority**: Optional/Story 2
**Links**:
- [Table Formatting](task-90e473c7-table-formatting.md)
