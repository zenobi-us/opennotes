---
id: a1b2c3d4
title: Kanban GROUP BY Return Structure Comparison
type: "research"
created_at: 2026-01-28T11:37:00+10:30
updated_at: 2026-01-28T11:37:00+10:30
status: draft
---

# Kanban GROUP BY Return Structure Comparison

## Quick Visual Comparison

### Option 1: Flat List ❌ (Not Recommended)
```
Input View:
  {
    name: "kanban",
    query: {
      groupBy: "status",
      orderBy: "priority DESC"
    }
  }

Output (flat SQL results):
[
  {id: "1", title: "Auth", status: "in-progress", priority: 1},
  {id: "2", title: "DB", status: "in-progress", priority: 2},
  {id: "3", title: "Docs", status: "done", priority: 1},
]

Display Problem:
  - How does render layer know to group by status?
  - Must infer grouping from repeated "status" values
  - Fragile: what if ordering changes?
  - No semantic meaning in result structure
```

---

### Option 2: Grouped Structure ✅ (RECOMMENDED)
```
Input View (same):
  {
    name: "kanban",
    query: {
      groupBy: "status",
      orderBy: "priority DESC"
    }
  }

Output (semantically grouped):
{
  "in-progress": [
    {id: "1", title: "Auth", status: "in-progress", priority: 1},
    {id: "2", title: "DB", status: "in-progress", priority: 2},
  ],
  "done": [
    {id: "3", title: "Docs", status: "done", priority: 1},
  ],
  "backlog": []
}

Display Benefit:
  ✅ Structure shows exactly what's grouped by what
  ✅ Render layer gets organized data
  ✅ Type-safe: know when to expect grouping
  ✅ Can iterate columns in order
```

---

### Option 3: Hybrid with Metadata ⚠️ (Over-engineered)
```
Output (redundant):
{
  "is_grouped": true,
  "group_field": "status",
  "grouping": {
    "in-progress": [...],
    "done": [...],
    "backlog": []
  },
  "flat_list": [  // Redundant copy of all notes
    {id: "1", title: "Auth", status: "in-progress", priority: 1},
    {id: "2", title: "DB", status: "in-progress", priority: 2},
    {id: "3", title: "Docs", status: "done", priority: 1},
  ],
  "metadata": {
    "total_count": 3,
    "group_count": 2,
    "max_group_size": 2
  }
}

Problems:
  ❌ Duplicates data (grouped + flat)
  ❌ Unnecessary metadata for CLI
  ❌ More complex to serialize/maintain
  ❌ Over-designed for current use case
```

---

## Real-World Display Examples

### Flat List (Option 1) - Raw SQL results
```
Title                 Status          Priority  Updated
──────────────────────────────────────────────────────
Implement auth        in-progress     1         2026-01-27
Setup DB              in-progress     2         2026-01-26
Write docs            done            1         2026-01-25
```
❌ Doesn't convey grouping visually

---

### Grouped Results (Option 2) - Kanban board
```
┌─ BACKLOG ──────────────────────────────────────┐
│                                                 │
│  (empty)                                        │
│                                                 │
└─────────────────────────────────────────────────┘

┌─ IN-PROGRESS ─────────────────────────────────┐
│                                                 │
│  • [P1] Implement auth (2026-01-27)           │
│  • [P2] Setup DB (2026-01-26)                 │
│                                                 │
└─────────────────────────────────────────────────┘

┌─ DONE ────────────────────────────────────────┐
│                                                 │
│  • [P1] Write docs (2026-01-25)               │
│                                                 │
└─────────────────────────────────────────────────┘
```
✅ Visually clear grouping, kanban-style columns

---

## Type Definition (Option 2)

```go
// ViewResults represents the result of executing a view
type ViewResults struct {
    IsGrouped bool
    GroupBy   string                        // "status", "priority", etc.
    Grouped   map[string][]Note            // {groupValue: notes}
    Flat      []Note                        // for non-grouped views
}

// Usage in command:
results, err := viewService.ExecuteView(view, params)
if results.IsGrouped {
    display.RenderKanbanBoard(results.GroupBy, results.Grouped)
} else {
    display.RenderNotesList(results.Flat)
}
```

---

## Implementation Cost Comparison

| Aspect | Option 1 | Option 2 | Option 3 |
|--------|----------|----------|----------|
| **Code Changes** | 0 lines | ~80 lines | ~150 lines |
| **Tests Needed** | 2 (existing) | 6 new | 8 new |
| **Display Logic** | Simple | Requires kanban renderer | Complex metadata handling |
| **Backward Compat** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Data Redundancy** | None | None | ❌ High (grouped + flat) |
| **Extensibility** | ❌ No | ✅ Yes (timeline, dashboard) | ✅ Yes but overly complex |

---

## Kanban View Use Cases

### Current: GROUP BY status
```
kanban {
  query: {
    groupBy: "status"  // ← Returns grouped structure
    orderBy: "priority DESC"
  }
}

Result:
{
  "backlog": [...],
  "in-progress": [...],
  "done": [...]
}

Render: Columns per status ✅
```

### Future: GROUP BY priority, then aggregate
```
priority_dashboard {
  query: {
    groupBy: "priority",
    selectColumns: ["priority"],
    aggregateColumns: {
      "count": "COUNT(*)",
      "avg_priority": "AVG(CAST(metadata->>'priority' AS INTEGER))"
    }
  }
}

Result (Option 2):
{
  "1": [{count: 5, avg_priority: 1.0}],
  "2": [{count: 3, avg_priority: 2.0}],
  "3": [{count: 1, avg_priority: 3.0}]
}

Render: Grouped summary table ✅
```

### Future: GROUP BY date, show timeline
```
timeline {
  query: {
    groupBy: "DATE(metadata->>'created_at')",
    orderBy: "DATE(metadata->>'created_at') DESC"
  }
}

Result (Option 2):
{
  "2026-01-27": [{id: "1", ...}, {id: "2", ...}],
  "2026-01-26": [{id: "3", ...}],
  "2026-01-25": [{id: "4", ...}, {id: "5", ...}]
}

Render: Timeline with grouped entries ✅
```

---

## Final Recommendation

**Choose: Option 2 (Grouped Structure)**

### Why:
1. **Semantically Correct** - Structure matches intent (GROUP BY)
2. **Future-Proof** - Enables timeline, dashboard, analytics views
3. **Type-Safe** - Command layer knows exactly what to expect
4. **Clean Code** - No redundant data, clear separation
5. **Backward Compatible** - Views without GroupBy stay flat
6. **Reasonable Effort** - ~80 lines of code, ~6 test cases

### Implementation Steps:
1. ✅ Define `ViewResults` in `internal/core/view.go`
2. ✅ Create `ExecuteView()` method in ViewService  
3. ✅ Update `notes view` command to handle both paths
4. ✅ Add kanban board renderer to display service
5. ✅ Write tests for grouped/flat branches
6. ✅ Update JSON output format

**Status**: Ready to implement in next session
