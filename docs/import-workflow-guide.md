# Import Workflow Guide

Use this guide to bring existing markdown collections into Jot safely.

## Supported source patterns

- Single markdown folder
- Nested project trees
- Existing knowledge-base repos
- Notes migrated from other tools (Obsidian, Bear, etc.)

## 1) Create notebook from existing notes

```bash
jot notebook create "Imported Notes" --path ~/my-notes
jot notes list
```

## 2) Validate notebook detection

```bash
jot notebook list
jot notes list
```

If running from multiple contexts, review:

- [notebook-discovery.md](notebook-discovery.md)

## 3) Validate searchability

```bash
jot notes search "todo"          # fieldless => title search
jot notes search "body:todo"     # explicit body/content search
jot notes search query --and path=projects/**/*.md
```

Optional semantic validation:

```bash
jot notes search semantic "project planning notes" --mode hybrid
```

## 4) Validate metadata expectations

If you rely on frontmatter tags/status fields:

```bash
jot notes search query --and data.tag=work
jot notes search query --and data.status=active
```

## 5) Normalize workflows with views

```bash
jot notes view --list
jot notes view recent
jot notes view kanban
```

## Common migration pitfalls

### Notebook exists but wrong one is selected

- clear `JOT_NOTEBOOK` if stale
- pass `--notebook` explicitly
- run in notebook root with `.jot.json`

### Notes exist but search returns little/no results

- verify files are markdown
- verify notebook path correctness
- try title search first (`jot notes search "..."`), then `body:` or query filters

### Semantic search behavior differs from keyword search

- expected
- use hybrid mode for balanced behavior

## OpenNotes rename migration

If you’re coming from old `opennotes` naming, run:

```bash
jot notebook migrate list
jot notebook migrate --apply
```

See full details:

- [migration-opennotes-to-jot.md](migration-opennotes-to-jot.md)

## Recommended post-import checklist

- [ ] `jot notes list` shows expected notes
- [ ] text search returns known notes
- [ ] query filters work for your metadata conventions
- [ ] views run for daily workflow
- [ ] semantic search validated for conceptual lookup (optional)
