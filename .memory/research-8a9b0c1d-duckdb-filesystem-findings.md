---
id: 8a9b0c1d
title: DuckDB Go Bindings - Filesystem Research Findings
type: "research"
created_at: 2026-01-23T22:00:00+10:30
updated_at: 2026-01-23T22:00:00+10:30
status: completed
epic_id: a9b3f2c1
tags:
  - research
  - duckdb
  - filesystem
  - storage-abstraction
---

# DuckDB Go Bindings - Filesystem Research Findings

## Executive Summary

**Question**: Can we use afero virtual filesystem with DuckDB's `read_markdown()` function?

**Answer**: No, not directly. DuckDB's `read_markdown()` bypasses Go entirely and uses DuckDB's internal C++ filesystem. However, there are **alternative approaches** available.

---

## Findings

### 1. DuckDB's Filesystem Architecture

DuckDB has a sophisticated **VirtualFileSystem** in C++:

```cpp
// From duckdb/common/virtual_file_system.hpp
class VirtualFileSystem : public FileSystem {
    void RegisterSubSystem(unique_ptr<FileSystem> fs);
    // ... supports custom filesystem registration
};
```

This supports:
- Custom filesystem registration (`RegisterSubSystem`)
- Multiple backends (httpfs, s3, gcs, azure)
- Extension-based filesystem loading

**But**: This API is **not exposed in the Go bindings**.

### 2. Go Bindings Available Features

The duckdb-go bindings expose:

| Feature | Available | Use Case |
|---------|-----------|----------|
| `ReplacementScan` | ✅ Yes | Replace table names with functions |
| `TableFunction` (UDF) | ✅ Yes | Custom table-producing functions |
| `RowTableSource` | ✅ Yes | Return rows from Go code |
| `ChunkTableSource` | ✅ Yes | Return chunks from Go code |
| `httpfs` extension | ✅ Yes | Load via SQL |
| Custom FileSystem | ❌ No | Not exposed in Go |

### 3. Potential Solutions

#### Option A: Table UDF (Custom `read_markdown_vfs`)

Create a Go-based Table UDF that:
1. Accepts a path/glob pattern
2. Uses afero to read files
3. Parses markdown in Go
4. Returns rows to DuckDB

```go
// Hypothetical implementation
type MarkdownVFSSource struct {
    fs      afero.Fs
    files   []string
    current int
}

func (s *MarkdownVFSSource) FillRow(row duckdb.Row) (bool, error) {
    if s.current >= len(s.files) {
        return false, nil
    }
    
    data, _ := afero.ReadFile(s.fs, s.files[s.current])
    parsed := parseMarkdown(data)
    
    row.SetVarchar(0, s.files[s.current])  // filepath
    row.SetVarchar(1, parsed.Title)         // title
    // ... etc
    
    s.current++
    return true, nil
}

// Register as: read_markdown_vfs(glob, fs_handle)
duckdb.RegisterTableFunction(c, "read_markdown_vfs", bindFunc)
```

**Pros**:
- Full afero integration
- Works with any Storage backend
- Pure Go, no C++ modifications

**Cons**:
- Must reimplement markdown parsing (DuckDB's extension is feature-rich)
- Performance may be slower than native extension
- Maintenance burden for feature parity

#### Option B: ReplacementScan for Abstraction Layer

Use ReplacementScan to intercept queries and rewrite paths:

```go
duckdb.RegisterReplacementScan(c, func(tableName string) (string, []any, error) {
    if strings.HasPrefix(tableName, "vfs://") {
        // Materialize files from afero to temp dir
        tempPath := materializeToTemp(aferoFs, tableName)
        return "read_markdown", []any{tempPath}, nil
    }
    return "", nil, errors.New("not handled")
})
```

**Pros**:
- Uses native DuckDB markdown parsing
- Simpler implementation

**Cons**:
- Still requires temp file materialization
- Overhead for large notebooks

#### Option C: Hybrid (Recommended)

Different storage backends use different query strategies:

| Storage Type | Query Method | Performance |
|--------------|--------------|-------------|
| OsFs | Native `read_markdown()` | ⚡ Fast |
| MemMapFs (tests) | Go-based parser OR temp dir | 🔧 Acceptable |
| S3/Cloud (future) | Materialize + native OR custom UDF | 🐢 Slower |

```go
func (ns *NoteService) Query(notebook *Notebook) ([]Note, error) {
    switch notebook.Storage.Type() {
    case "os":
        // Fast path: native DuckDB
        return ns.queryWithDuckDB(notebook)
    case "memory":
        // Test path: Go parser or temp materialization
        return ns.queryWithGoParser(notebook)
    default:
        // Cloud path: materialize then query
        return ns.queryWithMaterialization(notebook)
    }
}
```

### 4. httpfs Extension

DuckDB supports remote files via httpfs:

```sql
INSTALL httpfs;
LOAD httpfs;
SELECT * FROM read_markdown('https://example.com/notes/*.md');
```

Could be used for cloud-accessible notebooks, but:
- Requires public/signed URLs
- Doesn't work with afero abstraction
- Different from local filesystem semantics

---

## Recommendations

### For Immediate Goal (Test Isolation)

1. **ConfigService**: Full afero abstraction ✅
2. **NotebookService discovery**: Full afero abstraction ✅  
3. **Note/View queries (OsFs)**: Native DuckDB `read_markdown()` ✅
4. **Note/View queries (tests)**: Use `t.TempDir()` for isolated real directories

This solves the **primary problem** (test pollution of ~/.config) without reimplementing markdown parsing.

### For Future Cloud Storage

Three options ranked by effort:

1. **Materialize + Native** (Low effort)
   - Copy files from afero to temp dir
   - Run native `read_markdown()` on temp dir
   - Clean up after query

2. **Custom Table UDF** (Medium effort)
   - Implement `read_markdown_vfs` in Go
   - Use afero for file access
   - Reimplement markdown parsing

3. **Contribute to duckdb-go** (High effort)
   - Add FileSystem registration to Go bindings
   - Upstream contribution
   - Full native integration

---

## Architecture Decision

Based on research, the **Hybrid Approach** is recommended:

```
┌─────────────────────────────────────────────────────────────┐
│                    Storage Abstraction                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ConfigService ──→ afero.Fs ──→ Any Backend (works!)       │
│                                                             │
│  NotebookService (discovery) ──→ afero.Fs ──→ Any Backend  │
│                                                             │
│  NoteService (queries):                                     │
│    ├─ OsFs ──→ Native DuckDB read_markdown() [fast]        │
│    ├─ MemMapFs ──→ t.TempDir() + native [test isolation]   │
│    └─ S3/Cloud ──→ Materialize + native [future]           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Key Insight

The **query layer** and **storage layer** can have different abstraction strategies:

- **Storage Layer**: Full afero abstraction (all backends)
- **Query Layer**: Optimized per-backend (native where possible)

This gives us:
- ✅ Test isolation for config (primary goal)
- ✅ Fast production queries (native DuckDB)
- ✅ Future extensibility (cloud storage via materialization)
- ✅ No markdown parser reimplementation needed

---

## Next Steps

1. Proceed with hybrid storage abstraction
2. Use `t.TempDir()` for DuckDB-related tests
3. Document the query strategy pattern
4. Consider Table UDF for cloud storage when needed

---

## References

- duckdb-go repo: https://github.com/duckdb/duckdb-go
- DuckDB VirtualFileSystem: duckdb/common/virtual_file_system.hpp
- Table UDFs: duckdb-go/table_udf.go
- ReplacementScan: duckdb-go/replacement_scan.go
