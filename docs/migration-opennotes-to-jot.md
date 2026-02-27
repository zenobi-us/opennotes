# OpenNotes → Jot Migration Guide

This project was renamed from **OpenNotes** to **Jot**.

If you previously used `opennotes`, this guide explains what changed and how to migrate safely.

## What Changed

| Area                         | Old                                     | New                               |
| ---------------------------- | --------------------------------------- | --------------------------------- |
| Repository                   | `github.com/zenobi-us/opennotes`        | `github.com/zenobi-us/jot`        |
| CLI binary                   | `opennotes`                             | `jot`                             |
| Go install path              | `github.com/zenobi-us/opennotes@latest` | `github.com/zenobi-us/jot@latest` |
| Go module path               | `github.com/zenobi-us/opennotes`        | `github.com/zenobi-us/jot`        |
| Notebook config filename     | `.opennotes.json`                       | `.jot.json`                       |
| Env var (notebook selection) | `OPENNOTES_NOTEBOOK`                    | `JOT_NOTEBOOK`                    |

---

## 1) Update Installation

```bash
go install github.com/zenobi-us/jot@latest
```

Verify:

```bash
jot --version
```

---

## 2) Run Built-in Migration Checks

Jot includes notebook/config migration tooling.

```bash
# Show migration status
jot notebook migrate list

# Dry run (no changes)
jot notebook migrate

# Apply migrations
jot notebook migrate --apply
```

This handles built-in migration steps such as legacy notebook config rename where applicable.

---

## 3) Update Shell Scripts and Automation

Replace command invocations:

```bash
# Before
opennotes notes list

# After
jot notes list
```

Replace env vars:

```bash
# Before
export OPENNOTES_NOTEBOOK=~/notes

# After
export JOT_NOTEBOOK=~/notes
```

---

## 4) Update Go Imports (if embedding/integrating)

```go
// Before
import "github.com/zenobi-us/opennotes/..."

// After
import "github.com/zenobi-us/jot/..."
```

And in `go.mod`:

```go
module github.com/zenobi-us/jot
```

---

## 5) Search Behavior Notes

Current search surfaces are:

- `jot notes search "..."` (fieldless title-only; normalized to `title:...`)
- `jot notes search "body:..."` (explicit body/content search)
- `jot notes search "..."` (DSL filters/boolean operators)
- `jot notes search semantic "..."` (semantic/hybrid)

If you are migrating old workflows, verify command docs:

- [Notes Search Command](commands/notes-search.md)
- [Semantic Search Guide](semantic-search-guide.md)

---

## 6) Quick Validation Checklist

- [ ] `gh repo view` shows `zenobi-us/jot`
- [ ] `jot --version` works
- [ ] `jot notebook migrate list` runs cleanly
- [ ] Notebook uses `.jot.json`
- [ ] CI/scripts no longer reference `opennotes`
- [ ] Env vars use `JOT_NOTEBOOK`

---

## Troubleshooting

### `jot` command not found

Your old binary may still be on PATH, or Go bin path is missing.

- Re-run install command
- Ensure `$(go env GOPATH)/bin` (or configured Go bin) is on `PATH`

### Notebook not detected

- Confirm `.jot.json` exists in notebook root
- Set `JOT_NOTEBOOK` explicitly
- Run `jot notebook list` and `jot notebook migrate list`

### Legacy references still in shell profiles

Search and replace in shell config files:

- `~/.bashrc`
- `~/.zshrc`
- `~/.config/fish/config.fish`

Replace:

- `opennotes` → `jot`
- `OPENNOTES_` → `JOT_`
