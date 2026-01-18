---
id: b2c3d4e5
title: OpenNotes Data Flow Diagram
created_at: 2026-01-18T19:31:53+10:30
updated_at: 2026-01-18T19:31:53+10:30
status: active
area: data-flow
tags: [architecture, data-flow, state-machine]
learned_from: [test-improvement-epic, codebase-exploration]
---

# OpenNotes Data Flow Diagram

## Overview

Data flow patterns in OpenNotes showing how information moves through the system from user input to terminal output.

## Details

### ASCII Data Flow State Machine

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         OPENNOTES DATA FLOW                            │
│                       (Input → Processing → Output)                    │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                            PRIMARY FLOWS                               │
└─────────────────────────────────────────────────────────────────────────┘

Flow 1: SEARCH OPERATION
========================

[CLI Args] → [Flag Parse] → [Notebook Discovery] → [DuckDB Query] → [Template Render]
    │              │              │                     │               │
    │              ▼              ▼                     ▼               ▼
    │         [Validation]   [Config Load]        [Markdown Ext]   [Glamour Output]
    │              │              │                     │               │
    └──────────────┼──────────────┼─────────────────────┼───────────────┘
                   ▼              ▼                     ▼
               [Search Term]  [.opennotes.json] → [SQL Execution] → [Terminal Display]
                   │              │                     │               │
                   ▼              ▼                     ▼               ▼
               [SQL Builder]  [Notebook Path]     [Result Rows]    [User Success]

Flow 2: NOTE CREATION
====================

[CLI Args] → [Template Selection] → [File Creation] → [Editor Launch] → [File Save]
    │              │                     │                │              │
    │              ▼                     ▼                ▼              ▼
    │         [Template Load]       [Path Generate]  [User Edit]   [DuckDB Index]
    │              │                     │                │              │
    └──────────────┼─────────────────────┼────────────────┼──────────────┘
                   ▼                     ▼                ▼
               [Go Template]         [File System]   [Content Ready] → [Search Available]

Flow 3: NOTEBOOK MANAGEMENT
==========================

[CLI Args] → [Notebook Discovery] → [Config Operations] → [State Update] → [Display]
    │              │                      │                    │             │
    │              ▼                      ▼                    ▼             ▼
    │         [Ancestor Walk]        [JSON Parse/Write]  [Memory Update] [Template]
    │              │                      │                    │             │
    └──────────────┼──────────────────────┼────────────────────┼─────────────┘
                   ▼                      ▼                    ▼
               [Directory Tree]       [Config Merge]      [Global State] → [User View]

┌─────────────────────────────────────────────────────────────────────────┐
│                          SERVICE DATA FLOW                             │
└─────────────────────────────────────────────────────────────────────────┘

ConfigService Data Flow:
========================
[User Config] ←→ [Global Config] ←→ [Notebook Config] ←→ [Runtime Config]
      │                │                  │                    │
      ▼                ▼                  ▼                    ▼
[~/.config/...]  [Environment]    [.opennotes.json]    [Active Settings]
      │                │                  │                    │
      └────────────────┼──────────────────┼────────────────────┘
                       ▼                  ▼
                [Config Merge Logic] → [Final Configuration]

NotebookService Data Flow:
=========================
[Directory Path] → [Ancestor Search] → [Config Discovery] → [Validation]
       │                 │                   │                  │
       ▼                 ▼                   ▼                  ▼
[File System] → [Parent Directory] → [JSON Parse] → [Schema Check] → [Ready State]
       │                 │                   │                  │
       └─────────────────┼───────────────────┼──────────────────┘
                         ▼                   ▼
                    [Path Resolution] → [Configuration Object]

NoteService Data Flow:
=====================
[Search Query] → [SQL Generation] → [DuckDB Execute] → [Result Processing]
      │               │                   │                  │
      ▼               ▼                   ▼                  ▼
[Term Parse] → [WHERE Clause] → [Markdown Tables] → [Row Mapping] → [Note Objects]
      │               │                   │                  │
      └───────────────┼───────────────────┼──────────────────┘
                      ▼                   ▼
                [Query Builder] → [Database Results] → [Structured Data]

DisplayService Data Flow:
========================
[Data Object] → [Template Select] → [Template Execute] → [Glamour Render]
      │              │                    │                    │
      ▼              ▼                    ▼                    ▼
[JSON/Struct] → [Go Template] → [Markdown String] → [ANSI Output] → [Terminal]
      │              │                    │                    │
      └──────────────┼────────────────────┼────────────────────┘
                     ▼                    ▼
                [Data Binding] → [Formatted Content] → [User Display]

DatabaseService Data Flow:
=========================
[Connection Request] → [Pool Selection] → [DuckDB Instance] → [Query Execute]
         │                   │                 │                  │
         ▼                   ▼                 ▼                  ▼
[Read/Write Mode] → [Connection Pool] → [Extension Load] → [SQL Result] → [Close]
         │                   │                 │                  │
         └───────────────────┼─────────────────┼──────────────────┘
                             ▼                 ▼
                    [Pool Management] → [Markdown Extension] → [Safe Execute]

┌─────────────────────────────────────────────────────────────────────────┐
│                          ERROR FLOW PATTERNS                           │
└─────────────────────────────────────────────────────────────────────────┘

Error Propagation Flow:
======================
[Service Error] → [Error Wrap] → [Log Entry] → [User Message] → [Graceful Exit]
      │              │             │              │                │
      ▼              ▼             ▼              ▼                ▼
[Original Error] → [Context Add] → [Structured Log] → [Human Readable] → [Exit Code]
      │              │             │              │                │
      └──────────────┼─────────────┼──────────────┼────────────────┘
                     ▼             ▼              ▼
                [Error Chain] → [Debug Info] → [Recovery Option]

Validation Flow:
===============
[Input Data] → [Schema Check] → [Business Rules] → [Security Check] → [Accept/Reject]
     │              │               │                 │                  │
     ▼              ▼               ▼                 ▼                  ▼
[Raw Input] → [Type Validate] → [Range Check] → [Injection Guard] → [Clean Data]
     │              │               │                 │                  │
     └──────────────┼───────────────┼─────────────────┼──────────────────┘
                    ▼               ▼                 ▼
               [Input Parsing] → [Business Logic] → [Security Layer]

┌─────────────────────────────────────────────────────────────────────────┐
│                         STATE TRANSITIONS                              │
└─────────────────────────────────────────────────────────────────────────┘

Application State:
=================
[Startup] → [Service Init] → [Command Parse] → [Execute] → [Cleanup] → [Exit]
    │           │               │                │           │         │
    ▼           ▼               ▼                ▼           ▼         ▼
[Config] → [Singletons] → [Flag Processing] → [Business Logic] → [Resource Free] → [Status Code]

Database State:
==============
[Closed] → [Connect Request] → [Pool Get] → [Query Execute] → [Return Pool] → [Closed]
    │           │                │             │                │             │
    ▼           ▼                ▼             ▼                ▼             ▼
[No Conn] → [Auth Check] → [Active Conn] → [SQL Execute] → [Result Return] → [Disconnect]

Template State:
==============
[Template File] → [Embed Parse] → [Data Bind] → [Markdown Gen] → [Glamour Render] → [Display]
      │              │              │             │                │               │
      ▼              ▼              ▼             ▼                ▼               ▼
[Go Template] → [AST Parse] → [Variable Sub] → [MD String] → [ANSI Convert] → [Terminal Out]
```

### Data Transformation Points

1. **CLI → Flags**: String args to typed values
2. **Config → Objects**: JSON to Go structs
3. **SQL → Results**: DuckDB rows to Note objects
4. **Objects → Templates**: Structs to markdown strings
5. **Markdown → Display**: MD to ANSI terminal output

### Performance Characteristics

- **Memory**: Streaming where possible, minimal buffering
- **Database**: Connection pooling, read-only isolation
- **Templates**: Cached parsing, lazy evaluation
- **File I/O**: Minimal reads, batch operations
- **Network**: None (local-only operations)

### Security Boundaries

- **Input Validation**: At service entry points
- **SQL Injection**: Query parameterization
- **File System**: Path traversal prevention
- **Database**: Read-only connections for queries
- **Template**: Safe template execution context