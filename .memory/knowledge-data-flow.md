---
id: b2c3d4e5
title: OpenNotes Data Flow Diagram
type: "knowledge"
created_at: 2026-01-18T19:31:53+10:30
updated_at: 2026-02-14T18:33:00+10:30
status: active
area: data-flow
tags: [architecture, data-flow, state-machine, bleve, cli, pi-extension]
learned_from: [codemapper-stats-2026-02-14, source-exploration]
---

# OpenNotes Data Flow Diagram

## Overview

This captures the **current** end-to-end data movement for native CLI workflows and the Pi extension tool workflows.

## ASCII Data Flow State Machine

```text
┌──────────────────────────────────────────────────────────────────────────┐
│                      DATA FLOW: NATIVE GO CLI                           │
└──────────────────────────────────────────────────────────────────────────┘

[User command]
    │
    ▼
[Cobra arg/flag parsing]
    │
    ▼
[requireNotebook]
    │
    ├─ OPENNOTES_NOTEBOOK env
    ├─ --notebook flag
    └─ NotebookService.Infer (cwd/context/ancestor)
    │
    ▼
[NotebookService.Open/Create]
    │
    ▼
[Load .opennotes.json -> resolve notebook root]
    │
    ▼
[createIndex]
    │
    ▼
[Walk markdown files]
    │
    ▼
[Read file -> parse YAML frontmatter -> extract title/tags/lead/timestamps]
    │
    ▼
[search.Document]
    │
    ▼
[BleveDocument + Index.Add]
    │
    ▼
[NoteService]
    │
    ├─ SearchNotes(query,fuzzy)
    │   ├─ getAllNotes(Index.Count + Index.Find)
    │   ├─ SearchService.TextSearch OR FuzzySearch (in-memory note slice)
    │   └─ []Note
    │
    └─ SearchWithConditions(conditions)
        ├─ SearchService.BuildQuery -> search.Query AST
        ├─ Index.Find(opts.Query)
        └─ []Note
    │
    ▼
[TuiRender template OR fallback formatter]
    │
    ▼
[Terminal output]


┌──────────────────────────────────────────────────────────────────────────┐
│                 DATA FLOW: QUERY AST -> BLEVE                           │
└──────────────────────────────────────────────────────────────────────────┘

[QueryCondition list]
    │
    ▼
[BuildQuery]
    │
    ├─ normalize data.* -> metadata.*
    ├─ path exact/prefix/wildcard detection
    ├─ OR tree assembly
    └─ NOT wrapping
    │
    ▼
[search.Query]
    │
    ▼
[TranslateFindOpts / TranslateQuery]
    │
    ▼
[Bleve query primitives]
    │
    ▼
[Index.Search]
    │
    ▼
[search.Results + snippets + score]
    │
    ▼
[documentToNote mapping]


┌──────────────────────────────────────────────────────────────────────────┐
│                    DATA FLOW: PI EXTENSION                              │
└──────────────────────────────────────────────────────────────────────────┘

[Pi tool call]
    │
    ▼
[Tool schema validation (zod/typebox layer)]
    │
    ▼
[Tool execute()]
    │
    ▼
[TS Service (search/list/get/create/notebooks/views)]
    │
    ▼
[CliAdapter.exec("opennotes", args)]
    │
    ▼
[Go CLI execution]
    │
    ▼
[stdout/stderr]
    │
    ├─ parseJsonOutput<T>()
    ├─ pagination.fitToBudget/paginate
    └─ format* output utilities
    │
    ▼
[ToolResult content for LLM]
```

## Data Contracts (Key Boundaries)

- `search.Document` ↔ `BleveDocument` mapping in `internal/search/bleve/`.
- `search.Result` wraps `Document + score + snippets`.
- `services.Note` is the CLI-friendly shape derived from `search.Document`.
- Pi extension service contracts are in `pkgs/pi-opennotes/src/services/types.ts`.

## Notable Flow Constraints

1. Indexing is currently in-memory (`bleve.NewMemOnly`) during notebook open/create.
2. Boolean condition flow supports metadata/path/title; link graph queries currently error with guidance.
3. View execution via CLI command is intentionally disabled (migration messaging), while list mode remains available.
