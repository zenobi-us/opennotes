# Getting Started Troubleshooting

Common issues and fast fixes for Jot.

## Quick triage

1. `jot` command missing → [CLI issues](#cli-issues)
2. Notebook not found/resolution confusion → [Notebook issues](#notebook-issues)
3. Search results unexpected → [Search issues](#search-issues)
4. Semantic search behavior confusing → [semantic-issues](#semantic-issues)
5. Views or automation scripts failing → [views--automation-issues](#views--automation-issues)

---

## CLI issues

### `command not found: jot`

```bash
go install github.com/zenobi-us/jot@latest
jot --version
```

Ensure your Go bin path is on `PATH`.

### Wrong binary/version

```bash
which jot
jot --version
```

---

## Notebook issues

### `no notebook found`

```bash
jot notebook list
jot notebook create "My Notes" --path ~/notes
```

You can also set explicit context:

```bash
export JOT_NOTEBOOK=~/notes
jot notes list
```

### Notebook resolution seems wrong

Check precedence using docs:

- [notebook-discovery.md](notebook-discovery.md)

Typical fixes:

- clear stale env var (`unset JOT_NOTEBOOK`)
- pass `--notebook` explicitly
- run from the intended notebook directory

---

## Search issues

### No results for expected notes

```bash
jot notes list
jot notes search "keyword"        # fieldless => title search
jot notes search "body:keyword"   # explicit body/content search
```

Try structured filters to verify metadata assumptions:

```bash
jot notes search query --and data.tag=project
```

### Invalid query condition

Use `field=value` with supported fields only.

Good:

```bash
jot notes search query --and data.status=active --not data.priority=low
```

---

## Semantic issues

### Semantic mode gives different results than keyword search

Expected behavior. Use hybrid as default:

```bash
jot notes search semantic "topic" --mode hybrid --explain
```

### Semantic backend unavailable warning

Fallback strategy:

```bash
jot notes search semantic "topic" --mode keyword
```

See full guide:

- [semantic-search-guide.md](semantic-search-guide.md)

---

## Views & automation issues

### View not found

```bash
jot notes view --list
```

Verify view location (`.jot.json` vs global config) and naming.

### Script works locally but fails in CI

Use explicit notebook path and deterministic shell env:

```bash
JOT_NOTEBOOK=/abs/path/to/notebook jot notes view recent --format json
```

### JSON parsing fails

Make sure command actually outputs JSON (e.g. `notes view --format json`).

---

## Migration issues (OpenNotes → Jot)

Use:

- [migration-opennotes-to-jot.md](migration-opennotes-to-jot.md)

Quick checks:

```bash
jot notebook migrate list
jot notebook migrate --apply
```

---

## Still blocked?

Capture:

- exact command
- exact error output
- output of `jot --version`
- output of `jot notebook list`

Then open an issue at: https://github.com/zenobi-us/jot
