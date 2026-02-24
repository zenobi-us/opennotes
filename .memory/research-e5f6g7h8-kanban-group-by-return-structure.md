---
id: e5f6g7h8
title: GROUP BY Return Structure Analysis for Kanban View
type: "research"
created_at: 2026-01-27T23:44:00+10:30
updated_at: 2026-01-28T11:40:00+10:30
status: draft
---

# GROUP BY Return Structure Decision Analysis

## Current Status

The kanban view implementation is **working** with GROUP BY:
- ✅ `GroupBy` field defined in `ViewQuery`
- ✅ SQL generation implemented (generates `GROUP BY metadata->>'status'`)
- ✅ Tests passing (GROUP BY + ORDER BY + LIMIT + OFFSET)

**Problem**: The return structure hasn't been decided. Currently returns:
```go
[]map[string]interface{}  // Flat list from SQL results
```

The command `notes view kanban` receives raw SQL results and displays them as-is.

## Three Options & Analysis

### Option 1: Flat List (Current Implementation)

**What it returns:**
```go
Notes[]  // or []map[string]interface{}
```

**Example output:**
```sql
SELECT title, metadata->>'status' as status, ... FROM notes WHERE status IN (...)
GROUP BY metadata->>'status'
ORDER BY status
LIMIT 100
```

Returns:
```json
[
  {
    "id": "note-1",
    "title": "Implement auth",
    "status": "in-progress",
    "priority": 1,
    "updated_at": "2026-01-27"
  },
  {
    "id": "note-2",
    "title": "Setup DB",
    "status": "in-progress",
    "priority": 2,
    "updated_at": "2026-01-26"
  },
  {
    "id": "note-3",
    "title": "Write docs",
    "status": "done",
    "priority": 3,
    "updated_at": "2026-01-25"
  }
]
```

**Pros:**
- Simple, no transformation needed
- Works with existing display logic
- Compatible with all output formats (JSON, table, list)
- Minimal code changes

**Cons:**
- ❌ Kanban view **fundamentally needs grouped structure**
- ❌ Without grouping, display layer must infer groups (fragile)
- ❌ Can't easily render card columns in TUI
- ❌ Frontend would need to group client-side
- ❌ Loses semantic meaning of "what is grouped by what"

**When to use:** Simple, flat data displays (lists, reports)

---

### Option 2: Grouped Structure (RECOMMENDED)

**What it returns:**
```go
map[string][]map[string]interface{}  // {groupValue: notes}
// OR
map[string][]Note  // {groupValue: notes}
```

**Example output for `GROUP BY status`:**
```json
{
  "backlog": [
    {"id": "note-4", "title": "Future feature", ...}
  ],
  "in-progress": [
    {"id": "note-1", "title": "Implement auth", ...},
    {"id": "note-2", "title": "Setup DB", ...}
  ],
  "done": [
    {"id": "note-3", "title": "Write docs", ...}
  ]
}
```

**Real-world use:** This is what kanban, timeline, and dashboard views need.

**Pros:**
- ✅ Exact structure kanban view needs (columns per status)
- ✅ Semantic: clearly shows what's grouped by what
- ✅ Display layer knows how to render columns
- ✅ Supports pagination per-group
- ✅ Works for dashboards (group by priority, then count)
- ✅ Frontend doesn't duplicate grouping logic
- ✅ Extensible (can group by multiple fields later)

**Cons:**
- Requires small transformation in ViewService
- Needs different display logic (grouped vs flat)
- JSON structure changes based on GROUP BY field

**Implementation:**
1. Check if `view.Query.GroupBy != ""`
2. If grouped: Return `map[groupValue][]rows`
3. If flat: Return `[]rows` (current)
4. Command layer detects grouping and passes to appropriate display handler

---

### Option 3: Hybrid Structure with Metadata

**What it returns:**
```go
struct {
  IsGrouped    bool
  GroupField   string
  Grouping     map[string][]Note
  FlatList     []Note
  Metadata     map[string]interface{}
}
```

**Example output:**
```json
{
  "is_grouped": true,
  "group_field": "status",
  "grouping": {
    "backlog": [...],
    "in-progress": [...],
    "done": [...]
  },
  "flat_list": [...],
  "metadata": {
    "total_count": 42,
    "group_count": 3,
    "pagination": {...}
  }
}
```

**Pros:**
- Contains both grouped and flat views
- Metadata helps client understand structure
- Flexible for multiple use cases

**Cons:**
- Over-engineered for current needs
- Redundant data (both grouped + flat)
- More complex serialization
- Harder to version/maintain
- Usually for complex API responses, not ideal for CLI tool

---

## Recommendation: **Option 2 (Grouped Structure)**

### Why Option 2 is Best for OpenNotes

1. **Kanban View Purpose**
   - Kanban is fundamentally grouped (columns per status)
   - Return structure should match semantic intent
   - No client-side regrouping needed

2. **Extensibility**
   - Timeline view → group by date
   - Dashboard view → group by priority, then aggregate
   - Summary view → group by category, then COUNT
   - All benefit from same grouped return structure

3. **Display Layer Clarity**
   - Flat list: `display.RenderNotesList()`
   - Grouped: `display.RenderKanbanBoard(grouped)` or `display.RenderGroupedResults(grouped)`
   - Command layer can branch based on `GroupBy != ""`

4. **Backward Compatibility**
   - Views without `GroupBy` still return `[]Note`
   - No impact on existing flat views
   - Clean separation of concerns

5. **Type Safety**
   ```go
   // In view.go
   type ViewResults struct {
       IsGrouped bool
       GroupBy   string
       Grouped   map[string][]Note  // if IsGrouped
       Flat      []Note              // if !IsGrouped
   }
   
   // In command
   if results.IsGrouped {
       display.RenderGroupedResults(results.GroupBy, results.Grouped)
   } else {
       display.RenderNotesList(results.Flat)
   }
   ```

---

## Implementation Plan for Option 2

### Step 1: Define Return Type
```go
// internal/core/view.go
type ViewResults struct {
    IsGrouped bool
    GroupBy   string
    Grouped   map[string][]Note  // {groupValue: notes}
    Flat      []Note
}
```

### Step 2: Update ViewService
```go
// internal/services/view.go
func (vs *ViewService) ExecuteView(view *ViewDefinition, params map[string]string) (*ViewResults, error) {
    // ... existing SQL generation ...
    rows := db.Query(sqlQuery, args...)
    
    if view.Query.GroupBy != "" {
        // Group results
        grouping := make(map[string][]Note)
        for rows.Next() {
            note := scanNote(rows)
            groupValue := note.GetField(view.Query.GroupBy)  // e.g., "in-progress"
            grouping[groupValue] = append(grouping[groupValue], note)
        }
        return &ViewResults{
            IsGrouped: true,
            GroupBy: view.Query.GroupBy,
            Grouped: grouping,
        }
    }
    
    // Flat results
    var flat []Note
    for rows.Next() {
        flat = append(flat, scanNote(rows))
    }
    return &ViewResults{
        IsGrouped: false,
        Flat: flat,
    }
}
```

### Step 3: Update Command Handler
```go
// cmd/notes_view.go
results, err := vs.ExecuteView(view, userParams)
if err != nil { return err }

switch viewFormat {
case "json":
    return display.RenderViewResults(results)
case "table":
    return display.RenderViewResults(results)
case "list":
    if results.IsGrouped {
        return display.RenderKanbanBoard(results.GroupBy, results.Grouped)
    }
    return display.RenderNotesList(results.Flat)
}
```

### Step 4: Display Layer
```go
// internal/services/display.go
func (d *Display) RenderKanbanBoard(groupField string, groups map[string][]Note) error {
    // For each group value (status), display column
    for _, status := range []string{"backlog", "todo", "in-progress", "done"} {
        if notes, ok := groups[status]; ok {
            fmt.Printf("\n%s (%d)\n", status, len(notes))
            fmt.Println(strings.Repeat("─", 40))
            for _, note := range notes {
                fmt.Printf("  • %s\n", note.Title)
            }
        }
    }
}
```

---

## Impact Analysis

| Aspect | Impact |
|--------|--------|
| **Breaking Changes** | None (views without GroupBy unaffected) |
| **Code Complexity** | Low (+50 lines in ViewService, +30 in display) |
| **Test Coverage** | Need tests for grouped vs flat results |
| **JSON Output** | Structure changes only for grouped views |
| **TUI Rendering** | New kanban board renderer |
| **Performance** | Minimal (grouping happens after SQL) |

---

## Decision Record

**Chosen:** Option 2 (Grouped Structure)

**Rationale:**
- Kanban view semantically requires grouping
- Enables future views (timeline, dashboard) to reuse pattern
- Clean separation between grouped and flat views
- No breaking changes to existing code
- Type-safe and maintainable

**Next Steps:**
1. Define `ViewResults` type in `internal/core/view.go`
2. Create `ExecuteView()` method in ViewService
3. Update `notes_view.go` command to handle grouped results
4. Add kanban board renderer to display service
5. Write tests for grouped result handling

