# Getting Started for Power Users

This guide is for users who want to move fast with Jot’s advanced search, views, and automation patterns.

## 1) Import or open a notebook

```bash
jot notebook create "Work" --path ~/notes/work
jot notes list
```

If you already have notebooks registered, `jot` will resolve context automatically.

## 2) Master search surfaces

### Text search

```bash
jot notes search "incident"          # fieldless => title:incident
jot notes search "body:incident"     # explicit body/content search
```

### Structured boolean filters

```bash
jot notes search query --and data.tag=ops --not data.status=archived
jot notes search query --and path=runbooks/**/*.md
```

### Semantic retrieval

```bash
jot notes search semantic "postmortem follow up actions"
jot notes search semantic "on-call handoff" --mode hybrid --explain
```

## 3) Use views for repeatable workflows

```bash
jot notes view --list
jot notes view recent
jot notes view kanban
jot notes view recent --format json
```

Views are the best way to standardize team workflows without rewriting command logic.

## 4) Build CLI automation

Examples:

```bash
# Daily quick report
jot notes view recent --format json | jq 'length'

# Find active engineering tasks
jot notes search query --and data.tag=engineering --and data.status=active

# Semantic discovery with explain output for review
jot notes search semantic "release blockers" --mode hybrid --explain
```

## 5) Practical troubleshooting checks

```bash
jot --version
jot notebook list
jot notes list
jot notes search "test"
```

If behavior looks wrong, use:

- [Troubleshooting](getting-started-troubleshooting.md)
- [Notebook discovery](notebook-discovery.md)

## Next docs

- [Search command reference](commands/notes-search.md)
- [Semantic search guide](semantic-search-guide.md)
- [Views guide](views-guide.md)
- [Automation recipes](automation-recipes.md)
