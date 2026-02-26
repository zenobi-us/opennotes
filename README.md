# Jot

![Jot Banner](./banner.png)

A simple, fast, and powerful CLI tool for managing your markdown notes. Jot helps you organize, search, and manage your thoughts without leaving the terminal.

## Features

- **Notebook Storage**: Choose where you want notes stored—in your repo, home directory, or anywhere on your filesystem
- **Create, list, search, and view notes** from the command line
- **Associate default metadata** based on groups or folders
- **Preset queries**: Kanban, Daily Notes, Tags, and more
- **Full-text search** with fuzzy matching and boolean queries
- **Semantic search** (optional): Find notes by meaning using hybrid keyword + semantic retrieval

## Installation

```bash
go install github.com/zenobi-us/jot@latest
```

*Requires Go 1.24+*

## Migrating from OpenNotes

If you used the old `opennotes` CLI/repo, use this migration guide:

- **[OpenNotes → Jot Migration Guide](docs/migration-opennotes-to-jot.md)**

Quick migration checks:

```bash
# Preview pending migrations
jot notebook migrate list

# Dry-run migration actions
jot notebook migrate

# Apply migrations
jot notebook migrate --apply
```

## Quick Start

Get started in seconds:

1.  **Initialize a notebook** in your current directory:
    ```bash
    jot init
    ```

2.  **Create your first note**:
    ```bash
    jot notes add "My First Idea"
    ```

3.  **List your notes**:
    ```bash
    jot notes list
    ```

4.  **Search your notes**:
    ```bash
    jot notes search "Idea"
    ```

## Common Commands

### Creating Notes

Create a new note with a title. You can also specify a path if you want to organize notes into folders.

```bash
# Create a note in the root
jot notes add "Meeting Notes"

# Create a note in a subfolder
jot notes add "Project Specs" projects/
```

### Listing Notes

See all your notes in the current notebook.

```bash
jot notes list
```

### Searching Notes

Fieldless search is title-only (`jot notes search "..."` normalizes to `title:...`).
Use explicit `body:` when you want body/content matching.

```bash
jot notes search "important"
jot notes search "body:important"
```

### Views 

Jot supports various preset views to help you organize and visualize your notes.

Output is JSON by default.

```bash
# view all available views
jot notes view

# display notes in kanban view
jot notes view kanban
```

## Configuration

Jot works out of the box, but you can customize it.

Global configuration is stored in:
- **Linux**: `~/.config/jot/config.json`
- **macOS**: `~/Library/Preferences/jot/config.json`
- **Windows**: `%APPDATA%\jot\config.json`

## Semantic Search (Optional)

Find notes by meaning, not just keywords. Semantic search understands concepts and paraphrases.

```bash
# Hybrid search (default): combines keyword + semantic ranking
jot notes search semantic "meeting notes about project timeline"

# Pure semantic mode: meaning-based, ideal for conceptual queries
jot notes search semantic "discussions about deadlines" --mode semantic

# With filters: combine semantic search with boolean conditions
jot notes search semantic "architecture decisions" --and data.tag=design

# Explain mode: see why each result matched
jot notes search semantic "workflow improvements" --explain
```

**When to use semantic vs regular search:**
- **Regular search**: Exact keywords, specific terms, quick lookups
- **Semantic search**: Conceptual queries, paraphrases, "find notes about X"

For more details, see **[Semantic Search Guide](docs/semantic-search-guide.md)**.

## Advanced Usage

Jot provides powerful search capabilities with full-text search, fuzzy matching, boolean query operators, and semantic retrieval for complex filtering.

For advanced features like boolean queries, JSON output for automation, and multi-notebook management, please see our **[Advanced Documentation](docs/getting-started-power-users.md)**.
