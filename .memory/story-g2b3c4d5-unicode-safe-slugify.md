---
id: g2b3c4d5
title: Unicode-safe Slugify
created_at: 2026-03-02T13:43:00+10:30
updated_at: 2026-03-02T13:43:00+10:30
status: proposed
epic_id: c5d7e9b1
priority: high
story_points: 3
test_coverage: none
---

# Unicode-safe Slugify

## User Story

**As a** Jot user  
**I want** my note titles with unicode/emoji to generate readable filenames  
**So that** `"会议 Notes"` becomes `hui-yi-notes.md` not `notes.md`

## Acceptance Criteria

- [ ] Chinese characters transliterated (会议 → hui-yi)
- [ ] Emoji stripped gracefully
- [ ] Accented chars normalized (café → cafe)
- [ ] Existing ASCII behavior unchanged
- [ ] Long titles truncated sensibly (50 char default?)

## Context

The current slugify implementation strips non-ASCII characters, resulting in empty or nonsensical filenames for international users. This story ensures filenames remain meaningful across all languages by transliterating unicode characters to their closest ASCII equivalents.

## Out of Scope

- Configurable transliteration rules per notebook
- Preserving emoji in filenames (filesystem compatibility issues)
- Bidirectional text handling (RTL languages)

## Tasks

- [task-a1b2c3d4](task-a1b2c3d4-integrate-slug-library.md) — Integrate gosimple/slug library (Phase 1)
- [task-b2c3d4e5](task-b2c3d4e5-slug-truncation-logic.md) — Slug truncation logic (Phase 1)

## Test Specification

### E2E Tests

| AC# | Criterion | Test file/case | Status |
|-----|-----------|----------------|--------|
| 1 | Chinese characters transliterated | slugify_test.go / TestChineseTransliteration | ⬜ |
| 2 | Emoji stripped gracefully | slugify_test.go / TestEmojiStripping | ⬜ |
| 3 | Accented chars normalized | slugify_test.go / TestAccentNormalization | ⬜ |
| 4 | Existing ASCII behavior unchanged | slugify_test.go / TestASCIIPreserved | ⬜ |
| 5 | Long titles truncated sensibly | slugify_test.go / TestLongTitleTruncation | ⬜ |

### Unit Test Coverage (via Tasks)

_To be populated as tasks are completed_

## Notes

- Recommended library: `github.com/mozillazg/go-unidecode` or similar
- Truncation should happen at word boundaries when possible
- Empty result after slugify should have defined fallback (e.g., "untitled" or nanoid)
- Edge cases: all-emoji titles, mixing CJK + Latin, combining characters
