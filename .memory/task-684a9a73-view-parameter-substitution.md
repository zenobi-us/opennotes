---
id: 684a9a73
title: Implement named view parameter substitution
type: "task"
created_at: 2026-02-20T19:13:00+10:30
updated_at: 2026-02-22T21:21:00+10:30
status: done
epic_id: f661c068
phase_id: 4adb81db
assigned_to: 3e30a6e5-e7ec-404d-b31b-82946451beb0
---

# Implement named view parameter substitution

## Objective

Add `{{param_name}}` substitution for saved views so view definitions can be reused with runtime values through `--param`.

## Related Story

- Plan deferred item #4 in [plan-b4e2f7a1-dsl-views-implementation.md](plan-b4e2f7a1-dsl-views-implementation.md)
- API/docs context in `docs/views-api.md` and `docs/views-guide.md`

## Steps

1. Confirm parameter schema contract (types, required/default behavior).
2. Implement substitution pipeline before query parse/execute.
3. Validate provided params against definition with clear errors.
4. Add tests for required params, defaults, type mismatches, and unknown params.
5. Ensure template variable support (`{{today}}`) remains compatible.
6. Document parameterized view examples and troubleshooting.

## Expected Outcome

Users can define parameterized views and execute with:
- `jot notes view by-author --param author="Alice"`

with deterministic validation and rendering.

## Actual Outcome

Completed.

### Execution order fix (`internal/services/view_executor.go`)
- Added pipeline: runtime params → defaults → validation → named substitution → template resolution.
- Added `resolveQueryWithParameters()` helper to centralize behavior.
- Validation now runs during execution and surfaces required/unknown/type errors before query parsing.

### Type compatibility fix (`internal/services/view.go`)
- Updated date param validation to accept template tokens like `{{today-30}}`.

### Test coverage
- Expanded `TestViewService_ExecuteView_ParameterSubstitution` to cover:
  - required parameter enforcement
  - default parameter application
  - unknown parameter rejection
  - invalid type rejection
  - template parameter values
  - template defaults for date params
- Added unit check for template date acceptance in `TestViewService_ValidateParamType_Date`.

### Docs
- Updated parameter resolution order in `docs/views-api.md`.
- Updated parameter validation section in `docs/views-guide.md` to reflect unknown-param checks, template date values, and execution order.

### Verification
- `mise run test` passed across repo.

## Lessons Learned

- Resolving template variables too early causes parameter placeholders to be treated as template vars and left unresolved.
- Validation must execute on merged params (runtime + defaults), not runtime-only input.
- Date-type validation needs explicit allowance for template tokens to preserve documented behavior.
