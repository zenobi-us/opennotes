---
id: a9b8c7d6
title: Workflow discovery brief
epic_id: b2f4e6a8
related_task_id: d1e2f3a4
created_at: 2026-02-28T01:06:00+10:30
updated_at: 2026-02-28T01:22:00+10:30
status: in-progress
---

# Workflow discovery brief

## Research Questions
1. What rollout path should workflow capability take in V1?
2. Which note-flow should anchor design and validation semantics?
3. Where should workflows live and how should transitions/schema be expressed?
4. What are V1 non-goals and constraints?

## Summary
Discovery decisions establish a V1 foundation centered on validation-governed notebook workflows. The chosen path is hybrid sequencing: model all candidate stories but execute governance/safety first, then execution engine capabilities. Workflow definitions are notebook-scoped in `.jot.json` as keyed objects. Each state defines full JSON Schema requirements and explicit allowed transitions.

## Discovery Q&A (verbatim with options)

### Pre-Q1 clarification
- **User:** "i cant answer this without example user stories for those four."
- **User:** "i think these are all desirable stories to model.

We should write stories for them all, but yes. start with D"

### Q1
- **Question:** "What is the first workflow outcome we should optimize for in Jot?"
- **Options:**
  - **A)** Authoring governance
  - **B)** Execution orchestration
  - **C)** Agent safety/compliance
  - **D)** Hybrid (A + C first, B later)
- **Answer:** "start with D"

### Q2
- **Question:** "What is the first concrete note type/state flow to anchor V1 on?"
- **Options:**
  - **A)** Task notes: `draft -> ready -> doing -> done -> archived`
  - **B)** Project/story notes: `proposed -> planned -> in-progress -> completed`
  - **C)** Meeting/decision notes: `captured -> reviewed -> approved -> published`
  - **D)** Custom flow
- **Answer:** "no lets to B. because from be we derive A"

### Q3
- **Question:** "On invalid transitions/metadata in V1, what enforcement mode should be default?"
- **Options:**
  - **A)** Hard block
  - **B)** Warn-only
  - **C)** Configurable per notebook (`enforce|warn`)
  - **D)** Configurable per workflow/state
- **Answer:** "D. because workflows are per notebook."

### Q4
- **Question:** "Where should workflow definitions live for V1?"
- **Options:**
  - **A)** `.jot.json` only
  - **B)** Dedicated file per notebook (e.g. `.jot/workflows.yaml`)
  - **C)** Notes-as-config
  - **D)** Hybrid file + note fragments
- **Answer:** "A"

### Q5
- **Question:** "How should inheritance/override work in V1 for project/story workflows?"
- **Options:**
  - **A)** No inheritance
  - **B)** Global defaults + per-workflow overrides
  - **C)** Base workflow + extends/override
  - **D)** Two-level only (notebook defaults + state overrides)
- **Answer:** "A"

### Q6
- **Question:** "What should V1 failure output prioritize?"
- **Options:**
  - **A)** Human-readable CLI messages only
  - **B)** Machine-readable JSON only
  - **C)** Both JSON + concise human summary
  - **D)** Verbosity levels with stable JSON shape
- **Answer:** "as much information as required, but in a format that matches --format arg supplied"

### Q7
- **Question:** "For story/project flow (`proposed -> planned -> in-progress -> completed`), what are the minimum required metadata fields per state in V1?"
- **Options:**
  - **A)**
    - proposed: `title`
    - planned: `title, owner, priority`
    - in-progress: `title, owner, priority, started_at`
    - completed: `title, owner, completed_at`
  - **B)**
    - proposed: `title, epic_id`
    - planned: `title, epic_id, priority`
    - in-progress: `title, epic_id, priority, assignee`
    - completed: `title, epic_id, completed_at, outcome`
  - **C)** Same required fields across all states
  - **D)** Custom required fields per state
- **Answer:** "b"

### Q8
- **Question:** "Are transitions strictly linear in V1, or can states move backward?"
- **Options:**
  - **A)** Strict forward only
  - **B)** Rollback one step
  - **C)** Reopen from completed
  - **D)** Fully configurable transition graph per workflow
- **Answer:** "valid  next state transitions are defined in the workflow step.

> { "state": "proposed", "schema" { ... }, "transitions": ["planned", "proposed", "cancelled"] }

Is just one example"

### Q9
- **Question:** "What schema language should `schema` use in V1?"
- **Options:**
  - **A)** JSON Schema subset
  - **B)** Full JSON Schema
  - **C)** Custom lightweight Jot schema
  - **D)** Hybrid (Jot now, JSON Schema later)
- **Answer:** "B"

### Q10
- **Question:** "How should workflows be identified/invoked in `.jot.json`?"
- **Options:**
  - **A)** Keyed object (`workflows.project_story`)
  - **B)** Array with explicit IDs (`[{ id: "project_story", ... }]`)
  - **C)** Both accepted, normalize internally
  - **D)** Custom shape
- **Answer:** "i'm not sure we care about any of those reasons. lets use A. keyed objects."

### Q11
- **Question:** "Pick V1 non-goals (A-F)."
- **Options:**
  - **A)** Cross-notebook workflows
  - **B)** Time/schedule triggers
  - **C)** Parallel/branching step execution
  - **D)** External side-effect actions (HTTP/webhooks/shell)
  - **E)** Workflow version migration tooling
  - **F)** UI/TUI workflow editor
- **Answer:** "All except E.

We definitely want to utilise our existing migration framework for the initial version"

## Findings
- **Execution path:** Start with option **D** (governance + agent safety first, execution staged next).
- **Use-case priority (P0/P1/P2):**
  - **P0:** Authoring governance (metadata/state validation during edits/transitions)
  - **P0:** Agent safety/compliance (pre-write policy validation)
  - **P1:** Execution orchestration (`dry-run`/`apply`, machine-readable results)
  - **P2:** State rollback/reopen patterns (explicit backward transitions where needed)
- **Anchor flow:** `proposed -> planned -> in-progress -> completed` for project/story notes.
- **Enforcement model:** Defined at workflow-step level.
- **Storage model:** Workflow config in `.jot.json` (keyed objects).
- **State model:** Each step includes:
  - `state`
  - `schema` (full JSON Schema)
  - `transitions` (explicit allowed next states)
- **Output requirement:** Diagnostics must preserve required detail and render in the format requested by `--format`.
- **Non-goals for V1:**
  - Cross-notebook workflows
  - Time/schedule triggers
  - Parallel/branching execution
  - External side-effect actions (HTTP/webhook/shell)
  - UI/TUI workflow editor
- **Migration constraint (explicit):** No compromise path needed. Workflows are a new `.jot.json` schema section and are covered by the existing config migration system from initial release.
- **Explicit inclusion:** Reuse existing migration framework from initial version (therefore migration concerns are not excluded from V1 planning).

## References
- Interactive discovery session with project owner (Q), 2026-02-28.
