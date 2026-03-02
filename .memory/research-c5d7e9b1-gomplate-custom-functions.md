---
id: c5d7e9b1-gomplate
title: Gomplate Custom Functions Research
created_at: 2026-03-02T11:27:00+10:30
updated_at: 2026-03-02T11:27:00+10:30
status: complete
parent_epic: c5d7e9b1
---

# Gomplate Custom Functions Research

## Executive Summary

**Good news: gomplate has `strings.Slug` built-in**, and it wraps `github.com/gosimple/slug` under the hood. However, **Jot uses Go's standard `text/template`**, not gomplate. Two paths forward exist.

## Key Findings

### 1. Gomplate Has Built-in Slug

From [gomplate docs](https://docs.gomplate.ca/functions/strings/):

```
strings.Slug

Creates a "slug" from a given string - supports Unicode correctly.
This wraps the github.com/gosimple/slug package.

Usage:
  strings.Slug input
  input | strings.Slug

Examples:
  {{ "Hello, world!" | strings.Slug }}    → hello-world
  {{ "Rock & Roll @ Cafe Wha?" | strings.Slug }} → rock-and-roll-at-cafe-wha
```

### 2. Jot's Current State

Jot uses Go's **standard `text/template`** (not gomplate):

```go
// internal/services/templates.go
import "text/template"

tmpl, err := template.New(name).Parse(string(content))
```

Jot already has `core.Slugify()` in `internal/core/strings.go`:
- Custom regex-based implementation
- Lowercase, remove special chars, spaces→hyphens
- Does NOT use `gosimple/slug` (no Unicode transliteration)

### 3. Options for Template Slug Function

#### Option A: Add FuncMap to Existing Templates (Recommended)

Minimal change - register `core.Slugify` as a template function:

```go
// internal/services/templates.go

import (
    "github.com/zenobi-us/jot/internal/core"
    "text/template"
)

// Template function map
var templateFuncs = template.FuncMap{
    "slug":      core.Slugify,
    "slugify":   core.Slugify,  // alias
    "kebabCase": core.Slugify,  // alias for clarity
}

// loadTemplate loads a template with custom functions
func loadTemplate(name string) (*template.Template, error) {
    content, err := templateFiles.ReadFile(fmt.Sprintf("templates/%s.gotmpl", name))
    if err != nil {
        return nil, fmt.Errorf("failed to read template file: %w", err)
    }

    // IMPORTANT: .Funcs() must come BEFORE .Parse()
    tmpl, err := template.New(name).Funcs(templateFuncs).Parse(string(content))
    if err != nil {
        return nil, fmt.Errorf("failed to parse template: %w", err)
    }

    return tmpl, nil
}
```

**Template usage:**
```gotmpl
{{ .title | slug }}
{{ slug .title }}
```

#### Option B: Import gomplate's FuncMap

Use gomplate as a library to get all 200+ functions:

```go
import (
    "text/template"
    gfuncs "github.com/hairyhenderson/gomplate/v4/funcs"
)

func createTemplateFuncs() template.FuncMap {
    f := template.FuncMap{}
    gfuncs.AddStringFuncs(f)  // adds strings.Slug, etc.
    gfuncs.AddCollFuncs(f)    // adds collection functions
    // ... other namespaces as needed
    return f
}
```

**Template usage:**
```gotmpl
{{ strings.Slug .title }}
```

**Tradeoff:** Adds ~5MB to binary, many unused functions.

#### Option C: Switch to gosimple/slug

Replace custom `core.Slugify` with `github.com/gosimple/slug`:

```go
import "github.com/gosimple/slug"

var templateFuncs = template.FuncMap{
    "slug": slug.Make,
}
```

**Pros:** Better Unicode support (e.g., "Héllo Wörld" → "hello-world")
**Cons:** Adds dependency, current impl may be sufficient

### 4. Pre-processing Alternative

Instead of template functions, slug during data preparation:

```go
// Before passing to template
data := map[string]any{
    "title":    title,
    "slug":     core.Slugify(title),  // pre-computed
    "filename": core.Slugify(title) + ".md",
}
```

**Template:**
```gotmpl
Path: {{ .slug }}.md
```

**Tradeoff:** Less flexible, can't slug arbitrary fields in template.

## Recommendation

**Use Option A** - minimal change, keeps existing `core.Slugify`:

1. Add `templateFuncs` map to `templates.go`
2. Call `.Funcs(templateFuncs)` before `.Parse()`
3. Use `{{ .field | slug }}` in templates

This approach:
- Zero new dependencies
- Reuses tested `core.Slugify`
- Matches project philosophy (thin abstractions)
- Can easily swap to `gosimple/slug` later if Unicode issues arise

## Code Example

Complete implementation:

```go
// internal/services/templates.go
package services

import (
    "bytes"
    "embed"
    "fmt"
    "text/template"
    
    "github.com/zenobi-us/jot/internal/core"
)

//go:embed templates/*.gotmpl
var templateFiles embed.FS

// templateFuncs provides custom functions for all templates
var templateFuncs = template.FuncMap{
    "slug": core.Slugify,
}

var loadedTemplates map[string]*template.Template

func init() {
    loadedTemplates = make(map[string]*template.Template)

    templateNames := []string{"note-list", "note-detail", "notebook-info", "notebook-list", "note-search-semantic"}
    for _, name := range templateNames {
        tmpl, err := loadTemplate(name)
        if err != nil {
            fmt.Printf("warning: failed to load template %s: %v\n", name, err)
            continue
        }
        loadedTemplates[name] = tmpl
    }
}

func loadTemplate(name string) (*template.Template, error) {
    content, err := templateFiles.ReadFile(fmt.Sprintf("templates/%s.gotmpl", name))
    if err != nil {
        return nil, fmt.Errorf("failed to read template file: %w", err)
    }

    // Register custom functions BEFORE parsing
    tmpl, err := template.New(name).Funcs(templateFuncs).Parse(string(content))
    if err != nil {
        return nil, fmt.Errorf("failed to parse template: %w", err)
    }

    return tmpl, nil
}
```

## References

- [gomplate strings.Slug docs](https://docs.gomplate.ca/functions/strings/)
- [Go text/template FuncMap](https://pkg.go.dev/text/template#FuncMap)
- [gomplate funcs package](https://pkg.go.dev/github.com/hairyhenderson/gomplate/v4/funcs)
- [gosimple/slug](https://github.com/gosimple/slug)
