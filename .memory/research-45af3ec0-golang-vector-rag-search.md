---
id: 45af3ec0
title: Go-based Vector RAG Search Implementation
type: "research"
created_at: 2026-02-01T15:07:00+10:30
updated_at: 2026-02-01T15:07:00+10:30
status: todo
epic_id: f661c068
related_task_id: dbb5cdc8
---

# Go-based Vector RAG Search Implementation

## Research Questions

1. What are the leading Go-based vector database/search libraries available?
2. Which libraries support semantic search and RAG (Retrieval Augmented Generation) patterns?
3. How do these libraries compare in terms of:
   - Performance (indexing speed, query latency)
   - Memory footprint
   - API ergonomics and ease of integration
   - Dependencies and build complexity
   - Support for embeddings (OpenAI, local models, etc.)
4. Can any of these integrate with `spf13/afero` for filesystem abstraction?
5. What would a minimal RAG implementation look like in Go?
6. How does vector search complement or replace traditional text-based search?

## Summary

This research explores Go-based vector search and RAG capabilities as a potential search strategy for OpenNotes. While zk-org/zk (analyzed in [research-dbb5cdc8](research-dbb5cdc8-zk-search-analysis.md)) provides traditional text-based search, vector embeddings enable semantic search that can find notes by meaning rather than exact keywords.

**Inspiration**: The qmd tool (https://github.com/tobi/qmd) is a Node.js-based markdown query tool that demonstrates RAG patterns, but cannot be used directly since OpenNotes is written in Go. This research identifies Go equivalents.

**Key Goal**: Determine if vector-based search should be:
- A replacement for traditional text search
- A complementary feature (hybrid search)
- Deferred to a future phase
- Not pursued due to complexity/tradeoffs

## Findings

### Go Vector Database Libraries

_To be populated with research findings:_

#### Option 1: Chroma-go
- **Repository**: 
- **License**: 
- **Description**: 
- **Pros**: 
- **Cons**: 
- **Integration with afero**: 

#### Option 2: Milvus
- **Repository**: 
- **License**: 
- **Description**: 
- **Pros**: 
- **Cons**: 
- **Integration with afero**: 

#### Option 3: Weaviate Go Client
- **Repository**: 
- **License**: 
- **Description**: 
- **Pros**: 
- **Cons**: 
- **Integration with afero**: 

#### Option 4: Pure Go Implementations
- **vecty / go-ann / other pure-Go options**
- **Description**: 
- **Pros**: 
- **Cons**: 
- **Integration with afero**: 

### RAG Pattern in Go

_To be populated:_

1. **Embedding Generation**
   - Local models (e.g., sentence-transformers via ONNX)
   - API-based (OpenAI, Cohere, etc.)
   - Tradeoffs: cost, latency, privacy

2. **Vector Storage & Retrieval**
   - In-memory vs persistent
   - Index types (HNSW, IVF, etc.)
   - Similarity metrics (cosine, L2, dot product)

3. **Query Flow**
   ```
   User Query → Embed → Vector Search → Top-K Results → Rerank → Display
   ```

4. **Integration with OpenNotes**
   - Indexing: Generate embeddings during note scanning
   - Storage: Embed vectors in note frontmatter or separate index?
   - Query: Hybrid search (text + vector) or pure semantic?

### Performance Considerations

_To be populated:_

- Indexing time for 10k notes
- Query latency targets (<100ms goal)
- Memory overhead per note
- Build/deployment complexity

### Hybrid Search Strategy

_To be populated:_

Combining traditional text search (from zk-style implementation) with vector search:

1. **Use Cases**:
   - Text search: Exact matches, field filters, date ranges
   - Vector search: "Find notes similar to X", concept exploration
   
2. **Architecture**:
   - Both indexes run in parallel
   - Unified query interface (DSL supports both modes)
   - Results can be merged/ranked

3. **Tradeoffs**:
   - Increased complexity
   - Higher memory usage
   - Richer user experience

## References

### Inspiration
- [qmd (Node.js)](https://github.com/tobi/qmd) - Markdown query tool demonstrating RAG patterns (Node.js, cannot use directly)

### Go Vector Libraries
_To be populated with actual findings:_

- [ ] Chroma-go - https://github.com/...
- [ ] Milvus Go SDK - https://github.com/...
- [ ] Weaviate Go Client - https://github.com/...
- [ ] Pure Go vector search implementations
- [ ] Embedding libraries for Go (ONNX, TensorFlow Lite, etc.)

### RAG Resources
_To be populated:_

- [ ] RAG architecture patterns
- [ ] Go embedding generation examples
- [ ] Hybrid search strategies
- [ ] Performance benchmarks

### Related OpenNotes Research
- [research-dbb5cdc8-zk-search-analysis.md](research-dbb5cdc8-zk-search-analysis.md) - Traditional text search via zk-org/zk
- [research-4e873bd0-vfs-summary.md](research-4e873bd0-vfs-summary.md) - Filesystem abstraction research
- [research-7f4c2e1a-afero-vfs-integration.md](research-7f4c2e1a-afero-vfs-integration.md) - Afero integration patterns

## Next Steps

1. **Initial Survey** (1-2 hours)
   - Search GitHub for "golang vector search", "go embedding", "go rag"
   - Identify 3-5 candidate libraries
   - Quick README review for each

2. **Deep Dive** (2-4 hours)
   - Clone top 2-3 candidates
   - Run example code
   - Measure performance with sample data
   - Check afero compatibility

3. **Prototype** (4-8 hours, if promising)
   - Build minimal RAG example with OpenNotes-like data
   - Compare to text search baseline
   - Estimate integration effort

4. **Decision** (after findings)
   - Recommend: vector search, hybrid, or defer
   - Update epic-f661c068 with decision
   - Create phase plan if proceeding

## Recommendations

_To be filled after research completes:_

### Short-term (Epic f661c068)
- [ ] Recommendation on including vector search in current epic
- [ ] If yes: Which library and why
- [ ] If no: Defer to future epic with justification

### Long-term Vision
- [ ] Pure text search (Phase 1)
- [ ] Hybrid search (Phase 2, if valuable)
- [ ] Full semantic RAG (Phase 3, if justified)

---

**Research Status**: 📋 TODO - Ready to start

**Assigned**: Unassigned (waiting for delegation)

**Estimated Effort**: 4-8 hours (survey + deep dive + prototype)

**Dependencies**: 
- None (can run in parallel with zk-search-analysis)
- Results inform Phase 2 (Query DSL Design) of epic-f661c068
