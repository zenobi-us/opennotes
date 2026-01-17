# OpenNotes - TODO

## Active Tasks

- [x] [task-a1b2c3d4-sql-flag-spec.md](.memory/task-a1b2c3d4-sql-flag-spec.md) - Create spec for `--sql` flag in search command ✅ Complete

## Awaiting Review

- [ ] [NEEDS-HUMAN] Review SQL flag specification (task-a1b2c3d4-sql-flag-spec.md)
- [ ] [NEEDS-HUMAN] Technical approval for implementation approach
- [ ] [NEEDS-HUMAN] Security review of threat model and controls

## Future Work

Once approved, implement in phases:

### Phase 1: Core Functionality (MVP)
- [ ] Add `--sql` flag to search command
- [ ] Implement query validation
- [ ] Add `NoteService.ExecuteSQL()` method
- [ ] Add `DbService.GetReadOnlyConnection()` method
- [ ] Implement basic table display
- [ ] Add error handling
- [ ] Write unit tests

### Phase 2: Enhanced Display
- [ ] Improve table formatting
- [ ] Handle long content
- [ ] Add column type detection
- [ ] Support multiple output formats

### Phase 3: Documentation
- [ ] Write user guide
- [ ] Document schema
- [ ] Provide example queries
- [ ] Add security warnings

### Phase 4: Advanced Features (Future)
- [ ] `--explain` flag
- [ ] Query templates
- [ ] SQL shell mode
- [ ] Query history

## Completed

- ✅ Research DuckDB Go client documentation
- ✅ Research DuckDB markdown extension
- ✅ Create comprehensive research document
- ✅ Write detailed specification
