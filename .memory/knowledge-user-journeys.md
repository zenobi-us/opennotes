---
id: c7f2a9d1
title: OpenNotes User Journeys Map
type: "knowledge"
created_at: 2026-02-14T18:33:00+10:30
updated_at: 2026-02-14T18:33:00+10:30
status: active
area: user-journeys
tags: [journeys, cli, pi-extension, ux]
learned_from: [codemapper-stats-2026-02-14, source-exploration]
---

# OpenNotes User Journeys Map

## Overview

Primary user journeys supported by the current codebase across both the native CLI and Pi extension.

## ASCII Journey Diagram

```text
┌──────────────────────────────────────────────────────────────────────────┐
│                         JOURNEY MAP: CLI USERS                          │
└──────────────────────────────────────────────────────────────────────────┘

(1) Start / Setup Journey
[Install opennotes]
   -> [opennotes init]
   -> [opennotes notebook create|register]
   -> [Notebook discoverable]

(2) Browse Notes Journey
[opennotes notes list]
   -> [requireNotebook]
   -> [index-backed load]
   -> [template list output]

(3) Text/Fuzzy Search Journey
[opennotes notes search <query> (--fuzzy)]
   -> [requireNotebook]
   -> [SearchNotes]
   -> [TextSearch OR FuzzySearch]
   -> [ranked/matched note list]

(4) Structured Filter Journey
[opennotes notes search query --and/--or/--not ...]
   -> [ParseConditions + validation]
   -> [BuildQuery AST]
   -> [Bleve execution]
   -> [filtered result list]

(5) Add Note Journey
[opennotes notes add ...]
   -> [path/title/frontmatter resolution]
   -> [markdown file write]
   -> [available on next index build/open]

(6) View Discovery Journey
[opennotes notes view --list (--format json)]
   -> [ViewService list hierarchy]
   -> [built-in + global + notebook views]
   -> [display list]

(7) View Execution Attempt Journey (current behavior)
[opennotes notes view <name>]
   -> [parameter parse]
   -> [migration message: SQL views removed]
   -> [guided fallback to search/list/json tooling]


┌──────────────────────────────────────────────────────────────────────────┐
│                      JOURNEY MAP: PI AGENT USERS                        │
└──────────────────────────────────────────────────────────────────────────┘

(A) Agent Session Boot Journey
[pi session_start]
   -> [pi-opennotes extension loads]
   -> [register tools]
   -> [CLI installation check + warning if missing]

(B) Tool-based Search Journey
[Tool: opennotes_search]
   -> [mode select: query/fuzzy/sql/filters]
   -> [TS SearchService]
   -> [CliAdapter.exec(opennotes ...)]
   -> [parse + paginate + format]
   -> [LLM-friendly response]

(C) Tool-based Note Retrieval Journey
[Tool: opennotes_get]
   -> [validate path]
   -> [CLI SQL query]
   -> [structured note payload]

(D) Tool-based Creation Journey
[Tool: opennotes_create]
   -> [validate title/path/template]
   -> [CLI notes add]
   -> [created note metadata returned]

(E) Tool-based Notebook/View Discovery Journey
[Tool: opennotes_notebooks / opennotes_views]
   -> [CLI notebook/view commands]
   -> [normalization + fallback behavior]
   -> [structured list for agent planning]
```

## Journey Notes

- The CLI currently provides strongest support for list/search/query and notebook management.
- View listing is supported; view execution is intentionally blocked by migration messaging.
- Pi extension journeys are tool-first and depend on reliable CLI JSON/text outputs.
