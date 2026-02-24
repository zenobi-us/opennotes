---
id: 8361d3a2
title: Rename Project from OpenNotes to Jot
type: "epic"
created_at: 2026-02-18T19:41:00+10:30
updated_at: 2026-02-19T09:55:00+10:30
status: in-progress
---

# Rename Project from OpenNotes to Jot

## Vision/Goal

Rebrand the project from "OpenNotes" to "Jot" — a shorter, more memorable, action-oriented name that better captures the essence of quick note-taking from the command line.

**New Identity:**
- **Name**: Jot
- **Binary**: `jot`
- **Module**: `github.com/zenobi-us/jot`
- **Repo**: `github.com/zenobi-us/jot`

## Success Criteria

- [x] All in-repo references updated (module path, imports, binary name, docs)
- [ ] GitHub repository renamed to `jot` [NEEDS-HUMAN]
- [x] Go module path updated and working
- [x] Binary builds as `jot`
- [x] README and documentation reflect new name
- [x] Config directory updated (`.config/jot/` vs `.config/opennotes/`)
- [x] Notebook config file renamed (`.jot.json` vs `.opennotes.json`)
- [x] All tests pass with new naming
- [x] Release/distribution updated (goreleaser, etc.)
- [x] External references identified and migration plan documented

## Phases

1. **Phase 1: Discovery** — ✅ COMPLETE
   - Task: [task-b66d2263](task-b66d2263-identify-rename-locations.md)
   - Found: 47 Go files, 3 constants, 2 env vars, 10 CLI files, 10+ doc files, tests, build config
   - No naming conflicts on Linux; BSD `jot(1)` is documented concern

2. **Phase 2: In-Repo Changes** — ✅ COMPLETE
   - Commit 1 (`bcee714`): Core rename — module, imports, constants, env vars, CLI, schema, build config, tests, docs
   - Commit 2 (`8e332de`): Pi extension package rename (pkgs/pi-opennotes/ → pkgs/pi-jot/)
   - All tests pass, 0 lint issues
   - Branch: `feat/rename-to-jot`

3. **Phase 3: GitHub Rename** — [NEEDS-HUMAN]
   - Merge `feat/rename-to-jot` branch
   - Rename repository `zenobi-us/opennotes` → `zenobi-us/jot`
   - Update local git remotes

4. **Phase 4: External Updates** — TODO after Phase 3
   - Package managers, documentation sites
   - Migration guide for existing users (config paths, notebook files)
   - Regenerate `docs/notebook-discovery.svg`

## Dependencies

- ~~Current feature branch should be merged or rebased before major rename~~ Done on `main`

## Notes

- Name change rationale: "Jot" is short (3 letters), action-oriented (verb), memorable, and perfectly describes the CLI note-taking use case
- Hard rename — no backward compatibility code (users manually migrate configs)
- BSD `jot(1)` conflict: macOS/FreeBSD have a `jot` utility — document in README, not a blocker
- SVG diagram `docs/notebook-discovery.svg` still has embedded "opennotes" text — needs regeneration from d2 source
