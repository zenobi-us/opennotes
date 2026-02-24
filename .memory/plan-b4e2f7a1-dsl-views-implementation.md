---
id: b4e2f7a1-plan
title: "DSL-Based Views Implementation Plan"
type: "plan"
created_at: 2026-02-18T20:58:00+10:30
updated_at: 2026-02-18T20:58:00+10:30
status: ready
epic_id: f661c068
research_id: b4e2f7a1
assigned_to: unassigned
---

# DSL-Based Views Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace SQL-based view system with DSL-based views using pipe syntax (`filter | directives`)

**Architecture:** Views become DSL query strings stored as `ViewDefinition.Query`. Filter portion parsed by existing Participle parser, directives parsed by simple key:value splitter. Results returned via existing `Index.Find()` with `FindOpts`.

**Tech Stack:** Go, Participle parser, Bleve search index, Cobra CLI

---

## Prerequisites

Before starting, ensure:
- Working directory: `/mnt/Store/Projects/Mine/Github/opennotes`
- Branch: `feat/remove-duckdb-migrate-to-afero-chromedb-with-bleve-search` (or create feature branch from it)
- Tests pass: `mise run test`

---

## Task 1: Add `has:` and `missing:` Keywords to DSL Grammar

**Files:**
- Modify: `internal/search/parser/grammar.go`
- Modify: `internal/search/parser/parser.go`
- Create: `internal/search/parser/existence_test.go`

### Step 1.1: Write failing tests for existence keywords

```go
// internal/search/parser/existence_test.go
package parser

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/zenobi-us/opennotes/internal/search"
)

func TestParser_HasKeyword(t *testing.T) {
    p := New()
    
    t.Run("has:tag matches notes with any tag", func(t *testing.T) {
        result, err := p.Parse("has:tag")
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.Equal(t, search.OpExists, result.Op)
        assert.Equal(t, "tag", result.Field)
    })
    
    t.Run("has:status matches notes with status field", func(t *testing.T) {
        result, err := p.Parse("has:status")
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.Equal(t, search.OpExists, result.Op)
        assert.Equal(t, "status", result.Field)
    })
}

func TestParser_MissingKeyword(t *testing.T) {
    p := New()
    
    t.Run("missing:tag matches notes without tags", func(t *testing.T) {
        result, err := p.Parse("missing:tag")
        require.NoError(t, err)
        require.NotNil(t, result)
        assert.Equal(t, search.OpNotExists, result.Op)
        assert.Equal(t, "tag", result.Field)
    })
}

func TestParser_ExistenceWithOtherTerms(t *testing.T) {
    p := New()
    
    t.Run("combines existence with field match", func(t *testing.T) {
        result, err := p.Parse("has:tag status:todo")
        require.NoError(t, err)
        require.NotNil(t, result)
        // Should be AND of two conditions
        assert.Equal(t, search.OpAnd, result.Op)
    })
}
```

### Step 1.2: Run tests to verify they fail

```bash
mise run test -- -run "TestParser_HasKeyword|TestParser_MissingKeyword|TestParser_ExistenceWithOtherTerms"
```

Expected: FAIL with "OpExists not defined" or similar

### Step 1.3: Add new operators to search types

```go
// internal/search/types.go - add to Op constants
const (
    // ... existing operators ...
    OpExists    Op = "exists"    // has:field
    OpNotExists Op = "notexists" // missing:field
)
```

### Step 1.4: Update Participle grammar

```go
// internal/search/parser/grammar.go - add existence expression
type ExistenceExpr struct {
    Keyword string `@("has" | "missing")`
    Sep     string `":"`
    Field   string `@Ident`
}

// Update Expression to include ExistenceExpr
type Expression struct {
    Existence  *ExistenceExpr  `@@`
    Field      *FieldExpr      `| @@`
    Comparison *ComparisonExpr `| @@`
    Negation   *NegationExpr   `| @@`
    Term       *string         `| @(Ident | String)`
}
```

### Step 1.5: Update parser convert function

```go
// internal/search/parser/parser.go - in convert()
func (p *Parser) convertExpr(expr *Expression) *search.Query {
    if expr.Existence != nil {
        op := search.OpExists
        if expr.Existence.Keyword == "missing" {
            op = search.OpNotExists
        }
        return &search.Query{
            Op:    op,
            Field: expr.Existence.Field,
        }
    }
    // ... rest of existing conversion
}
```

### Step 1.6: Run tests to verify they pass

```bash
mise run test -- -run "TestParser_HasKeyword|TestParser_MissingKeyword|TestParser_ExistenceWithOtherTerms"
```

Expected: PASS

### Step 1.7: Commit

```bash
git add internal/search/parser/ internal/search/types.go
git commit -m "feat(parser): add has: and missing: existence keywords

Add support for existence checks in DSL:
- has:field - matches documents where field exists
- missing:field - matches documents where field is absent

Required for DSL-based views (untagged, kanban)."
```

---

## Task 2: Implement Quote-Aware Pipe Splitting

**Files:**
- Create: `internal/services/view_query.go`
- Create: `internal/services/view_query_test.go`

### Step 2.1: Write failing tests for pipe splitting

```go
// internal/services/view_query_test.go
package services

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestSplitViewQuery(t *testing.T) {
    tests := []struct {
        name       string
        input      string
        wantFilter string
        wantDirs   string
    }{
        {
            name:       "simple pipe split",
            input:      "tag:work | sort:modified:desc",
            wantFilter: "tag:work",
            wantDirs:   "sort:modified:desc",
        },
        {
            name:       "no pipe returns filter only",
            input:      "tag:work status:todo",
            wantFilter: "tag:work status:todo",
            wantDirs:   "",
        },
        {
            name:       "empty filter with directives",
            input:      "| sort:modified:desc limit:20",
            wantFilter: "",
            wantDirs:   "sort:modified:desc limit:20",
        },
        {
            name:       "pipe inside quoted string is not split",
            input:      `title:"A | B" tag:work`,
            wantFilter: `title:"A | B" tag:work`,
            wantDirs:   "",
        },
        {
            name:       "pipe after quoted string splits correctly",
            input:      `title:"A | B" | sort:title:asc`,
            wantFilter: `title:"A | B"`,
            wantDirs:   "sort:title:asc",
        },
        {
            name:       "trims whitespace",
            input:      "  tag:work  |  sort:modified:desc  ",
            wantFilter: "tag:work",
            wantDirs:   "sort:modified:desc",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            filter, dirs := SplitViewQuery(tt.input)
            assert.Equal(t, tt.wantFilter, filter)
            assert.Equal(t, tt.wantDirs, dirs)
        })
    }
}
```

### Step 2.2: Run tests to verify they fail

```bash
mise run test -- -run "TestSplitViewQuery"
```

Expected: FAIL with "SplitViewQuery not defined"

### Step 2.3: Implement SplitViewQuery

```go
// internal/services/view_query.go
package services

import "strings"

// SplitViewQuery splits a view query string on the first unquoted pipe character.
// Returns (filter, directives) where filter is the DSL query portion
// and directives is the presentation options portion.
// Pipe characters inside quoted strings are preserved.
func SplitViewQuery(query string) (filter, directives string) {
    inQuote := false
    for i, ch := range query {
        switch ch {
        case '"':
            inQuote = !inQuote
        case '|':
            if !inQuote {
                return strings.TrimSpace(query[:i]), strings.TrimSpace(query[i+1:])
            }
        }
    }
    return strings.TrimSpace(query), ""
}
```

### Step 2.4: Run tests to verify they pass

```bash
mise run test -- -run "TestSplitViewQuery"
```

Expected: PASS

### Step 2.5: Commit

```bash
git add internal/services/view_query.go internal/services/view_query_test.go
git commit -m "feat(views): add quote-aware pipe splitting for view queries

SplitViewQuery separates filter DSL from presentation directives.
Pipe characters inside quoted strings are preserved."
```

---

## Task 3: Implement Directives Parser

**Files:**
- Modify: `internal/services/view_query.go`
- Modify: `internal/services/view_query_test.go`

### Step 3.1: Write failing tests for directive parsing

```go
// Add to internal/services/view_query_test.go

func TestParseDirectives(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        wantSort  string
        wantDir   string
        wantLimit int
        wantGroup string
        wantErr   bool
    }{
        {
            name:      "sort directive",
            input:     "sort:modified:desc",
            wantSort:  "modified",
            wantDir:   "desc",
            wantLimit: 0,
        },
        {
            name:      "sort with default direction",
            input:     "sort:title",
            wantSort:  "title",
            wantDir:   "asc",
            wantLimit: 0,
        },
        {
            name:      "limit directive",
            input:     "limit:20",
            wantLimit: 20,
        },
        {
            name:      "group directive",
            input:     "group:status",
            wantGroup: "status",
        },
        {
            name:      "multiple directives",
            input:     "sort:modified:desc limit:50 group:status",
            wantSort:  "modified",
            wantDir:   "desc",
            wantLimit: 50,
            wantGroup: "status",
        },
        {
            name:    "unknown directive errors",
            input:   "foo:bar",
            wantErr: true,
        },
        {
            name:      "case insensitive",
            input:     "Sort:Modified:DESC Limit:10",
            wantSort:  "modified",
            wantDir:   "desc",
            wantLimit: 10,
        },
        {
            name:      "last directive wins on conflict",
            input:     "limit:10 limit:20",
            wantLimit: 20,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            d, err := ParseDirectives(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.wantSort, d.SortField)
            assert.Equal(t, tt.wantDir, d.SortDirection)
            assert.Equal(t, tt.wantLimit, d.Limit)
            assert.Equal(t, tt.wantGroup, d.GroupBy)
        })
    }
}
```

### Step 3.2: Run tests to verify they fail

```bash
mise run test -- -run "TestParseDirectives"
```

Expected: FAIL

### Step 3.3: Implement ParseDirectives

```go
// Add to internal/services/view_query.go

import (
    "fmt"
    "strconv"
    "strings"
)

// ViewDirectives holds parsed presentation options from the directive portion of a view query.
type ViewDirectives struct {
    SortField     string
    SortDirection string // "asc" or "desc"
    Limit         int
    Offset        int
    GroupBy       string
}

// ParseDirectives parses the directive portion of a view query.
// Valid directives: sort:<field>:<asc|desc>, limit:<n>, offset:<n>, group:<field>
// Directives are case-insensitive. Last directive wins on conflict.
func ParseDirectives(input string) (*ViewDirectives, error) {
    d := &ViewDirectives{
        SortDirection: "asc", // default
    }
    
    if strings.TrimSpace(input) == "" {
        return d, nil
    }
    
    parts := strings.Fields(input)
    for _, part := range parts {
        colonIdx := strings.Index(part, ":")
        if colonIdx == -1 {
            return nil, fmt.Errorf("invalid directive %q: missing colon", part)
        }
        
        key := strings.ToLower(part[:colonIdx])
        value := part[colonIdx+1:]
        
        switch key {
        case "sort":
            // sort:field or sort:field:dir
            sortParts := strings.Split(value, ":")
            d.SortField = strings.ToLower(sortParts[0])
            if len(sortParts) > 1 {
                d.SortDirection = strings.ToLower(sortParts[1])
            }
        case "limit":
            n, err := strconv.Atoi(value)
            if err != nil {
                return nil, fmt.Errorf("invalid limit %q: %w", value, err)
            }
            d.Limit = n
        case "offset":
            n, err := strconv.Atoi(value)
            if err != nil {
                return nil, fmt.Errorf("invalid offset %q: %w", value, err)
            }
            d.Offset = n
        case "group":
            d.GroupBy = strings.ToLower(value)
        default:
            return nil, fmt.Errorf("unknown directive %q. Valid: sort, limit, offset, group", key)
        }
    }
    
    return d, nil
}
```

### Step 3.4: Run tests to verify they pass

```bash
mise run test -- -run "TestParseDirectives"
```

Expected: PASS

### Step 3.5: Commit

```bash
git add internal/services/view_query.go internal/services/view_query_test.go
git commit -m "feat(views): add directive parser for view queries

ParseDirectives handles sort, limit, offset, group directives.
Case-insensitive, last-wins semantics on conflicts."
```

---

## Task 4: Update ViewDefinition Type

**Files:**
- Modify: `internal/core/view.go`
- Modify: `internal/core/view_test.go` (if exists)

### Step 4.1: Create new ViewDefinition type

The current `ViewDefinition` uses `ViewQuery` struct. Replace with string-based `Query`.

```go
// internal/core/view.go - NEW simplified type

package core

// ViewDefinition defines a named view with a DSL query string.
type ViewDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  []ViewParameter `json:"parameters,omitempty"`
    Query       string          `json:"query"`          // "filter DSL | directives"
    Type        string          `json:"type,omitempty"` // "query" (default) or "special"
}

// ViewParameter defines a parameter that can be substituted into a view query.
type ViewParameter struct {
    Name        string `json:"name"`
    Type        string `json:"type"` // "string", "number", "date"
    Required    bool   `json:"required,omitempty"`
    Default     string `json:"default,omitempty"`
    Description string `json:"description,omitempty"`
}

// IsSpecialView returns true if this view requires special execution (not DSL-based).
func (v *ViewDefinition) IsSpecialView() bool {
    return v.Type == "special"
}
```

### Step 4.2: Write test for new type

```go
// internal/core/view_test.go
package core

import (
    "encoding/json"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestViewDefinition_JSON(t *testing.T) {
    t.Run("marshals query-based view", func(t *testing.T) {
        v := ViewDefinition{
            Name:        "today",
            Description: "Notes modified today",
            Query:       "modified:>=today | sort:modified:desc",
        }
        
        data, err := json.Marshal(v)
        require.NoError(t, err)
        assert.Contains(t, string(data), `"query":"modified:>=today | sort:modified:desc"`)
    })
    
    t.Run("marshals special view", func(t *testing.T) {
        v := ViewDefinition{
            Name:        "orphans",
            Description: "Notes with no incoming links",
            Type:        "special",
        }
        
        data, err := json.Marshal(v)
        require.NoError(t, err)
        assert.Contains(t, string(data), `"type":"special"`)
    })
}

func TestViewDefinition_IsSpecialView(t *testing.T) {
    t.Run("returns true for special type", func(t *testing.T) {
        v := ViewDefinition{Type: "special"}
        assert.True(t, v.IsSpecialView())
    })
    
    t.Run("returns false for query type", func(t *testing.T) {
        v := ViewDefinition{Type: "query"}
        assert.False(t, v.IsSpecialView())
    })
    
    t.Run("returns false for empty type (default)", func(t *testing.T) {
        v := ViewDefinition{}
        assert.False(t, v.IsSpecialView())
    })
}
```

### Step 4.3: Run tests

```bash
mise run test -- -run "TestViewDefinition"
```

Expected: PASS (or update existing tests if type change breaks them)

### Step 4.4: Commit

```bash
git add internal/core/view.go internal/core/view_test.go
git commit -m "refactor(core): simplify ViewDefinition to use string Query

Replace complex ViewQuery struct with simple string field.
View queries now use pipe syntax: 'filter DSL | directives'.
Add IsSpecialView() helper for dispatch routing."
```

---

## Task 5: Rewrite Builtin Views with DSL

**Files:**
- Modify: `internal/services/view.go` (lines 48-157)

### Step 5.1: Write tests for new builtin views

```go
// Add to internal/services/view_test.go

func TestBuiltinViews_DSLFormat(t *testing.T) {
    s := NewViewService(nil, nil)
    
    tests := []struct {
        name           string
        viewName       string
        expectQuery    string
        expectType     string
    }{
        {
            name:        "today view uses DSL",
            viewName:    "today",
            expectQuery: "modified:>=today | sort:modified:desc",
        },
        {
            name:        "recent view uses DSL",
            viewName:    "recent",
            expectQuery: "| sort:modified:desc limit:20",
        },
        {
            name:        "kanban view uses DSL",
            viewName:    "kanban",
            expectQuery: "has:status | group:status sort:title:asc",
        },
        {
            name:        "untagged view uses DSL",
            viewName:    "untagged",
            expectQuery: "missing:tag | sort:created:desc",
        },
        {
            name:       "orphans is special view",
            viewName:   "orphans",
            expectType: "special",
        },
        {
            name:       "broken-links is special view",
            viewName:   "broken-links",
            expectType: "special",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            view := s.builtinViews[tt.viewName]
            require.NotNil(t, view, "builtin view %s not found", tt.viewName)
            
            if tt.expectQuery != "" {
                assert.Equal(t, tt.expectQuery, view.Query)
            }
            if tt.expectType != "" {
                assert.Equal(t, tt.expectType, view.Type)
            }
        })
    }
}
```

### Step 5.2: Run tests to verify they fail

```bash
mise run test -- -run "TestBuiltinViews_DSLFormat"
```

Expected: FAIL (old SQL format doesn't match)

### Step 5.3: Rewrite initializeBuiltinViews

```go
// internal/services/view.go - replace initializeBuiltinViews()

func (s *ViewService) initializeBuiltinViews() {
    s.builtinViews = map[string]*core.ViewDefinition{
        "today": {
            Name:        "today",
            Description: "Notes created or updated today",
            Query:       "modified:>=today | sort:modified:desc",
        },
        "recent": {
            Name:        "recent",
            Description: "Recently modified notes (last 20)",
            Query:       "| sort:modified:desc limit:20",
        },
        "kanban": {
            Name:        "kanban",
            Description: "Notes grouped by status",
            Query:       "has:status | group:status sort:title:asc",
        },
        "untagged": {
            Name:        "untagged",
            Description: "Notes without any tags",
            Query:       "missing:tag | sort:created:desc",
        },
        "orphans": {
            Name:        "orphans",
            Description: "Notes with no incoming links",
            Type:        "special",
        },
        "broken-links": {
            Name:        "broken-links",
            Description: "Notes with broken references",
            Type:        "special",
        },
    }
}
```

### Step 5.4: Run tests to verify they pass

```bash
mise run test -- -run "TestBuiltinViews_DSLFormat"
```

Expected: PASS

### Step 5.5: Commit

```bash
git add internal/services/view.go internal/services/view_test.go
git commit -m "feat(views): rewrite builtin views to use DSL pipe syntax

Builtin views now use 'filter | directives' format:
- today: modified:>=today | sort:modified:desc
- recent: | sort:modified:desc limit:20
- kanban: has:status | group:status sort:title:asc
- untagged: missing:tag | sort:created:desc
- orphans/broken-links: remain as special views"
```

---

## Task 6: Implement View Query Execution

**Files:**
- Modify: `internal/services/view.go`
- Create: `internal/services/view_executor.go`
- Create: `internal/services/view_executor_test.go`

### Step 6.1: Write failing tests for ExecuteView

```go
// internal/services/view_executor_test.go
package services

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/zenobi-us/opennotes/internal/core"
)

func TestViewService_ExecuteView(t *testing.T) {
    // Setup: create test notebook with notes
    ctx := context.Background()
    notebook := createTestNotebook(t) // helper that creates temp notebook
    defer notebook.Cleanup()
    
    // Add test notes
    notebook.AddNote("note1.md", "---\ntags: [work]\nstatus: todo\n---\n# Note 1")
    notebook.AddNote("note2.md", "---\ntags: [personal]\nstatus: done\n---\n# Note 2")
    notebook.AddNote("note3.md", "---\nstatus: todo\n---\n# Note 3 (no tags)")
    
    // Create service
    vs := NewViewService(notebook.NoteService, notebook.SearchService)
    
    t.Run("executes simple filter view", func(t *testing.T) {
        view := &core.ViewDefinition{
            Name:  "work",
            Query: "tag:work",
        }
        
        results, err := vs.ExecuteView(ctx, view, nil)
        require.NoError(t, err)
        assert.Len(t, results.Notes, 1)
        assert.Equal(t, "note1.md", results.Notes[0].Path)
    })
    
    t.Run("executes view with limit", func(t *testing.T) {
        view := &core.ViewDefinition{
            Name:  "limited",
            Query: "| limit:1",
        }
        
        results, err := vs.ExecuteView(ctx, view, nil)
        require.NoError(t, err)
        assert.Len(t, results.Notes, 1)
    })
    
    t.Run("executes view with grouping", func(t *testing.T) {
        view := &core.ViewDefinition{
            Name:  "by-status",
            Query: "has:status | group:status",
        }
        
        results, err := vs.ExecuteView(ctx, view, nil)
        require.NoError(t, err)
        assert.NotNil(t, results.Groups)
        assert.Contains(t, results.Groups, "todo")
        assert.Contains(t, results.Groups, "done")
    })
    
    t.Run("delegates special views to special executor", func(t *testing.T) {
        view := &core.ViewDefinition{
            Name: "orphans",
            Type: "special",
        }
        
        results, err := vs.ExecuteView(ctx, view, nil)
        require.NoError(t, err)
        // Special view should work (may return empty for test data)
        assert.NotNil(t, results)
    })
}
```

### Step 6.2: Run tests to verify they fail

```bash
mise run test -- -run "TestViewService_ExecuteView"
```

Expected: FAIL

### Step 6.3: Implement ExecuteView

```go
// internal/services/view_executor.go
package services

import (
    "context"
    "fmt"

    "github.com/zenobi-us/opennotes/internal/core"
    "github.com/zenobi-us/opennotes/internal/search"
    "github.com/zenobi-us/opennotes/internal/search/parser"
)

// ViewResults holds the results of executing a view.
type ViewResults struct {
    Notes  []core.Note              // Flat list (when not grouped)
    Groups map[string][]core.Note   // Grouped results (when group directive used)
}

// ExecuteView executes a view definition and returns results.
func (s *ViewService) ExecuteView(ctx context.Context, view *core.ViewDefinition, params map[string]string) (*ViewResults, error) {
    // Handle special views
    if view.IsSpecialView() {
        return s.executeSpecialView(ctx, view)
    }
    
    // Resolve template variables ({{today}}, etc.)
    resolvedQuery, err := s.ResolveTemplateVariables(view.Query, params)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve template variables: %w", err)
    }
    
    // Split query into filter and directives
    filterPart, directivesPart := SplitViewQuery(resolvedQuery)
    
    // Parse directives
    directives, err := ParseDirectives(directivesPart)
    if err != nil {
        return nil, fmt.Errorf("failed to parse directives: %w", err)
    }
    
    // Build FindOpts
    opts := search.FindOpts{
        Limit:  directives.Limit,
        Offset: directives.Offset,
    }
    
    // Parse filter DSL if present
    if filterPart != "" {
        p := parser.New()
        query, err := p.Parse(filterPart)
        if err != nil {
            return nil, fmt.Errorf("failed to parse filter: %w", err)
        }
        opts.Query = query
    }
    
    // Set sort
    if directives.SortField != "" {
        opts.Sort = s.directiveToSortSpec(directives.SortField, directives.SortDirection)
    }
    
    // Execute search
    notes, err := s.noteService.SearchWithOptions(ctx, opts)
    if err != nil {
        return nil, fmt.Errorf("search failed: %w", err)
    }
    
    // Handle grouping
    if directives.GroupBy != "" {
        groups := s.groupNotes(notes, directives.GroupBy)
        return &ViewResults{Groups: groups}, nil
    }
    
    return &ViewResults{Notes: notes}, nil
}

// directiveToSortSpec converts directive strings to search.SortSpec
func (s *ViewService) directiveToSortSpec(field, direction string) search.SortSpec {
    desc := direction == "desc"
    switch field {
    case "modified":
        return search.SortSpec{Field: search.SortByModified, Descending: desc}
    case "created":
        return search.SortSpec{Field: search.SortByCreated, Descending: desc}
    case "title":
        return search.SortSpec{Field: search.SortByTitle, Descending: desc}
    default:
        return search.SortSpec{Field: search.SortByRelevance, Descending: desc}
    }
}

// groupNotes groups notes by a field value
func (s *ViewService) groupNotes(notes []core.Note, field string) map[string][]core.Note {
    groups := make(map[string][]core.Note)
    for _, note := range notes {
        key := s.getNoteFieldValue(note, field)
        if key == "" {
            key = "(none)"
        }
        groups[key] = append(groups[key], note)
    }
    return groups
}

// getNoteFieldValue extracts a field value from a note for grouping
func (s *ViewService) getNoteFieldValue(note core.Note, field string) string {
    switch field {
    case "status":
        return note.Metadata.Status
    // Add other fields as needed
    default:
        return ""
    }
}

// executeSpecialView dispatches to special view executor
func (s *ViewService) executeSpecialView(ctx context.Context, view *core.ViewDefinition) (*ViewResults, error) {
    switch view.Name {
    case "orphans":
        notes, err := s.specialExecutor.ExecuteOrphansView(ctx)
        if err != nil {
            return nil, err
        }
        return &ViewResults{Notes: notes}, nil
    case "broken-links":
        notes, err := s.specialExecutor.ExecuteBrokenLinksView(ctx)
        if err != nil {
            return nil, err
        }
        return &ViewResults{Notes: notes}, nil
    default:
        return nil, fmt.Errorf("unknown special view: %s", view.Name)
    }
}
```

### Step 6.4: Run tests to verify they pass

```bash
mise run test -- -run "TestViewService_ExecuteView"
```

Expected: PASS

### Step 6.5: Commit

```bash
git add internal/services/view_executor.go internal/services/view_executor_test.go
git commit -m "feat(views): implement DSL-based view execution

ExecuteView parses query, builds FindOpts, executes search.
Supports filtering, sorting, limiting, grouping.
Special views delegate to existing special executor."
```

---

## Task 7: Remove Dead SQL Code (Phase 5 Cleanup)

**Files:**
- Modify: `internal/services/view.go`
- Modify: `internal/services/view_test.go`

### Step 7.1: Remove SQL tests first (safe)

Remove these test functions from `view_test.go`:
- `TestViewService_FormatQueryValue_*` (lines ~466-504)
- `TestViewService_GenerateSQL_*` (lines ~658-1368)
- `TestViewService_ValidateViewDefinition_*` that test SQL fields (lines ~184-290)

```bash
# Use editor to remove test functions, then run:
mise run test
```

Expected: Tests should pass (removed tests were for dead code)

### Step 7.2: Commit test removal

```bash
git add internal/services/view_test.go
git commit -m "test(views): remove SQL-specific tests

Remove tests for SQL generation and validation:
- FormatQueryValue tests
- GenerateSQL tests  
- SQL field validation tests

These tested dead code paths after Bleve migration."
```

### Step 7.3: Remove SQL implementation (leaf functions first)

Remove in this order from `view.go`:

1. `escapeSQL()` 
2. `FormatQueryValue()`
3. `validateOperator()`
4. `validateField()`
5. `validateAggregateFunction()`
6. `validateHavingCondition()`
7. `validateViewCondition()`
8. `transformSQLGroupedResults()`
9. `GenerateSQL()`

After each removal, run `mise run test` to ensure nothing breaks.

### Step 7.4: Commit SQL implementation removal

```bash
git add internal/services/view.go
git commit -m "refactor(views): remove dead SQL generation code

Remove ~500 lines of SQL-specific code:
- GenerateSQL and all SQL helpers
- SQL validation functions
- SQL escaping utilities

View execution now uses DSL parser and Bleve search."
```

### Step 7.5: Update ValidateViewDefinition for DSL

```go
// Replace old ValidateViewDefinition with DSL-aware version

func (s *ViewService) ValidateViewDefinition(view *core.ViewDefinition) error {
    if view.Name == "" {
        return fmt.Errorf("view name is required")
    }
    
    if !isValidViewName(view.Name) {
        return fmt.Errorf("invalid view name: %s", view.Name)
    }
    
    // Special views don't need query validation
    if view.IsSpecialView() {
        return nil
    }
    
    // Validate query can be parsed
    filter, directives := SplitViewQuery(view.Query)
    
    if filter != "" {
        p := parser.New()
        if _, err := p.Parse(filter); err != nil {
            return fmt.Errorf("invalid filter DSL: %w", err)
        }
    }
    
    if directives != "" {
        if _, err := ParseDirectives(directives); err != nil {
            return fmt.Errorf("invalid directives: %w", err)
        }
    }
    
    // Validate parameters
    for _, param := range view.Parameters {
        if err := s.validateViewParameter(&param); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Step 7.6: Run full test suite

```bash
mise run test
```

Expected: All tests pass

### Step 7.7: Commit validation update

```bash
git add internal/services/view.go
git commit -m "refactor(views): update ValidateViewDefinition for DSL queries

Validation now:
- Checks query can be parsed by DSL parser
- Validates directives format
- Keeps parameter validation

Removes all SQL-specific validation logic."
```

---

## Task 8: Update CLI Commands

**Files:**
- Modify: `cmd/notes_view.go`
- Add tests as needed

### Step 8.1: Update view command to use new execution

The `notes view` command should call `ExecuteView()` instead of SQL-based execution.

```go
// cmd/notes_view.go - update RunE function

func runViewCommand(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    
    // ... existing flag parsing ...
    
    // Get view definition
    viewDef, err := viewService.GetView(ctx, viewName, notebookPath)
    if err != nil {
        return err
    }
    
    // Parse parameters
    params, err := viewService.ParseViewParameters(paramFlag)
    if err != nil {
        return err
    }
    
    // Execute view
    results, err := viewService.ExecuteView(ctx, viewDef, params)
    if err != nil {
        return err
    }
    
    // Render results
    if results.Groups != nil {
        return renderGroupedResults(results.Groups, format)
    }
    return renderNoteList(results.Notes, format)
}
```

### Step 8.2: Test CLI manually

```bash
mise run build
./dist/opennotes notes view today
./dist/opennotes notes view recent
./dist/opennotes notes view --list
```

### Step 8.3: Commit CLI update

```bash
git add cmd/notes_view.go
git commit -m "feat(cli): update notes view to use DSL execution

View command now uses ExecuteView() which:
- Parses DSL query with pipe syntax
- Executes via Bleve search
- Supports grouping for kanban view"
```

---

## Task 9: Add Pipe Syntax Support to `notes search`

**Files:**
- Modify: `cmd/notes_search.go`

### Step 9.1: Update search command to detect pipe syntax

```go
// cmd/notes_search.go - update to handle pipe syntax in query string

func runSearchCommand(cmd *cobra.Command, args []string) error {
    query := args[0]
    
    // Check if query contains pipe syntax
    if strings.Contains(query, "|") {
        return runSearchWithPipeSyntax(cmd.Context(), query)
    }
    
    // ... existing search logic ...
}

func runSearchWithPipeSyntax(ctx context.Context, query string) error {
    filter, directives := services.SplitViewQuery(query)
    dirs, err := services.ParseDirectives(directives)
    if err != nil {
        return err
    }
    
    // Build FindOpts from directives
    opts := search.FindOpts{
        Limit:  dirs.Limit,
        Offset: dirs.Offset,
    }
    
    if filter != "" {
        p := parser.New()
        q, err := p.Parse(filter)
        if err != nil {
            return err
        }
        opts.Query = q
    }
    
    if dirs.SortField != "" {
        opts.Sort = directiveToSortSpec(dirs.SortField, dirs.SortDirection)
    }
    
    // Execute and render
    notes, err := noteService.SearchWithOptions(ctx, opts)
    if err != nil {
        return err
    }
    
    return renderNoteList(notes, format)
}
```

### Step 9.2: Test manually

```bash
mise run build
./dist/opennotes notes search "tag:work | sort:modified:desc limit:5"
```

### Step 9.3: Commit

```bash
git add cmd/notes_search.go
git commit -m "feat(cli): add pipe syntax support to notes search

Search command now accepts pipe-separated queries:
  notes search 'tag:work | sort:modified:desc limit:10'

Same syntax works in both view definitions and ad-hoc search."
```

---

## Task 10: Integration Testing

**Files:**
- Create: `internal/services/view_integration_test.go`

### Step 10.1: Write integration tests

```go
// internal/services/view_integration_test.go
package services_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
)

func TestViewExecution_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    ctx := context.Background()
    notebook := setupIntegrationNotebook(t)
    defer notebook.Cleanup()
    
    vs := NewViewService(notebook.NoteService, notebook.SearchService)
    
    t.Run("all builtin views execute without error", func(t *testing.T) {
        builtins := []string{"today", "recent", "kanban", "untagged", "orphans", "broken-links"}
        
        for _, name := range builtins {
            view, err := vs.GetView(ctx, name, notebook.Path)
            require.NoError(t, err, "failed to get view %s", name)
            
            _, err = vs.ExecuteView(ctx, view, nil)
            require.NoError(t, err, "failed to execute view %s", name)
        }
    })
}
```

### Step 10.2: Run integration tests

```bash
mise run test -- -run "TestViewExecution_Integration"
```

### Step 10.3: Commit

```bash
git add internal/services/view_integration_test.go
git commit -m "test(views): add integration tests for view execution

Verify all builtin views execute without error:
- today, recent, kanban, untagged (DSL-based)
- orphans, broken-links (special views)"
```

---

## Verification Checklist

After completing all tasks:

- [ ] `mise run test` — all tests pass
- [ ] `mise run lint` — no lint errors
- [ ] `mise run build` — binary compiles
- [ ] Manual test: `./dist/opennotes notes view today`
- [ ] Manual test: `./dist/opennotes notes view recent`
- [ ] Manual test: `./dist/opennotes notes view kanban`
- [ ] Manual test: `./dist/opennotes notes view untagged`
- [ ] Manual test: `./dist/opennotes notes view orphans`
- [ ] Manual test: `./dist/opennotes notes view --list`
- [ ] Manual test: `./dist/opennotes notes search "tag:work | sort:modified:desc"`
- [ ] `view.go` line count reduced from ~1355 to ~700
- [ ] No SQL-related imports in `view.go`

---

## Post-Implementation: Future Work

Items deferred from this implementation:

1. **View save/delete commands** (`--save`, `--delete` flags)
2. **CLI flag overrides** (`--sort`, `--limit` flags on view command)
3. **OR syntax in DSL** (parser can produce OrExpr but grammar doesn't support)
4. **View parameters** (`{{param_name}}` substitution)
5. **Global views** (views in `~/.config/opennotes/config.json`)

These can be implemented in follow-up tasks after the core DSL view system is working.

---

## Follow-up Tasks from Review (2026-02-18)

> Source: independent subagent review using plan checklist + CLI validation via `go run . ...`.

### F1. Implement `ExistsExpr` translation in Bleve backend (BLOCKER)

**Problem:** `kanban` and `untagged` views fail at runtime even though parser support exists.

**Evidence (CLI):**
- `go run . notes view kanban` → `failed to translate query: unsupported expression type: search.ExistsExpr`
- `go run . notes view untagged` → same root cause

**Scope:**
- Modify: `internal/search/bleve/query.go`
- Add `search.ExistsExpr` handling in `translateExpr()`.
- Implement translation for:
  - `has:<field>` (`Negated=false`)
  - `missing:<field>` (`Negated=true`)

**Acceptance Criteria:**
- `go run . notes view kanban` succeeds
- `go run . notes view untagged` succeeds
- `mise run test -- -run "TestViewExecution_Integration|TestViewService_ExecuteView"` passes without skipping `kanban`/`untagged` due to missing exists support

### F2. Remove temporary integration skips tied to missing exists translation

**Problem:** Integration tests currently skip paths for `kanban`/`untagged` because backend exists translation is incomplete.

**Scope:**
- Modify: `internal/services/view_integration_test.go`
- Remove conditional `t.Skip(...)` guards that reference missing `ExistsExpr` support.

**Acceptance Criteria:**
- `TestViewExecution_Integration` fully executes all builtin views without skip conditions for `kanban` and `untagged`

### F3. Optional cleanup: reduce `internal/services/view.go` size toward plan target

**Problem:** `view.go` remains above the plan’s rough target (~700 lines).

**Scope (optional):**
- Move cohesive helper logic to focused files if needed (no behavior change).

**Acceptance Criteria:**
- No functional regression (`mise run test`, `mise run lint`, `mise run build` pass)
- `view.go` size is reduced and responsibilities remain clear
