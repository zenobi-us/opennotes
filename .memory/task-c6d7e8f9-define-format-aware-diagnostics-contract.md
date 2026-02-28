---
id: c6d7e8f9
title: Define format aware diagnostics contract
epic_id: b2f4e6a8
story_id: 7a8b9c0d
created_at: 2026-02-28T01:32:00+10:30
updated_at: 2026-02-28T13:12:00+10:30
status: completed
assigned_to: "3219612528192551"
---

# Define format aware diagnostics contract

## Objective
Specify the diagnostics payload and rendering contract so output detail is preserved while matching `--format`.

## Related Story
[story-7a8b9c0d](story-7a8b9c0d-workflow-spec-metadata-dsl.md)

## Steps
1. Define canonical diagnostics payload fields.
2. Specify `--format` mapping rules for text/json/table outputs.
3. Define required actionable detail for validation and transition failures.

## Expected Outcome
A stable diagnostics contract for agent and human consumption.

## Actual Outcome
Diagnostics contract finalized for workflow V1.

### Canonical Diagnostics Payload (single diagnostic item)
```json
{
  "version": "workflow-diagnostics/v1",
  "timestamp": "RFC3339",
  "severity": "error|warning",
  "mode": "enforce|warn",
  "operation": "config-validate|transition-check|metadata-validate",
  "code": "WF_*",
  "message": "human-readable summary",
  "workflow": "<workflow_key>",
  "note": {
    "path": "<relative_note_path>",
    "id": "<optional_note_id>"
  },
  "state": {
    "current": "<state_or_empty>",
    "requested": "<state_or_empty>",
    "initial": "<initial_state_from_config>"
  },
  "location": {
    "config_file": ".jot.json",
    "json_pointer": "</workflows/...>",
    "schema_pointer": "<optional_schema_pointer>"
  },
  "details": {
    "allowed_targets": ["..."],
    "missing_fields": ["..."],
    "unknown_fields": ["..."],
    "validator": "<optional_json_schema_keyword>",
    "expected": "<optional_expected>",
    "actual": "<optional_actual>"
  },
  "hints": ["actionable next step"],
  "correlation_id": "<uuid_or_request_id>"
}
```

### Canonical Error Codes (V1)
- `WF_INITIAL_STATE_UNKNOWN`
- `WF_STATE_UNKNOWN`
- `WF_TRANSITION_NOT_ALLOWED`
- `WF_TRANSITION_TARGET_UNKNOWN`
- `WF_SCHEMA_VALIDATION_FAILED`
- `WF_CONFIG_UNKNOWN_FIELD`
- `WF_CONFIG_MISSING_REQUIRED`

### Renderer Mapping Rules (`--format`)
- **JSON**: emit canonical payload unchanged (full fidelity, no omitted fields).
- **Text**: render compact summary line first, then explicit key/value blocks for all fields not shown in summary (including arrays and pointers).
- **Table**: render primary columns (`severity`, `code`, `workflow`, `state.current`, `state.requested`, `message`) plus a `details_json` column containing serialized remainder of payload to guarantee no information loss.
- **Multi-diagnostic output**: preserve evaluation order; do not reorder by severity.
- **Warn mode behavior**: severity may be `warning`, but payload shape is identical.

### Required Actionable Detail by Failure Type
- **Transition failure (`WF_TRANSITION_NOT_ALLOWED`)** must include:
  - `state.current`, `state.requested`
  - `details.allowed_targets`
  - `location.json_pointer` to transition list
  - at least one hint (e.g., "add target to transitions" or "request valid state")
- **Schema failure (`WF_SCHEMA_VALIDATION_FAILED`)** must include:
  - failing field pointer (`location.schema_pointer` or `location.json_pointer`)
  - `details.validator`, `details.expected`, `details.actual` when available
  - hint showing compliant shape example
- **Config structural failures** must include:
  - exact unknown/missing keys
  - config pointer location

## Decision Log (Questions, Options, Answers)

### Q1: What is the source of truth for diagnostics?
**Options presented:**
- **A)** Separate payloads per output format
- **B)** One canonical payload, renderers map from it
- **C)** Human-oriented payload only, infer machine details downstream

**Answer:**
- **B**

**Recorded interpretation:**
- A single canonical payload is mandatory to avoid format drift.

### Q2: How do we guarantee no detail loss in `table` output?
**Options presented:**
- **A)** Accept partial detail in table mode
- **B)** Add expandable multiline sections after table rows
- **C)** Include `details_json` column carrying remaining structured fields

**Answer:**
- **C**

**Recorded interpretation:**
- Table remains compact but still lossless via serialized payload remainder.

### Q3: Should warn mode use a different payload schema?
**Options presented:**
- **A)** Yes, lighter warning-specific schema
- **B)** No, same schema; only severity/mode differ

**Answer:**
- **B**

**Recorded interpretation:**
- Keep one schema for deterministic automation and easier testing.

### Q4: Which diagnostics fields are mandatory for transition failures?
**Options presented:**
- **A)** Message + code only
- **B)** Message + code + current/requested/allowed + location + hint
- **C)** Message + code + free-form details map only

**Answer:**
- **B**

**Recorded interpretation:**
- Transition failures must be immediately actionable without extra lookups.

## Lessons Learned
A single canonical payload plus explicit renderer rules prevents accidental information loss and keeps machine + human consumers aligned.
