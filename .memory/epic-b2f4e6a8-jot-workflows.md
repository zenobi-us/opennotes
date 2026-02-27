---
id: b2f4e6a8
title: Jot Workflows
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-02-27T08:01:43+10:30
status: proposed
---

# Jot Workflows

## Vision/Goal
Introduce native Jot workflows to express multi-step operational flows, metadata-based routing, and validation gates beyond the current basic miniproject flow chart.

## Success Criteria
- Workflow definitions are first-class and executable by Jot.
- Workflows can validate note metadata, required transitions, and completion criteria.
- Workflows provide actionable status and clear failures for automation/agents.
- Workflow definitions are composable and notebook-scoped.

## Phases
- Discovery/specification of workflow model.
- Metadata validation and transition design.
- Execution engine + status/reporting interface.

## Dependencies
- Search/query DSL and views.
- Metadata conventions across note groups/types.
- Decision on workflow definition format (e.g., YAML in config vs notes).
