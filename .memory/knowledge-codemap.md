---
id: a1b2c3d4
title: OpenNotes Codebase Structure Map
type: "knowledge"
created_at: 2026-01-18T19:31:53+10:30
updated_at: 2026-02-14T18:33:00+10:30
status: active
area: codebase-structure
tags: [architecture, codebase, state-machine, bleve, search, pi-extension]
learned_from: [codemapper-stats-2026-02-14, source-exploration]
---

# OpenNotes Codebase Structure Map

## Overview

OpenNotes is a Go CLI with a Bleve-backed search engine and a TypeScript Pi extension (`pkgs/pi-opennotes`) that wraps the CLI as tools.

## Stats Snapshot (CodeMapper)

- Whole repo: **381 files** (`go:79`, `typescript:39`, `markdown:263`)
- `cmd/`: **15 Go files**, 28 functions
- `internal/`: **54 Go files**, 474 functions, 134 methods
- `pkgs/pi-opennotes/src`: **27 TS files**, 44 functions, 45 methods
- `tests/`: e2e and stress coverage present

## ASCII State Machine Codemap

```text
┌──────────────────────────────────────────────────────────────────────────┐
│                            OPENNOTES SYSTEM                             │
└──────────────────────────────────────────────────────────────────────────┘

                      [User / Script / Pi Tool Call]
                                   │
                                   ▼
                         [main.go -> cmd/root.go]
                                   │
                                   ▼
                 [PersistentPreRunE: InitLogger + Config + NotebookSvc]
                                   │
                 ┌─────────────────┴─────────────────┐
                 ▼                                   ▼
      [Cobra notebook/init/version]        [Cobra notes commands]
                                            list/search/search query/view

NATIVE CLI SEARCH PATH
======================

[notes search | notes list | notes search query]
                    │
                    ▼
             [requireNotebook(cmd)]
                    │
                    ▼
       [NotebookService.Infer/Open/Create]
                    │
                    ▼
 [Load .opennotes.json + createIndex(notebookRoot)]
                    │
                    ▼
 [walk *.md -> parse frontmatter -> build search.Document]
                    │
                    ▼
           [Bleve Index.Add (in-memory)]
                    │
                    ▼
               [NoteService]
             ┌───────────────┬──────────────────────┐
             ▼               ▼                      ▼
      SearchNotes()   SearchWithConditions()     Count()
             │               │
             │               ▼
             │      [SearchService.BuildQuery]
             │               │
             │               ▼
             │      [search.Query AST (AND/OR/NOT)]
             │               │
             └───────┬───────┘
                     ▼
          [search.Index.Find / Count]
                     │
                     ▼
       [bleve.TranslateFindOpts/TranslateQuery]
                     │
                     ▼
               [Bleve Search]
                     │
                     ▼
       [search.Results -> documentToNote]
                     │
                     ▼
  [TuiRender(note-list) -> fallback fmt.Printf -> terminal]

QUERY PARSER SUBSYSTEM (separate DSL path)
==========================================

[query string]
    │
    ▼
[internal/search/parser.Parser]
    │
    ▼
[Participle grammar AST]
    │
    ▼
[convert.go -> search.Query AST]
    │
    ▼
[bleve/query.go translator]

PI EXTENSION PATH (pkgs/pi-opennotes)
======================================

[Pi agent starts session]
          │
          ▼
[pi-opennotes/src/index.ts]
          │
          ├─ createServices(cli,pagination,search,list,note,notebook,views)
          └─ registerTools(opennotes_search/list/get/create/notebooks/views)
                     │
                     ▼
                [Tool execute()]
                     │
                     ▼
            [Service method (TS layer)]
                     │
                     ▼
      [CliAdapter.exec("opennotes", args...)]
                     │
                     ▼
          [Go CLI command execution path]
                     │
                     ▼
       [stdout JSON/text -> parse/format -> tool result]

STATE GROUPS
============

Index lifecycle:
[Unopened] -> [NewIndex] -> [Ready] -> [Find/Count] -> [Close]

Notebook resolution lifecycle:
[Env/Flag not set] -> [Infer current dir] -> [Context match] -> [Ancestor match] -> [Notebook ready | none]

Search execution lifecycle:
[Input] -> [Parse/validate] -> [Build query opts] -> [Index find] -> [Result mapping] -> [Render]
```

## Current Structural Notes

1. Go CLI search is Bleve-based and no longer depends on DuckDB service code paths.
2. `cmd/notes_view.go` currently supports **listing** views; execution path intentionally returns migration error text.
3. Pi extension still contains SQL-oriented service logic (`--sql`, DuckDB-style assumptions), so there is an integration drift to reconcile in future phases.
