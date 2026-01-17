# Task: Update CLI Help Text for SQL Flag

**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 3 - Documentation
**Priority**: MEDIUM
**Complexity**: Low
**Estimated Time**: 30 minutes

## Objective

Update the `opennotes search --help` text to document the `--sql` flag, include usage examples, and explain the security model.

## Context

Users discover features through help text. This task ensures:
- Flag is clearly documented
- Usage is shown with examples
- Security considerations are visible
- Link to detailed guide is provided

## Steps to Take

1. **Locate help text**
   - File: `cmd/search.go` or help command definition
   - Find where search command is defined

2. **Add SQL flag documentation**
   - Clear description of what it does
   - Mention it's for power users
   - Highlight that it's read-only
   - Note the 30-second timeout

3. **Add usage examples**
   - Simple SELECT example
   - Example with markdown functions
   - Example with WHERE clause
   - Explanation of what each does

4. **Add security warning**
   - Brief note about read-only access
   - Link to docs for more information
   - Explain validation is in place

5. **Add link to detailed guide**
   - Reference `docs/sql-guide.md` or similar
   - Point to schema documentation
   - Suggest examples

## Expected Outcomes

- [ ] Help text updated with --sql flag
- [ ] Examples shown clearly
- [ ] Security model explained
- [ ] Users know where to find more info
- [ ] Help is discoverable via `--help` and `help sql`

## Acceptance Criteria

- [x] `--sql` flag shows in help output
- [x] Description is clear and helpful
- [x] At least 2 examples shown
- [x] Security warning visible
- [x] Read-only access mentioned
- [x] Timeout mentioned
- [x] Link to guide included
- [x] Help text is properly formatted
- [x] Fits in terminal width (80 chars)

## Implementation Notes

### Help Text Example
```
Flags:
  --sql string    Execute custom SQL query against notes
                  Uses DuckDB markdown extension.
                  Read-only with 30s timeout.
                  See 'opennotes help sql' for details.

Examples:
  # Find notes with Python code blocks
  opennotes search --sql \
    "SELECT filepath, cb.language FROM read_markdown('**/*.md', include_filepath:=true), \
    LATERAL UNNEST(md_extract_code_blocks(content)) AS cb WHERE cb.language = 'python'"

  # Find short notes
  opennotes search --sql \
    "SELECT filepath, (md_stats(content)).word_count FROM read_markdown('**/*.md', include_filepath:=true) \
    WHERE (md_stats(content)).word_count < 300"

Security:
  ⚠️  Only SELECT queries allowed. Read-only access enforced.
  For full documentation and examples, see: docs/sql-guide.md
```

## Dependencies

- ✅ [task-710bd5bd-sql-flag-cli.md](task-710bd5bd-sql-flag-cli.md) - Flag implemented
- ✅ [task-66c1bc07-user-guide.md](task-66c1bc07-user-guide.md) - Detailed guide (reference)

## Blockers

- None identified

## Time Estimate

- Update help text: 15 minutes
- Test help output: 10 minutes
- Review: 5 minutes
- Total: 30 minutes

## Definition of Done

- [ ] `opennotes search --help` shows --sql flag
- [ ] Description is clear
- [ ] Examples shown
- [ ] Security warning visible
- [ ] Help is properly formatted
- [ ] Ready for documentation review

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Priority**: Documentation/Story 3
**Links**:
- [SQL Flag CLI](task-710bd5bd-sql-flag-cli.md)
- [User Guide](task-66c1bc07-user-guide.md)
