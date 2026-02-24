---
id: b3c4d5e6
title: Incremental Dependency Replacement Strategy
type: "learning"
created_at: 2026-02-19T08:11:00+10:30
updated_at: 2026-02-19T08:11:00+10:30
status: active
tags: [migration, architecture, refactoring, strategy, dependency-management]
epic_id: f661c068
---

# Incremental Dependency Replacement Strategy

## Summary

Replacing a deeply integrated dependency (DuckDB) across 14+ files was achieved safely through a strict 6-phase ordering: struct changes → simple methods → complex methods → CLI updates → dependency removal → cleanup. The key insight: **migration order is critical**—wrong order causes cascading merge conflicts and broken intermediate states.

## Details

### The Migration Order That Worked

```
Phase 5.2.1: Struct Update (add new field, keep old)     → 0 broken tests
Phase 5.2.2: Simple Methods (getAllNotes → Index.Find)    → 1 broken test
Phase 5.2.3: Complex Methods (SearchWithConditions)       → 1 broken test
Phase 5.2.4: Verify Count() (already migrated)            → 0 broken tests
Phase 5.2.5: CLI Commands (remove --sql flag)             → 0 broken tests
Phase 5.2.6: Delete Old Code (rm db.go, 373 lines)       → 0 broken tests
Phase 5.3:   Dependency Cleanup (go.mod purge)            → Clean build
```

### Why This Order Matters

1. **Add before remove**: Add `Index` field to NoteService while keeping `DbService`—all tests continue passing
2. **Simple before complex**: Migrate `getAllNotes()` first (straightforward Find with match-all) to validate the converter pattern (`documentToNote()`), then tackle `SearchWithConditions()` which needs query AST translation
3. **Services before CLI**: Service layer fully migrated means CLI changes are just flag removal—no logic changes
4. **Code deletion last**: Only delete `db.go` after every caller is migrated. Running tests between each step confirms nothing breaks.

### The Converter Pattern

A `documentToNote()` function bridges the old domain model (Note) with the new search model (Document). This was created in Phase 5.2.2 and reused in every subsequent phase:

```go
func documentToNote(doc search.Document) Note {
    // Maps search.Document fields → Note fields
    // Handles metadata extraction from frontmatter
    // Single point of truth for the conversion
}
```

### Test Infrastructure Investment

Creating `testutil.CreateTestIndex()` in Phase 5.2.2 simplified all subsequent phases. 40+ test callers used this helper, making it trivial to provide a working Bleve index in tests. **Investment**: 30 minutes. **Savings**: Hours of repeated test setup across 5 phases.

### Breaking Changes: Clean Break vs Gradual

We chose **clean break** (no dual-support period):
- Old: `SELECT * FROM notes WHERE tag='work'`
- New: `tag:work`

This worked because:
- The tool is single-user (no enterprise migration concerns)
- SQL syntax was a power-user feature (most used CLI flags)
- New syntax is strictly better (simpler, safer)
- Clear CHANGELOG migration guide provided

### Anti-Pattern: Dual Support

Maintaining both DuckDB and Bleve simultaneously would have:
- Doubled test surface area
- Required feature flags
- Slowed development for 2+ sprints
- Introduced subtle behavioral differences between backends

## Implications

- **For dependency replacement**: Always add the new dependency alongside the old, migrate callers one-by-one, then remove the old. Never try to swap in one step.
- **For migration ordering**: Start with struct/interface changes, proceed to simple methods, then complex ones, finish with cleanup. Test between every step.
- **For test helpers**: Invest in shared test infrastructure early in the migration—it pays dividends across every subsequent phase.
- **For breaking changes**: In single-user tools, prefer clean breaks over gradual migration. The complexity of dual support rarely justifies itself.
- **Related learnings**: [archive/duckdb-removal-f661c068/learning-f661c068-duckdb-removal-epic-complete.md](archive/duckdb-removal-f661c068/learning-f661c068-duckdb-removal-epic-complete.md)
