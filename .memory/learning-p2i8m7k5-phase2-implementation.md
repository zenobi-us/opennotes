---
id: p2i8m7k5
title: Phase 2 Implementation Insights - pi-opennotes
type: "learning"
created_at: 2026-01-29T12:00:00+10:30
updated_at: 2026-01-29T12:00:00+10:30
status: complete
epic_id: 1f41631e
---

# Phase 2 Implementation Insights

## Overview

Successfully implemented the complete pi-opennotes extension following the Phase 1 design specifications.

## Key Implementation Decisions

### 1. Service-Based Architecture

**Approach**: Fat services, thin tools following SOLID principles.

**Benefits**:
- Services are independently testable with mocked CliAdapter
- Tools are ~50 lines each, just orchestrating service calls
- Easy to add new tools that compose existing services

**Example**: `opennotes_search` tool delegates to 4 different SearchService methods based on parameters (textSearch, fuzzySearch, sqlSearch, booleanSearch).

### 2. TypeBox Schema Design

**Challenge**: Google API compatibility requires StringEnum instead of Type.Union for enum values.

**Solution**: Used `StringEnum` from `@mariozechner/pi-ai` for all enum types:
```typescript
export const SortField = StringEnum(
  ["modified", "created", "title", "path"] as const,
  { description: "Field to sort by", default: "modified" }
);
```

### 3. Error Handling Strategy

**Approach**: Structured errors with installation hints.

**Key Pattern**:
```typescript
throw new OpenNotesError(
  "OpenNotes CLI not found",
  ErrorCodes.CLI_NOT_FOUND,
  { searchedPaths: process.env.PATH?.split(":") }
);
```

**Benefit**: LLM can present user-friendly errors with actionable hints.

### 4. Pagination with Budget Management

**Strategy**: 75% content + 25% metadata budget.

**Implementation**:
- PaginationService calculates byte/line limits
- fitToBudget() truncates items to fit
- Response includes nextOffset for easy continuation

### 5. SQL Search via CLI

**Challenge**: Not all CLI commands support JSON output.

**Solution**: Convert text/boolean searches to SQL queries:
```sql
SELECT file_path, metadata->>'title', metadata->'tags'
FROM read_markdown('**/*.md')
WHERE content LIKE '%search_term%'
LIMIT 50 OFFSET 0
```

**Benefit**: Consistent JSON output via `--sql` flag.

## Test Strategy

### Coverage Summary
- **72 tests passing**
- Unit tests for services (mocked CLI)
- Integration tests for tools (mocked services)

### Mock Pattern
```typescript
const mockCli = createMockCliAdapter();
mockCli.exec = mock(async () => ({
  code: 0,
  stdout: JSON.stringify([{ path: "test.md" }]),
  stderr: "",
}));
```

### Key Test Categories
1. CLI Adapter - exec, parseJsonOutput, checkInstallation
2. PaginationService - paginate, fitToBudget, exceedsBudget
3. SearchService - textSearch, sqlSearch, booleanSearch, fuzzySearch
4. Tool wrappers - metadata, parameter handling, error wrapping

## Challenges Encountered

### 1. Signal Type Compatibility

**Issue**: `AbortSignal | null` vs `AbortSignal | undefined` mismatch.

**Solution**: Use union type `AbortSignal | null` in interfaces and pass through to pi.exec.

### 2. Fuzzy Search Output Parsing

**Issue**: CLI fuzzy search returns glamour-formatted text, not JSON.

**Solution**: Parse text output with regex:
```typescript
const match = line.match(/[•\-*]\s*(\S+\.md)|^\s*(\S+\.md)/);
```

### 3. Views Tool Dual-Mode

**Issue**: Tool needs to both list views AND execute a specific view.

**Solution**: Check for `view` parameter presence:
```typescript
if (params.view) {
  // Execute mode
  return services.views.executeView(params.view, options);
} else {
  // List mode
  return services.views.listViews(notebook);
}
```

## Files Created

| File | Purpose | Lines |
|------|---------|-------|
| `src/services/types.ts` | Service interfaces | 280 |
| `src/services/cli-adapter.ts` | CLI execution | 140 |
| `src/services/pagination.service.ts` | Budget management | 120 |
| `src/services/search.service.ts` | Search operations | 260 |
| `src/services/list.service.ts` | List operations | 150 |
| `src/services/note.service.ts` | Note CRUD | 220 |
| `src/services/notebook.service.ts` | Notebook management | 130 |
| `src/services/views.service.ts` | Views operations | 230 |
| `src/tools/*.ts` | 6 tool wrappers | ~350 total |
| `src/schemas/*.ts` | TypeBox schemas | ~350 total |
| `src/utils/*.ts` | Errors, validation, output | ~500 total |
| `tests/**/*.ts` | 72 tests | ~750 total |

## Recommendations for Phase 3

1. **E2E Tests**: Add tests that use real CLI (when installed)
2. **npm Publishing**: Set up publishing workflow
3. **Documentation**: Add examples for each tool
4. **Performance**: Add CLI response caching for repeated calls
5. **Views**: Consider adding view parameter validation

## Lessons Learned

1. **Design First**: Phase 1 design made implementation smooth
2. **Mock Early**: Creating mock fixtures first enabled parallel test writing
3. **Error Hints Matter**: Installation hints make CLI tools much more user-friendly
4. **Budget Management**: 75/25 split provides good balance for pagination
5. **Service Composition**: Dependency injection via createServices() simplifies testing
