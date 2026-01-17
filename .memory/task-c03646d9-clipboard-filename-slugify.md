# Bug: Clipboard Filename Should Be Slugified

**Status**: 🔴 NEW BUG REPORTED  
**Reported**: 2026-01-17 20:25  
**Priority**: HIGH (User-facing filename issue)  
**Component**: Clipboard/Temp File Naming  

## Issue Description

When clipboard content is saved to a temporary file, the filename uses UUID format instead of a human-readable slug format.

### Current Behavior
```
/tmp/pi-clipboard-7fe7cacf-9019-46a1-819b-b59854dd883f.png
```

### Expected Behavior
```
/tmp/pi-clipboard-<slugified-name>.png
```

## Root Cause Analysis

The temp filename generation uses raw UUID format without slug conversion. This applies to any clipboard content saved by the pi system.

## Impact

- **Severity**: MEDIUM
- **User Impact**: Unhelpful filenames in /tmp
- **Related Components**: Clipboard handling, temp file creation
- **Blocking**: NO

## Steps to Reproduce

1. User provides image via clipboard (e.g., screenshot)
2. System saves to `/tmp/pi-clipboard-<uuid>.png`
3. Filename is not human-readable slug format

## Acceptance Criteria

- [ ] Temp filename uses slugified format (lowercase, hyphens, no special chars)
- [ ] Works for all file types (images, text, etc.)
- [ ] Maintains uniqueness (no collisions)
- [ ] Changes are backwards compatible
- [ ] Filename follows kebab-case convention

## Solution Approach

This is a **pi system** issue (not OpenNotes specific).

The filename generation logic needs to:
1. Extract a meaningful base name from content type or context
2. Slugify it (convert to lowercase, replace spaces/special chars with hyphens)
3. Append UUID segment for uniqueness
4. Add file extension based on content type

### Example:
- Current: `pi-clipboard-7fe7cacf-9019-46a1-819b-b59854dd883f.png`
- Better: `pi-clipboard-user-request-7fe7cacf.png`

## Next Steps

1. Identify where pi generates temp clipboard filenames
2. Review current implementation
3. Design slug generation strategy
4. Implement and test
5. Verify backwards compatibility

## Dependencies

- Understanding of pi clipboard system
- Access to pi source code
- May need coordination with other pi developers

## Notes

This is a quality-of-life improvement that will make system debugging easier and filenames more user-friendly.
