---
id: a1b2c3d4
title: Integrate gosimple/slug Library
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: todo
epic_id: c5d7e9b1
phase_id: 1
story_id: g2b3c4d5
assigned_to: null
---

# Integrate gosimple/slug Library

## Objective

Add the `github.com/gosimple/slug` library and create a wrapper function that provides unicode-safe slugification for note titles.

## Related Story

- [story-g2b3c4d5](story-g2b3c4d5-unicode-safe-slugify.md) — Unicode-safe Slugify
- Contributes to AC#1 (Chinese transliteration), AC#2 (emoji stripping), AC#3 (accent normalization), AC#4 (ASCII preservation)

## Related Phase

- **Phase 1: Foundation** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Core dependency that filename patterns (Phase 2) will build upon

## Steps

1. Add `github.com/gosimple/slug` to `go.mod`:
   ```bash
   go get github.com/gosimple/slug
   ```

2. Create `internal/services/slug.go` with wrapper function:
   ```go
   package services

   import "github.com/gosimple/slug"

   // Slug converts a title to a URL-safe, filesystem-safe slug
   func Slug(title string) string {
       return slug.Make(title)
   }
   ```

3. Configure slug library for optimal unicode handling:
   - Set `slug.Lowercase = true`
   - Set `slug.MaxLength = 50` (truncation)
   - Configure custom substitutions if needed

4. Add fallback for empty results (all-emoji titles):
   ```go
   if result == "" {
       result = "untitled"
   }
   ```

5. Run tests to verify no regressions in existing note creation

## Unit Tests

- `TestSlug_ChineseCharacters`: 会议 → hui-yi → supports AC#1
- `TestSlug_Emoji`: 🎉 Party → party → supports AC#2
- `TestSlug_Accents`: café → cafe → supports AC#3
- `TestSlug_ASCII`: hello-world → hello-world → supports AC#4
- `TestSlug_EmptyFallback`: 🎉🎊🎈 → untitled → supports AC#2

## Expected Outcome

A `Slug()` function exists that can be called throughout the codebase, correctly handling unicode input and returning filesystem-safe slugs.

## Actual Outcome

_To be filled after completion_

## Lessons Learned

_To be filled after completion_
