# Task: Add SQL Query Validation

**Epic**: [SQL Flag Feature](epic-2f3c4d5e-sql-flag-feature.md)

**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 1 - Core Functionality (MVP)
**Priority**: HIGH
**Complexity**: Low
**Estimated Time**: 30 minutes

## Objective

Create a `validateSQLQuery()` function that safely validates user-provided SQL queries before execution, allowing only SELECT and WITH statements while blocking dangerous keywords.

## Context

The `--sql` flag will accept arbitrary SQL from users. Before execution, queries must be validated to:
1. Allow only SELECT and WITH (CTEs)
2. Block dangerous keywords (DROP, DELETE, UPDATE, etc.)
3. Prevent SQL injection vectors

This validation is a security control, though not foolproof for all edge cases.

## Steps to Take

1. **Create validation function in NoteService**
   - File: `internal/services/note.go`
   - Function signature: `func validateSQLQuery(query string) error`

2. **Implement validation logic**
   - Trim and convert query to uppercase
   - Check query starts with SELECT or WITH
   - Return error if neither
   - Check for dangerous keywords in blocklist
   - Return error if dangerous keyword found

3. **Define dangerous keywords blocklist**
   ```
   DROP, DELETE, UPDATE, INSERT, ALTER, CREATE, 
   TRUNCATE, REPLACE, ATTACH, DETACH, PRAGMA
   ```

4. **Add helpful error messages**
   - "only SELECT queries are allowed"
   - "keyword <KEYWORD> is not allowed"
   - Format errors for user readability

5. **Add validation to NoteService test file**
   - Will be tested in next task

## Expected Outcomes

- [ ] Validation function added to `internal/services/note.go`
- [ ] Dangerous keywords blocked successfully
- [ ] SELECT queries pass validation
- [ ] WITH (CTE) queries pass validation
- [ ] Errors are clear and actionable

## Acceptance Criteria

- [x] Function is private (lowercase name)
- [x] Accepts string parameter (raw query)
- [x] Returns error or nil
- [x] SELECT queries allowed
- [x] WITH queries allowed (for CTEs)
- [x] INSERT blocked
- [x] UPDATE blocked
- [x] DELETE blocked
- [x] DROP blocked
- [x] ALTER blocked
- [x] CREATE blocked
- [x] PRAGMA blocked
- [x] Case insensitive (converts to uppercase)
- [x] Rejects non-SELECT/WITH queries
- [x] Error messages are descriptive

## Implementation Notes

### Validation Pattern
```go
func validateSQLQuery(query string) error {
    q := strings.TrimSpace(strings.ToUpper(query))
    
    // Check query type
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

### Test Cases

```go
// Should pass
- "SELECT * FROM table"
- "  SELECT  *  FROM  table  "
- "WITH cte AS (SELECT ...) SELECT * FROM cte"
- "select * from table" (lowercase)

// Should fail
- "INSERT INTO table VALUES (...)"
- "UPDATE table SET col = val"
- "DELETE FROM table"
- "DROP TABLE table"
- "ALTER TABLE table ADD COLUMN col"
- "CREATE TABLE table (...)"
- "PRAGMA table_info(table)"
```

## Dependencies

- ✅ Standard Go libraries (strings, fmt)
- ✅ NoteService already initialized

## Blockers

- None identified

## Time Estimate

- Implementation: 20 minutes
- Review: 10 minutes
- Total: 30 minutes

## Definition of Done

- [ ] Function implemented and compiles
- [ ] Passes linting
- [ ] Passes formatting
- [ ] Ready for unit testing
- [ ] Linked to unit test task

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Links**: [SQL Unit Tests](task-a1e4fa4c-sql-unit-tests.md)
