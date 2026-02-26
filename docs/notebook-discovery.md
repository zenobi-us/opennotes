# Notebook Discovery

Jot automatically discovers and loads notebooks based on the user's current working directory using a sophisticated 3-tier priority system. This document outlines the complete algorithm and provides a visual flowchart of the discovery process.

## Overview

The notebook discovery follows a **5-tier priority system** (first wins):

1. **JOT_NOTEBOOK** (highest priority) - Environment variable
2. **--notebook flag** (CLI override) - Command line argument
3. **Current Directory** (direct check) - `.jot.json` in current working directory
4. **Registered Notebooks** (context matching) - Check each registered notebook for context match
5. **Ancestor Search** (fallback) - Walk up directory tree looking for `.jot.json`

## Discovery Flowchart

![Notebook Discovery Flowchart](notebook-discovery.svg)

### 1. JOT_NOTEBOOK Environment Variable (Tier 1 - Highest Priority)

The system first checks if the `JOT_NOTEBOOK` environment variable is set:

```bash
export JOT_NOTEBOOK=/path/to/notebook
jot notes list  # Uses the notebook from envvar
```

If the envvar is set:

1. Check if `.jot.json` exists in that path
2. If yes: Load and open the notebook → **SUCCESS**
3. If no: Continue to Tier 2

### 2. --notebook CLI Flag (Tier 2)

The system checks if the `--notebook` flag was provided on the command line:

```bash
jot notes list --notebook /path/to/notebook
```

If a flag path exists:

1. Check if `.jot.json` exists in that path
2. If yes: Load and open the notebook → **SUCCESS**
3. If no: Continue to Tier 3

### 3. Current Directory (Tier 3)

The system checks if `.jot.json` exists in the current working directory:

```bash
cd /home/user/project  # Contains .jot.json
jot notes list   # Auto-discovers notebook in cwd
```

If `.jot.json` exists in current directory:

1. Load and open the notebook → **SUCCESS**
2. If no: Continue to Tier 4

### 4. Registered Notebooks (Tier 4 - Context Matching)

The system checks notebooks registered in the global configuration:

1. Load global config from `~/.config/jot/config.json`
2. For each registered notebook path:
   - Check if `.jot.json` exists
   - If exists: Load notebook configuration
   - Check if current directory matches any context path using **Context Matching Algorithm**
   - If match found: Load and open the notebook → **SUCCESS**
3. If no registered notebooks match: Continue to Tier 3

#### Context Matching Algorithm

```go
// For each context path in notebook config
for _, context := range notebook.Contexts {
    if strings.HasPrefix(currentWorkingDirectory, context) {
        return notebook // Match found
    }
}
```

**Example:**

```
Notebook contexts: ["/home/user/project", "/tmp/work"]
Current directory: "/home/user/project/src"

Match check: strings.HasPrefix("/home/user/project/src", "/home/user/project")
Result: TRUE → Context matches → Return this notebook
```

### 5. Ancestor Search (Tier 5 - Fallback)

If no environment variable, flag, current directory, or registered notebooks match, the system performs an ancestor directory search:

1. Start with parent directory (not current, as that was checked in Tier 3)
2. Check if `.jot.json` exists in current directory
3. If yes: Load and open the notebook → **SUCCESS**
4. If no: Move to parent directory
5. Repeat until reaching filesystem root (`/`) or empty string
6. If root reached: → **NOT FOUND**

## File Locations & Formats

### Global Configuration

**Location:** `~/.config/jot/config.json`

```json
{
  "notebooks": [
    "/home/user/work-notebook",
    "/home/user/personal-notebook",
    "/tmp/temp-notebook"
  ]
}
```

### Notebook Configuration

**Location:** `<notebook_directory>/.jot.json`

```json
{
  "root": "./notes",
  "name": "Project Notebook",
  "contexts": ["/home/user/project", "/home/user/project-related"],
  "templates": {
    "default": "# {{.Title}}\n\nDate: {{.Date}}\n\n"
  },
  "groups": [
    {
      "name": "Default",
      "globs": ["**/*.md"],
      "metadata": {}
    }
  ]
}
```

## Key Features

### Deterministic Behavior

- **Clear Priority**: Declared > Registered > Ancestor
- **First Match Wins**: Stops at first successful discovery
- **No Ambiguity**: Priority order prevents conflicts

### Graceful Fallback

- If higher priority method fails, try next tier
- Comprehensive search ensures maximum discovery success
- Returns `nil` only when all methods exhausted

### Context-Aware

- Registered notebooks define active contexts
- Automatically selects appropriate notebook for current work environment
- Supports multiple context paths per notebook

### Efficient Discovery

- Stops immediately upon successful match
- Minimal filesystem operations
- Fast context matching using string prefix comparison

## State Transitions Summary

1. **TIER 1: JOT_NOTEBOOK envvar** → Success or Continue to Tier 2
2. **TIER 2: --notebook flag** → Success or Continue to Tier 3
3. **TIER 3: Current Directory** → Success or Continue to Tier 4
4. **TIER 4: REGISTERED SEARCH** → For each registered notebook:
   - Check exists → Check context match → Success or Continue
5. **TIER 5: ANCESTOR SEARCH** → Walk up directories until found or root
6. **SUCCESS** → Return notebook instance
7. **NOT FOUND** → Return nil

This discovery system ensures Jot works seamlessly across different project environments while maintaining predictable, efficient behavior. The priority order follows the principle of least surprise: environment variable (global) → flag (explicit) → auto-detection (implicit).
