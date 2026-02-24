---
id: c5d6e7f8
title: Pure Go Architecture - CGO Elimination Benefits
type: "learning"
created_at: 2026-02-19T08:11:00+10:30
updated_at: 2026-02-19T08:11:00+10:30
status: active
tags: [go, cgo, architecture, performance, deployment, ci-cd]
epic_id: f661c068
---

# Pure Go Architecture - CGO Elimination Benefits

## Summary

Eliminating CGO (by replacing DuckDB with Bleve) produced cascading benefits far beyond the original search performance goal: 64% smaller binary, 97% faster startup, cross-compilation enabled, CI reliability from 50-70% to 100%, and simpler deployment. CGO was the single biggest source of friction across build, test, and deploy.

## Details

### Quantified Impact

| Dimension | With CGO (DuckDB) | Pure Go (Bleve) | Improvement |
|-----------|-------------------|-----------------|-------------|
| Binary size | 64 MB | 23 MB | -64% |
| Startup time | 500ms | 17ms | -97% |
| Search latency | 29.9ms | 0.754ms | -97% |
| CI reliability | 50-70% | 100% | Deterministic |
| CI build time | ~3 min | ~1 min | -67% |
| Cross-compile | ❌ Broken | ✅ Works | Enabled |
| Dependencies | 12 (with CGO) | 9 (pure Go) | -25% |

### Why CGO Was So Costly

1. **Binary bloat**: DuckDB + markdown extension = 37.8MB of the 64MB binary
2. **Startup overhead**: `duckdb.OpenExt()` took 400ms+ to initialize (9.5% CPU)
3. **CGO boundary crossing**: `runtime.cgocall` consumed 19.5% of CPU during search—pure waste since search was already in-memory
4. **Extension downloads**: DuckDB markdown extension download failed 30-50% in GitHub Actions CI ([learning-c6cf829a](learning-c6cf829a-duckdb-ci-extension-caching.md))
5. **Cross-compilation blocked**: CGO prevents `GOOS=windows GOARCH=amd64 go build`

### The Hidden Cost: Operational Complexity

Beyond performance, CGO created operational burden:
- CI needed special caching for DuckDB extensions
- Build required C compiler toolchain on all platforms
- Test environment setup was fragile (extension version mismatches)
- Docker images needed additional system libraries

### When to Choose Pure Go Over CGO Dependencies

**Choose pure Go when**:
- Tool is single-binary CLI (deployment simplicity matters)
- Cross-compilation is desired
- CI reliability is critical
- The CGO dependency's unique features aren't essential (search was in-memory anyway)

**Accept CGO when**:
- The dependency provides genuinely irreplaceable functionality
- Server deployment (not distributed binary)
- Performance-critical paths where C libraries are 10x+ faster
- Ecosystem has no viable pure-Go alternative

### Bleve vs DuckDB: Right Tool for the Job

DuckDB was originally chosen for SQL querying of markdown files. But over time, the actual search path was:
1. Read all markdown files from filesystem
2. Parse frontmatter in Go
3. Load into DuckDB
4. Query via SQL
5. Convert results back to Go structs

Steps 3-5 added overhead with zero benefit—the data was already in Go memory. Bleve eliminated these unnecessary round-trips.

## Implications

- **For new Go projects**: Default to pure Go dependencies. Only introduce CGO when benchmarks prove the pure Go alternative is inadequate for your specific workload.
- **For existing CGO dependencies**: Profile actual usage. If the CGO dependency is on a non-critical path (like DuckDB was for search), replacing it with pure Go yields disproportionate benefits.
- **For CI/CD**: Pure Go builds are deterministic and fast. CGO introduces environmental dependencies that cause flaky builds.
- **Related learnings**: [learning-c6cf829a-duckdb-ci-extension-caching.md](learning-c6cf829a-duckdb-ci-extension-caching.md) (the CI pain that motivated this)
