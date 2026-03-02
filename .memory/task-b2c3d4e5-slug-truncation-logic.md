---
id: b2c3d4e5
title: Slug Truncation Logic
created_at: 2026-03-02T17:47:00+10:30
updated_at: 2026-03-02T17:47:00+10:30
status: todo
epic_id: c5d7e9b1
phase_id: 1
story_id: g2b3c4d5
assigned_to: null
---

# Slug Truncation Logic

## Objective

Implement sensible truncation for long titles, respecting word boundaries where possible, with a configurable maximum length (default 50 chars).

## Related Story

- [story-g2b3c4d5](story-g2b3c4d5-unicode-safe-slugify.md) — Unicode-safe Slugify
- Directly implements AC#5 (Long titles truncated sensibly)

## Related Phase

- **Phase 1: Foundation** in [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md)
- Depends on task-a1b2c3d4 (slug library integration)

## Steps

1. Extend `Slug()` function signature to accept optional max length:
   ```go
   func Slug(title string) string              // uses default 50
   func SlugWithMax(title string, max int) string
   ```

2. Implement word-boundary truncation:
   ```go
   func truncateAtWordBoundary(s string, max int) string {
       if len(s) <= max {
           return s
       }
       // Find last hyphen before max
       truncated := s[:max]
       if lastHyphen := strings.LastIndex(truncated, "-"); lastHyphen > max/2 {
           return truncated[:lastHyphen]
       }
       return truncated
   }
   ```

3. Handle edge cases:
   - Title shorter than max → return as-is
   - No word boundary found → truncate at max (hard cut)
   - Single very long word → truncate at max

4. Update existing slug calls if any

## Unit Tests

- `TestSlug_ShortTitle`: "hello" (5 chars) → "hello" → supports AC#5
- `TestSlug_ExactlyMax`: 50-char title → returns unchanged → supports AC#5
- `TestSlug_LongAtWordBoundary`: 60-char title → truncates at word → supports AC#5
- `TestSlug_LongSingleWord`: 60-char single word → hard truncate at 50 → supports AC#5

## Expected Outcome

Long titles produce reasonable filenames that don't exceed filesystem limits while remaining readable and meaningful.

## Actual Outcome

_To be filled after completion_

## Lessons Learned

_To be filled after completion_
