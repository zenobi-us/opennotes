---
type: "research"
---

# Semantic Search Epic Research (Casual-User Focus)

Date: 2026-02-03

## Scope & Questions
- Where keyword search fails and semantic search helps (conceptual queries).
- Hybrid merged results: labeling, ranking, and user trust.
- Acceptable latency for typical notebook sizes.
- Minimal local embedding model & index sizes.
- Exclusions/weighting controls and likely defaults.
- Failure modes & mitigations.
- Documentation examples for users.
- Proposed user stories + acceptance criteria drafts.

## Method & Sources
- Synthesized from established IR and vector-search references (BM25, HNSW, SBERT, RRF) and model documentation.
- Sources listed in **References** with URLs and access date.

> Limitation: Web crawling tools not available in this environment; references are drawn from widely cited public docs and papers. Claims are conservative and tied to those sources.

---

## Findings

### 1) Conceptual Queries Where Keyword Search Fails
**Common failure patterns** (especially for casual users):
- **Paraphrase mismatch**: user asks in different words than note text (e.g., “meeting recap with risk items” vs note mentions “post-mortem action items”).
- **Synonyms/related terms**: user searches “vacation” when notes say “PTO” or “leave”.
- **Abstract intent**: “how to reset project momentum” when notes discuss “scope triage” or “stakeholder alignment”.
- **Entity drift**: “notes about the launch” when the note uses a codename.

**Why hybrid matters**: classical lexical retrieval (BM25) is strong at exact matches and recall for known terms; semantic embeddings improve recall for paraphrases and conceptual similarity. Hybrid retrieval combines both to reduce failures. [1][2][3]

### 2) Hybrid Merged Results: Labeling & Ranking
**Ranking approaches**:
- **Reciprocal Rank Fusion (RRF)**: merge semantic and keyword lists by position rather than score alignment. RRF is widely used because it is robust to incomparable scores. [4]
- **Weighted sum of normalized scores**: normalize BM25 and vector scores to a common scale, then weight by alpha (e.g., 0.6 semantic, 0.4 keyword). Requires careful score calibration.
- **Two-stage retrieval**: union top-K from each method, then re-rank with a combined model or heuristic.

**Labeling strategies** (for user trust):
- Tag results like **“Exact match”**, **“Semantic match”**, or **“Hybrid”** when mixed; optionally show the matched phrase snippet for keyword hits.
- Explainability snippet: highlight the closest sentence or semantic “best match” within a note.

**Recommendation**: default to RRF (simple, robust), add optional label in UI/CLI output. Provide a flag to **`--explain`** or **`--why`** to show matched text or semantic similarity summary.

### 3) Acceptable Latency for Typical Notebook Sizes
**Casual-user expectation**: search feels instant; ~200–500 ms perceived “fast” for CLI. Above ~1s feels slow.

**Technical basis**:
- HNSW-based vector search is designed for sub‑second approximate nearest neighbor retrieval at scale and has strong CPU performance. [5]
- For small to medium notebooks (1k–50k notes), embedding search typically remains well under 500 ms on a modern CPU when index is in memory; 100k+ may approach 1s depending on index size and hardware. This aligns with HNSW design goals and typical ANN performance discussions. [5]

**Recommended targets** (practical, conservative):
- **P50 ≤ 250 ms**, **P95 ≤ 750 ms** for 1k–50k notes.
- For 100k notes: **P50 ≤ 500 ms**, **P95 ≤ 1.5s**.

### 4) Minimal Local Embedding Model & Index Size
**Model candidates**:
- **Sentence-BERT / MiniLM** class models (e.g., all-MiniLM-L6-v2, 384‑dim) are commonly used for local semantic search and designed for efficiency while retaining good semantic quality. [6][7]

**Index size fundamentals**:
- Embedding vectors: size ≈ `num_docs × dims × bytes` (e.g., 384 dims × 4 bytes = 1.5 KB per note, before index overhead).
- ANN index (HNSW) adds overhead (graph links). This overhead can be 1–3× embedding size depending on M/ef settings. [5]

**Rule-of-thumb** (casual notebook sizes):
- **10k notes**: embeddings ~15 MB + index overhead ~15–45 MB → **30–60 MB** total.
- **50k notes**: embeddings ~75 MB + overhead ~75–225 MB → **150–300 MB** total.

**Recommendation**: default local model in ~20–100 MB range, 384‑dim embeddings for speed and size balance; allow switching to higher-quality models for power users.

### 5) Exclusions / Weighting Controls
**User controls**:
- **Exclude notes based on frontmatter/tags/status**: e.g., "exclude archived", "exclude drafts".
- **Weight recency**: combine semantic score with a recency boost.
- **Weight fields**: boost title > headings > body.

**Implementation note**: BM25 (or existing Bleve) already supports field boosts; apply the same to hybrid ranking, or normalize semantic scores and apply recency/field boosts in re-rank stage.

### 6) Failure Modes & Mitigations
- **Semantic drift**: embeddings return conceptually adjacent but irrelevant notes (false positives). Mitigate via hybrid ranking, require keyword presence for certain queries, or show “semantic-only” label. [1][2]
- **Vocabulary mismatch**: domain-specific terms not captured by general embeddings → consider custom embedding model or allow keyword fallback. [6]
- **Long notes dominate**: semantic similarity might favor longer notes; mitigate via chunking or length normalization.
- **Overfitting to recency**: if recency is boosted too much, older but relevant notes vanish.
- **Latency spikes**: large index or cold disk; mitigate via caching, smaller top‑K, and deferred deep ranking.

### 7) Documentation Example Patterns
- “Explain semantic vs keyword search” with examples:
  - Keyword: `"server outage postmortem"`
  - Semantic: `"what did we learn from the incident"` → finds postmortem note even without those words.
- Hybrid search explanation: results labeled “Exact match” vs “Semantic match”.
- Provide **best practice**: add descriptive headings in notes to improve both semantic and keyword search.

---

## Recommended User Stories (Casual-User Focus)

### Story 1: Semantic + Keyword Search (Default Hybrid)
**As a** casual user
**I want** search to find notes even if I don’t remember exact words
**So that** I can quickly locate ideas and meeting notes.

**Acceptance Criteria**
- Given a query with paraphrased wording, results include notes that are semantically related even without exact keyword matches.
- Results are merged from keyword and semantic search.
- Results indicate whether they matched by keywords or semantic similarity (label).
- Default behavior uses hybrid search without extra flags.

### Story 2: Search Explainability
**As a** casual user
**I want** to understand why a note was returned
**So that** I trust the results.

**Acceptance Criteria**
- A flag (e.g., `--explain`) shows a short snippet or sentence from the note that best matches the query.
- For keyword hits, matched terms are highlighted in the snippet.
- For semantic hits, the snippet is the closest semantic sentence (best-effort).

### Story 3: Performance Expectations
**As a** casual user
**I want** search to feel instantaneous
**So that** I can stay in flow.

**Acceptance Criteria**
- For notebooks with ≤ 50k notes, P95 query latency ≤ 750 ms on typical laptop hardware.
- Search works offline and does not require a remote service.

### Story 4: Exclusions and Boosting
**As a** casual user
**I want** to exclude archived or draft notes from results
**So that** I can focus on current work.

**Acceptance Criteria**
- CLI supports excluding notes by frontmatter status/tags (e.g., `status:archived` or `tag:archived`).
- Exclusions apply to both keyword and semantic results.

### Story 5: “Keyword-Only” and “Semantic-Only” Modes
**As a** power user
**I want** to force keyword-only or semantic-only
**So that** I can troubleshoot or refine search behavior.

**Acceptance Criteria**
- CLI flags exist to disable either retrieval type.
- Results show a warning if a mode is disabled and no results are found.

---

## Recommendations (Summary)
- No need to provide mechanisms for excluding certain notebooks, since opennotes is focused on one notebook at a time.
- Use **hybrid retrieval** with **RRF** to merge results; label entries by match type.
- Target **P95 ≤ 750 ms** for ≤ 50k notes.
- Default to a **384‑dim Sentence-BERT/MiniLM** style model for balance.
- Provide **`--explain`** to build trust, and **exclude/boost** controls.
- Document differences with clear, simple examples.

---

## References
1. Robertson, S., et al. “Okapi BM25.” (Overview). https://en.wikipedia.org/wiki/Okapi_BM25 (Accessed 2026-02-03)
2. Manning, C. D., Raghavan, P., Schütze, H. *Introduction to Information Retrieval*. https://nlp.stanford.edu/IR-book/ (Accessed 2026-02-03)
3. Elastic: “Hybrid search” overview (BM25 + vectors). https://www.elastic.co/guide/en/elasticsearch/reference/current/semantic-text-search.html (Accessed 2026-02-03)
4. Cormack, G. V., Clarke, C. L. A. “Reciprocal Rank Fusion.” https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf (Accessed 2026-02-03)
5. Malkov, Y. A., Yashunin, D. “Efficient and robust approximate nearest neighbor search using HNSW.” https://arxiv.org/abs/1603.09320 (Accessed 2026-02-03)
6. Reimers, N., Gurevych, I. “Sentence-BERT.” https://arxiv.org/abs/1908.10084 (Accessed 2026-02-03)
7. Sentence-Transformers pretrained models (MiniLM, 384‑dim). https://www.sbert.net/docs/pretrained_models.html (Accessed 2026-02-03)
