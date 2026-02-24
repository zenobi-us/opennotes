---
id: c1d2e3f4
title: Option 2 Refactor - GroupResults Return Structure
type: "learning"
created_at: 2026-01-28T22:43:00+10:30
updated_at: 2026-01-28T22:50:00+10:30
status: completed
tags: [refactoring, api-design, go-patterns, json-serialization]
---

# Learning: Option 2 Refactor for GroupResults Return Structure

## Summary

Successfully refactored the `GroupResults()` return type from **Option 3 (Hybrid with Metadata)** to **Option 2 (Pure Grouped/Flat Structure)**. This learning captures the design decision, implementation details, and benefits achieved.

## Context

The view system needed a clean return structure for grouped vs flat results. Three options were researched:
- **Option 1**: Always return array (flat structure)
- **Option 2**: Polymorphic return (map for grouped, array for flat) ← **Chosen**
- **Option 3**: Hybrid with metadata wrapper

## Key Design Decision

**Option 2 was chosen** because it provides:
- ✅ **Cleaner JSON**: No metadata wrapper (is_grouped, group_by fields)
- ✅ **Pure Data**: Just the data, nothing else (data-centric approach)
- ✅ **Polymorphic**: Clients use type assertion to handle grouped vs flat
- ✅ **Smaller Payload**: No redundant metadata fields
- ✅ **Simple Semantics**: Array for lists, map for grouped data

## Implementation Details

### Return Structure

```go
// GroupResults() now returns interface{} with pure structure:
func (vs *ViewService) GroupResults(view *core.ViewDefinition, rows []map[string]interface{}) interface{}

// Return values:
// - Grouped views (GROUP BY specified): map[string][]map[string]interface{}
// - Flat views (no GROUP BY): []map[string]interface{}
```

### JSON Output Examples

**Flat View (No GROUP BY)**:
```json
[
  {"id": "1", "title": "Note 1", "status": "todo"},
  {"id": "2", "title": "Note 2", "status": "done"}
]
```

**Grouped View (GROUP BY status)**:
```json
{
  "done": [{"id": "2", "title": "Note 2", "status": "done"}],
  "todo": [
    {"id": "1", "title": "Note 1", "status": "todo"},
    {"id": "3", "title": "Note 3", "status": "todo"}
  ]
}
```

### Changes Made

1. **Removed ViewResults Type** (`internal/core/view.go`)
   - Struct contained: `IsGrouped`, `GroupBy`, `Grouped`, `Flat` fields
   - Now returns raw `interface{}` type

2. **Updated GroupResults() Method** (`internal/services/view.go`)
   - Old: `func (vs *ViewService) GroupResults(...) *core.ViewResults`
   - New: `func (vs *ViewService) GroupResults(...) interface{}`

3. **Updated Test Suite** (`internal/services/view_test.go`)
   - 5 test functions updated with proper type assertions
   - All 711+ existing tests still passing

4. **Command Handler** (`cmd/notes_view.go`)
   - No changes needed - `json.Marshal()` handles both types correctly

## Implications for Future Work

1. **API Consumers**: Must check type to determine grouped vs flat:
   ```go
   switch v := viewResults.(type) {
   case map[string][]map[string]interface{}:
       // Handle grouped
   case []map[string]interface{}:
       // Handle flat
   }
   ```

2. **Go's json.Marshal()**: Naturally serializes both types correctly, no special handling needed.

3. **Pattern Reuse**: This polymorphic return pattern can be applied to similar APIs where output structure depends on input configuration.

## Test Coverage

All GroupResults tests passing:
- ✅ TestViewService_GroupResults_FlatResults
- ✅ TestViewService_GroupResults_GroupedByString
- ✅ TestViewService_GroupResults_GroupedByNumber
- ✅ TestViewService_GroupResults_EmptyResults
- ✅ TestViewService_GroupResults_NullValues

## Related Documents

- **Research**: `.memory/research-e5f6g7h8-kanban-group-by-return-structure.md`
- **Comparison**: `.memory/research-a1b2c3d4-kanban-return-structure-comparison.md`

## Commit Reference

**Commit hash**: 52c7210
**Branch**: fix/kanban-view
