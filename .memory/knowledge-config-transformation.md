---
id: config-xform
title: Config Storage to Runtime Transformation
created_at: 2026-03-01T12:15:00+10:30
updated_at: 2026-03-01T12:15:00+10:30
area: architecture
tags: [config, storage, runtime, transformation]
learned_from: [epic-7c631839]
---

# Config Storage to Runtime Transformation

## Overview

Jot has two config scopes (global and notebook) that undergo transformation from storage (JSON files) to runtime (Go structs with resolved values).

## State Machine: Config Lifecycle

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CONFIG LIFECYCLE STATE MACHINE                       │
└─────────────────────────────────────────────────────────────────────────────┘

═══════════════════════════════════════════════════════════════════════════════
                              GLOBAL CONFIG FLOW
═══════════════════════════════════════════════════════════════════════════════

  ┌──────────────────────┐
  │   STORAGE LAYER      │
  │ ~/.config/jot/       │
  │   config.json        │
  └──────────┬───────────┘
             │
             │  os.ReadFile()
             ▼
  ┌──────────────────────┐
  │   RAW JSON           │
  │ {                    │
  │   "notebooks": [...],│
  │   "notebookpath": "" │
  │ }                    │
  └──────────┬───────────┘
             │
             │  koanf.Unmarshal()
             ▼
  ┌──────────────────────┐      ┌─────────────────────────────────────────┐
  │   Config (Storage)   │      │  type Config struct {                   │
  │                      │ ───▶ │    Notebooks    []string  `json:"..."`  │
  │   (No transformation │      │    NotebookPath string    `json:"..."`  │
  │    currently - same  │      │  }                                      │
  │    as runtime!)      │      └─────────────────────────────────────────┘
  └──────────┬───────────┘
             │
             │  ConfigService wraps
             ▼
  ┌──────────────────────┐      ┌─────────────────────────────────────────┐
  │   ConfigService      │      │  type ConfigService struct {            │
  │   (Runtime)          │ ───▶ │    k     *koanf.Koanf                   │
  │                      │      │    Store Config      // embedded        │
  │                      │      │    path  string      // resolved path   │
  │                      │      │  }                                      │
  └──────────────────────┘      └─────────────────────────────────────────┘


═══════════════════════════════════════════════════════════════════════════════
                             NOTEBOOK CONFIG FLOW
═══════════════════════════════════════════════════════════════════════════════

  ┌──────────────────────┐
  │   STORAGE LAYER      │
  │ /project/.jot.json   │
  └──────────┬───────────┘
             │
             │  os.ReadFile()
             ▼
  ┌──────────────────────┐
  │   RAW JSON           │
  │ {                    │
  │   "root": ".notes",  │
  │   "name": "project", │
  │   "groups": [...],   │
  │   "workflows": {...} │
  │ }                    │
  └──────────┬───────────┘
             │
             │  json.Unmarshal()
             ▼
  ┌──────────────────────┐      ┌─────────────────────────────────────────┐
  │ StoredNotebookConfig │      │  type StoredNotebookConfig struct {     │
  │ (STORAGE TYPE)       │ ───▶ │    ConfigVersion Version                │
  │                      │      │    Root           string                │
  │                      │      │    Name           string                │
  │                      │      │    Contexts       []string              │
  │                      │      │    Templates      map[string]string     │
  │                      │      │    Groups         []NotebookGroup       │
  │                      │      │    Workflows      map[string]Workflow   │
  │                      │      │  }                                      │
  └──────────┬───────────┘      └─────────────────────────────────────────┘
             │
             │  LoadConfig() transforms:
             │  - Validates workflow states
             │  - Resolves root path
             │  - Creates root dir if missing
             │  - Adds config file path
             ▼
  ┌──────────────────────┐      ┌─────────────────────────────────────────┐
  │   NotebookConfig     │      │  type NotebookConfig struct {           │
  │   (RUNTIME TYPE)     │ ───▶ │    StoredNotebookConfig  // embedded    │
  │                      │      │    Path string `json:"-"` // RUNTIME    │
  │                      │      │  }                                      │
  └──────────┬───────────┘      └─────────────────────────────────────────┘
             │
             │  Open() creates services
             ▼
  ┌──────────────────────┐      ┌─────────────────────────────────────────┐
  │     Notebook         │      │  type Notebook struct {                 │
  │   (FULL RUNTIME)     │ ───▶ │    Config NotebookConfig                │
  │                      │      │    Notes  *NoteService  // runtime svc  │
  │                      │      │  }                                      │
  └──────────────────────┘      └─────────────────────────────────────────┘


═══════════════════════════════════════════════════════════════════════════════
                            SCHEMA GENERATION TARGET
═══════════════════════════════════════════════════════════════════════════════

  ┌─────────────────────────────────────────────────────────────────────────┐
  │                                                                         │
  │   JSON SCHEMA should be generated from STORAGE TYPES only:              │
  │                                                                         │
  │   ┌─────────────────────┐         ┌─────────────────────┐               │
  │   │ StoredNotebookConfig│ ──────▶ │  jot.schema.json    │               │
  │   │ NotebookGroup       │         │  (Draft-07)         │               │
  │   │ NotebookGroupWF     │         │                     │               │
  │   │ WorkflowDefinition  │         │  Validates:         │               │
  │   │ WorkflowState       │         │  - .jot.json        │               │
  │   └─────────────────────┘         └─────────────────────┘               │
  │                                                                         │
  │   NOT from runtime types (NotebookConfig, Notebook, ConfigService)      │
  │                                                                         │
  └─────────────────────────────────────────────────────────────────────────┘


═══════════════════════════════════════════════════════════════════════════════
                              CURRENT GAPS
═══════════════════════════════════════════════════════════════════════════════

  1. Global Config has NO storage/runtime separation
     - Config struct serves both purposes
     - No StoredConfig type exists

  2. Runtime-only fields use `json:"-"` but are on same struct
     - Works but muddles the conceptual boundary

  3. Schema generation needs clear "storage types" package/grouping
     - Currently mixed with runtime types in notebook.go
```

## Current Type Inventory

### Storage Types (Schema Generation Targets)

| Type | File | Purpose |
|------|------|---------|
| `StoredNotebookConfig` | `notebook.go` | Root notebook config |
| `NotebookGroup` | `notebook.go` | Group definition |
| `NotebookGroupWorkflow` | `notebook.go` | Group→workflow binding |
| `WorkflowDefinition` | `workflow_validation.go` | Workflow spec |
| `WorkflowState` | `workflow_validation.go` | State transitions |
| `Config` | `config.go` | Global config (⚠️ no separation) |

### Runtime Types (NOT for schema)

| Type | File | Purpose |
|------|------|---------|
| `NotebookConfig` | `notebook.go` | Storage + resolved paths |
| `Notebook` | `notebook.go` | Full runtime with services |
| `ConfigService` | `config.go` | Global config manager |
| `NotebookService` | `notebook.go` | Notebook manager |

## Recommendations

1. **Create `internal/services/schema/` package** for storage-only types
2. **Add `StoredGlobalConfig`** to separate global config storage from runtime
3. **Schema generator** should only reflect types from schema package
4. **Runtime types** embed storage types + add runtime fields
