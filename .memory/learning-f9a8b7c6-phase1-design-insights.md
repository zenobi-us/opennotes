---
id: f9a8b7c6
title: Phase 1 Design Insights - pi-opennotes Extension
type: "learning"
created_at: 2026-01-29T10:00:00+10:30
updated_at: 2026-01-29T10:00:00+10:30
status: completed
epic_id: 1f41631e
---

# Phase 1 Design Insights - pi-opennotes Extension

## Overview

This document captures key decisions, trade-offs, and insights from designing the pi-opennotes extension during Phase 1 (Research & Design).

---

## Key Architectural Decisions

### 1. Service-Based Architecture (Fat Services, Thin Tools)

**Decision**: All business logic lives in discrete services; tools are thin wrappers.

**Rationale**:
- Follows SOLID principles (especially Single Responsibility)
- Services are testable in isolation
- Tools become simple orchestration (~30-50 lines each)
- Enables code reuse across tools

**Trade-off**: More files/modules, but worth it for maintainability.

**Implementation**:
```
Tool Layer (thin)  →  Service Layer (fat)  →  CliAdapter  →  CLI Binary
     │                      │                      │
  validate              execute               shell out
  format               paginate               parse output
```

### 2. CLI Adapter as Central Abstraction

**Decision**: Single `CliAdapter` interface for all CLI interactions.

**Rationale**:
- Easy to mock for testing
- Consistent error handling
- Single place for timeout/signal handling
- Enables future optimization (caching, batching)

**Interface**:
```typescript
interface ICliAdapter {
  exec(command, args, options): Promise<CliResult>
  checkInstallation(): Promise<InstallationStatus>
  parseJsonOutput<T>(stdout): T
  buildNotebookArgs(notebook?): string[]
}
```

### 3. Pagination Strategy: 75% Budget + Metadata

**Decision**: When results exceed output limits, return first N items fitting 75% of budget plus pagination metadata.

**Rationale**:
- LLM needs guidance to fetch more
- 75% leaves room for metadata + response formatting
- Consistent across all tools
- `nextOffset` makes continuation easy

**Format**:
```json
{
  "pagination": {
    "total": 127,
    "returned": 50,
    "page": 1,
    "pageSize": 50,
    "hasMore": true,
    "nextOffset": 50
  }
}
```

### 4. Views Tool Dual-Mode Design

**Decision**: Single tool handles both listing views AND executing views based on presence of `view` parameter.

**Rationale**:
- Reduces tool count (1 vs 2)
- Natural UX: "show me views" vs "run this view"
- Parameter presence determines mode

**Behavior**:
- No `view` param → List all available views
- With `view` param → Execute named view

### 5. Error Handling with Installation Hints

**Decision**: Every error includes actionable hints; CLI_NOT_FOUND includes full installation guide.

**Rationale**:
- LLMs can relay helpful information to users
- `recoverable` flag helps LLM decide when to ask user
- Installation guide is copy-pasteable

**Structure**:
```typescript
{
  error: true,
  message: "Human-readable message",
  code: "OPENNOTES_ERROR_CODE",
  hint: "How to fix this...",
  recoverable: boolean
}
```

---

## Technical Trade-offs

### JSON Output Limitations

**Problem**: Not all CLI commands support `--format json`.

**Solution**: Convert to SQL queries for consistent JSON output.
- `notes list` → `notes search --sql "SELECT ... FROM read_markdown(...)"`
- `notes search <text>` → `notes search --sql "... WHERE content LIKE ..."`

**Trade-off**: More complex implementation, but consistent interface.

### Fuzzy Search Output

**Problem**: Fuzzy search returns glamour-formatted text, not JSON.

**Solution**: Parse text output to extract note paths.

**Trade-off**: Less metadata available, but still functional. May improve in CLI later.

### TypeBox StringEnum Requirement

**Problem**: `Type.Union` with string literals doesn't work with Google's Gemini API.

**Solution**: Use `StringEnum` from `@mariozechner/pi-ai`.

```typescript
// Instead of:
Type.Union([Type.Literal("list"), Type.Literal("table")])

// Use:
StringEnum(["list", "table"] as const)
```

---

## Testing Insights

### Test Pyramid Ratios

**Chosen**: 70% Unit / 25% Integration / 5% E2E

**Rationale**:
- Services contain most logic → most unit tests
- Tools are thin → less unit testing, more integration
- E2E requires real CLI → expensive, few tests

### Mock Strategy

**Unit Tests**: Mock `CliAdapter` entirely
**Integration Tests**: Mock services, test tool orchestration
**E2E Tests**: Real CLI, real notebook, test full flow

**Key Insight**: E2E tests should be skippable (`it.skipIf(!cliInstalled)`) to not block CI.

---

## What Worked Well

1. **Starting with interfaces**: Defined all service interfaces before implementation details
2. **Documenting CLI first**: Complete CLI reference prevents surprises later
3. **Error codes upfront**: Forces thinking about failure modes early
4. **TypeBox schemas**: Serve as both validation AND documentation

---

## What to Watch For in Phase 2

1. **CLI output parsing**: Text parsing for non-JSON outputs may be fragile
2. **Timeout handling**: 30s default may be too short for large notebooks
3. **Pagination edge cases**: Empty results, exact page boundaries
4. **Error message clarity**: Hints should be tested with actual users

---

## Recommendations for Implementation

### Start Order
1. `CliAdapter` + installation check
2. `PaginationService` (shared utility)
3. `SearchService` (most complex, most used)
4. Other services in parallel
5. Tools (once services stable)

### Quality Gates
- [ ] Each service must have 85%+ test coverage before moving on
- [ ] Integration tests must pass before merging tool PRs
- [ ] E2E tests run on CI (but can be skipped if CLI unavailable)

### Documentation
- README should include troubleshooting section
- Each tool needs at least one example in description
- Error hints should link to full documentation

---

## Summary

Phase 1 established a solid foundation for pi-opennotes:
- Service-based architecture enables testability and maintainability
- CLI adapter abstraction allows consistent error handling
- Pagination strategy respects LLM context limits
- Error handling guides users to solutions
- Test strategy balances coverage with practicality

**Ready for Phase 2 implementation after human review.**
