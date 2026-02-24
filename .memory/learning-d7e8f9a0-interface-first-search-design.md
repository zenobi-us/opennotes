---
id: d7e8f9a0
title: Interface-First Design for Search Engine Replacement
type: "learning"
created_at: 2026-02-19T08:11:00+10:30
updated_at: 2026-02-19T08:11:00+10:30
status: active
tags: [architecture, interfaces, search, bleve, parser, decoupling]
epic_id: f661c068
---

# Interface-First Design for Search Engine Replacement

## Summary

Defining interfaces (Index, Parser, Query AST) before any implementation enabled: parallel development of parser and backend, zero coupling between components, and trivial integration. The parser knows nothing about Bleve; Bleve knows nothing about the parser. The Query AST is the only shared boundary, and it took 6 lines to wire them together.

## Details

### The Three-Layer Architecture

```
User Query String → [Parser] → Query AST → [Translator] → Bleve Query → [Index] → Results
```

Each layer has a clean interface boundary:
1. **Parser** (`internal/search/parser/`): String → `search.Query` AST
2. **Translator** (`internal/search/bleve/query.go`): `search.Query` → Bleve-native queries
3. **Index** (`internal/search/bleve/index.go`): Executes Bleve queries, returns `search.Document`

### Why Interface-First Mattered

**Phase 2** defined interfaces. **Phase 3** built the parser. **Phase 4** built Bleve. These could have been parallel because the interfaces were the contract.

The integration point was trivial:

```go
func (idx *Index) FindByQueryString(ctx context.Context, queryString string, opts search.FindOpts) (search.Results, error) {
    p := parser.New()
    query, err := p.Parse(queryString)
    if err != nil {
        return search.Results{}, fmt.Errorf("failed to parse query: %w", err)
    }
    opts.Query = query
    return idx.Find(ctx, opts)
}
```

### Key Design Decisions

**1. Query AST as Domain Model**

The `search.Query` type is our domain model, not Bleve's. This means:
- Parser produces our types (not Bleve queries)
- Translator converts to Bleve (could convert to anything else)
- Swapping Bleve for another engine = new translator only

**2. Minimal Index Interface (10 methods)**

```go
type Index interface {
    Add(ctx, doc) error
    Remove(ctx, path) error
    Find(ctx, opts) (Results, error)
    FindByPath(ctx, path) (*Document, error)
    Count(ctx) (int, error)
    Stats(ctx) (IndexStats, error)
    Reindex(ctx) error
    Close() error
    // + FindByQueryString, Open
}
```

Keeping the interface small prevented over-abstraction. Every method has a direct use case in the CLI.

**3. Field Weights in Mapping, Not Queries**

BM25 field weights (Path:1000, Title:500, Tags:300, Lead:50, Body:1) are set once in the document mapping. Every query automatically benefits—no per-query boosting logic needed.

**4. MatchQuery vs TermQuery for Analyzed Fields**

Critical Bleve subtlety: `TermQuery` bypasses the analyzer (exact byte match), while `MatchQuery` uses the field's analyzer. For tag fields using `SimpleAnalyzer` (lowercases), you must use `MatchQuery` or tags like "Work" won't match queries for "work".

### Grammar Evolution Strategy

Started with MVP grammar, extended incrementally:
1. Simple terms: `meeting notes` (Phase 3)
2. Field expressions: `tag:work` (Phase 3)
3. Negation: `-archived` (Phase 3)
4. OR logic: `tag:a OR tag:b` (Phase 3)
5. Date expressions: `created:>2024-01-01` (Phase 3)

Each addition was a grammar rule extension, not a rewrite. Participle's type-safe AST made this safe.

## Implications

- **For search replacements**: Define your Query AST as a domain model independent of any search engine. The translator pattern makes engines swappable.
- **For parser design**: Start with the simplest grammar that's useful. Extend incrementally. Participle (Go) makes grammar changes safe via type checking.
- **For Bleve specifically**: Always match query type to field mapping—MatchQuery for analyzed fields, TermQuery for keyword fields. Set field weights in mapping, not per-query.
- **Related learnings**: [archive/duckdb-removal-f661c068/learning-6ba0a703-bleve-backend-implementation.md](archive/duckdb-removal-f661c068/learning-6ba0a703-bleve-backend-implementation.md)
