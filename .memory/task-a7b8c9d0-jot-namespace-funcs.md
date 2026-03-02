---
id: a7b8c9d0
title: Implement jot Namespace Functions
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: done
epic_id: c5d7e9b1
phase_id: 2
story_id: h3c4d5e6
assigned_to: null
---

# Implement jot Namespace Functions

## Objective

Create the `jot` namespace with all required functions: Slug, NanoID, Timestamp, DatePath, UUID, and Now.

## Related Story

- [story-h3c4d5e6](story-h3c4d5e6-group-filename-patterns.md) — Group Filename Patterns
- Directly implements AC#3 (Available functions: `jot.Slug`, `jot.NanoID`, `jot.Timestamp`, `jot.DatePath`, `jot.UUID`, `jot.Now`)

## Related Phase

- **Phase 2: Schema & Resolution** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Depends on task-f6a7b8c9 (template engine exists)
- Depends on task-a1b2c3d4 (Slug function exists)

## Steps

1. Create `internal/services/template_funcs.go`:
   ```go
   package services

   import (
       "time"
       "github.com/google/uuid"
       "github.com/jaevor/go-nanoid"
   )

   // JotFuncs returns the jot namespace function map
   func JotFuncs() map[string]interface{} {
       return map[string]interface{}{
           "Slug":      Slug,
           "NanoID":    NanoID,
           "Timestamp": Timestamp,
           "DatePath":  DatePath,
           "UUID":      UUID,
           "Now":       Now,
       }
   }
   ```

2. Implement each function:

   ```go
   // NanoID generates a nanoid of specified length
   func NanoID(length int) string {
       gen, _ := nanoid.Standard(length)
       return gen()
   }

   // Timestamp returns Unix timestamp
   func Timestamp() int64 {
       return time.Now().Unix()
   }

   // DatePath returns date as path segment (e.g., "2026/03/02")
   func DatePath() string {
       return time.Now().Format("2006/01/02")
   }

   // UUID returns a UUID v4
   func UUID() string {
       return uuid.New().String()
   }

   // Now returns formatted current time
   func Now(format string) string {
       return time.Now().Format(format)
   }
   ```

3. Register jot namespace in template engine:
   ```go
   te.funcs["jot"] = JotFuncs()
   ```

4. Add nanoid and uuid dependencies:
   ```bash
   go get github.com/jaevor/go-nanoid
   go get github.com/google/uuid
   ```

## Unit Tests

- `TestJotSlug`: `{{ jot.Slug "Hello World" }}` → "hello-world" → supports AC#3
- `TestJotNanoID`: `{{ jot.NanoID 8 }}` → 8-char string → supports AC#3
- `TestJotTimestamp`: `{{ jot.Timestamp }}` → unix timestamp → supports AC#3
- `TestJotDatePath`: `{{ jot.DatePath }}` → "2026/03/02" → supports AC#3
- `TestJotUUID`: `{{ jot.UUID }}` → valid UUID format → supports AC#3
- `TestJotNow`: `{{ jot.Now "2006-01-02" }}` → "2026-03-02" → supports AC#3

## Expected Outcome

Users can use `{{ jot.NanoID 8 }}` and similar functions in their filename_format and template fields.

## Actual Outcome

✅ Successfully implemented jot namespace functions:

- Created `internal/services/template_funcs.go` with `JotNamespace` struct
- Implemented all 6 methods: `Slug`, `NanoID`, `Timestamp`, `DatePath`, `UUID`, `Now`
- Registered `jot` namespace in template engine
- Added `github.com/jaevor/go-nanoid v1.4.0` dependency
- Created comprehensive test suite (33 new tests)
- Template syntax works: `{{ jot.Slug "text" }}` and `{{ .title | jot.Slug }}`
- All 359 tests pass

## Lessons Learned

- Go templates support method calls on struct values, enabling namespace-like syntax (`jot.Method`)
- NanoID with URL-safe alphabet is ideal for filenames - no special characters to escape
