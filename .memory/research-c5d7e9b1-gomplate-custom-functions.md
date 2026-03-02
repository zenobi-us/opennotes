---
id: c5d7e9b1-gomplate
title: Gomplate Custom Functions Research
created_at: 2026-03-02T11:27:00+10:30
updated_at: 2026-03-02T13:28:00+10:30
status: complete
parent_epic: c5d7e9b1
---

# Gomplate Custom Functions Research

## Executive Summary

**Yes, gomplate fully supports custom functions when embedded as a library.** The `RenderOptions.Funcs` field accepts a standard `template.FuncMap` that merges with gomplate's built-in functions. You can also create custom namespaces using the same pattern gomplate uses internally.

## Key Findings

### 1. Gomplate Library API

Gomplate v4/v5 provides a clean embedding API:

```go
import (
    "context"
    "github.com/hairyhenderson/gomplate/v4"  // or v5
)

// Create a renderer with options
renderer := gomplate.NewRenderer(gomplate.RenderOptions{
    Funcs: customFuncs,  // Your custom template.FuncMap
})

// Render templates
err := renderer.Render(ctx, "template-name", templateText, outputWriter)
```

### 2. Custom Function Injection

**Method 1: Direct FuncMap (Simple)**

```go
import "text/template"

customFuncs := template.FuncMap{
    "nanoid":  generateNanoID,
    "slug":    slugify,
    "reverse": reverseString,
}

renderer := gomplate.NewRenderer(gomplate.RenderOptions{
    Funcs: customFuncs,
})
```

**Method 2: Merge with gomplate's Built-ins**

```go
// Get all gomplate built-in functions
funcs := gomplate.CreateFuncs(ctx)

// Add your custom functions (overwrites on collision)
funcs["nanoid"] = generateNanoID
funcs["slug"] = mySlugify

renderer := gomplate.NewRenderer(gomplate.RenderOptions{
    Funcs: funcs,
})
```

### 3. Creating Custom Namespaces

Gomplate namespaces are implemented as struct methods returned by a factory function:

```go
// Define namespace struct
type JotFuncs struct {
    ctx context.Context
}

// Add methods to namespace
func (f *JotFuncs) Slug(s string) string {
    return core.Slugify(s)
}

func (f *JotFuncs) NanoID() string {
    id, _ := nanoid.New()
    return id
}

func (f *JotFuncs) Timestamp() string {
    return time.Now().Format(time.RFC3339)
}

// Create FuncMap with namespace
func CreateJotFuncs(ctx context.Context) template.FuncMap {
    ns := &JotFuncs{ctx: ctx}
    return template.FuncMap{
        // Namespace access: {{ jot.Slug "Hello World" }}
        "jot": func() any { return ns },
        
        // Top-level aliases: {{ slug "Hello World" }}
        "slug":      ns.Slug,
        "nanoid":    ns.NanoID,
        "timestamp": ns.Timestamp,
    }
}
```

**Template usage:**
```gotmpl
{{ jot.Slug "Hello World" }}     → hello-world
{{ jot.NanoID }}                 → V1StGXR8_Z5jdHi6B-myT
{{ "My Title" | slug }}          → my-title (using top-level alias)
```

### 4. Complete Implementation Example

```go
// internal/services/gomplate_renderer.go
package services

import (
    "bytes"
    "context"
    "text/template"
    "time"

    "github.com/hairyhenderson/gomplate/v4"
    "github.com/jaevor/go-nanoid"
    "github.com/zenobi-us/jot/internal/core"
)

// JotFuncs provides jot-specific template functions
type JotFuncs struct {
    ctx context.Context
}

func (f *JotFuncs) Slug(s string) string {
    return core.Slugify(s)
}

func (f *JotFuncs) NanoID() string {
    gen, _ := nanoid.Standard(21)
    return gen()
}

func (f *JotFuncs) ShortID() string {
    gen, _ := nanoid.Standard(8)
    return gen()
}

func (f *JotFuncs) Now() time.Time {
    return time.Now()
}

// CreateJotFuncs creates the jot namespace functions
func CreateJotFuncs(ctx context.Context) template.FuncMap {
    ns := &JotFuncs{ctx: ctx}
    return template.FuncMap{
        // Namespace: {{ jot.Slug "text" }}
        "jot": func() any { return ns },
        
        // Top-level shortcuts
        "slug":    ns.Slug,
        "nanoid":  ns.NanoID,
        "shortid": ns.ShortID,
    }
}

// GomplateRenderer wraps gomplate for jot templates
type GomplateRenderer struct {
    renderer gomplate.Renderer
}

// NewGomplateRenderer creates a renderer with jot + gomplate functions
func NewGomplateRenderer(ctx context.Context) *GomplateRenderer {
    // Start with gomplate's 200+ built-in functions
    funcs := gomplate.CreateFuncs(ctx)
    
    // Merge jot custom functions
    for k, v := range CreateJotFuncs(ctx) {
        funcs[k] = v
    }
    
    renderer := gomplate.NewRenderer(gomplate.RenderOptions{
        Funcs: funcs,
    })
    
    return &GomplateRenderer{renderer: renderer}
}

// Render executes a template string with the given data
func (r *GomplateRenderer) Render(ctx context.Context, name, tmpl string, data any) (string, error) {
    buf := &bytes.Buffer{}
    err := r.renderer.Render(ctx, name, tmpl, buf)
    if err != nil {
        return "", err
    }
    return buf.String(), nil
}
```

### 5. Available gomplate Functions (Partial List)

When embedding gomplate, you get 200+ built-in functions:

| Namespace | Key Functions |
|-----------|--------------|
| `strings` | `Slug`, `ToUpper`, `ToLower`, `TrimSpace`, `ReplaceAll`, `Split`, `Join` |
| `coll` | `Dict`, `Slice`, `Keys`, `Values`, `Has`, `Merge`, `JQ`, `JSONPath` |
| `conv` | `ToBool`, `ToInt`, `ToString`, `Default`, `Join` |
| `time` | `Now`, `Parse`, `Format`, `ParseDuration` |
| `math` | `Add`, `Sub`, `Mul`, `Div`, `Round`, `Floor`, `Ceil` |
| `crypto` | `SHA256`, `Bcrypt`, `EncryptAES`, `DecryptAES` |
| `base64` | `Encode`, `Decode` |
| `regexp` | `Match`, `Find`, `Replace`, `Split` |
| `filepath` | `Base`, `Dir`, `Ext`, `Join`, `Clean` |
| `data` | `JSON`, `YAML`, `TOML`, `CSV` |
| `random` | `ASCII`, `Alpha`, `AlphaNum`, `Number` |
| `uuid` | `V4`, `V1`, `Parse` |

**Note:** `strings.Slug` wraps `github.com/gosimple/slug` internally.

### 6. Limitations

| Limitation | Impact | Workaround |
|------------|--------|------------|
| Binary size | +5-8MB from gomplate dependencies | Acceptable for CLI tool |
| Internal funcs package | `github.com/hairyhenderson/gomplate/v4/internal/funcs` not importable | Use `CreateFuncs()` then merge |
| Cannot modify existing namespace | Can't add methods to `strings` namespace | Create parallel namespace or top-level func |
| Context dependency | Some funcs need context (datasources, env) | Pass context to `CreateFuncs()` |

### 7. Performance Considerations

- **Template parsing**: One-time cost, cache parsed templates
- **Function lookup**: O(1) map lookup, negligible
- **gomplate overhead**: Minimal - just function registration
- **Recommendation**: Parse templates at startup, reuse renderer

```go
// Cache parsed templates
var templateCache = sync.Map{}

func getCachedRenderer() *GomplateRenderer {
    if r, ok := templateCache.Load("renderer"); ok {
        return r.(*GomplateRenderer)
    }
    r := NewGomplateRenderer(context.Background())
    templateCache.Store("renderer", r)
    return r
}
```

## Recommendation for Jot

**Two viable paths:**

### Path A: Stick with text/template + FuncMap (Current)

If you only need a few custom functions (slug, nanoid):

```go
var templateFuncs = template.FuncMap{
    "slug":   core.Slugify,
    "nanoid": generateNanoID,
}

tmpl, _ := template.New(name).Funcs(templateFuncs).Parse(content)
```

**Pros:** Zero new deps, small binary, full control
**Cons:** Manual function implementation, no data sources

### Path B: Embed gomplate (Full Power)

If you want gomplate's rich function library:

```go
renderer := gomplate.NewRenderer(gomplate.RenderOptions{
    Funcs: CreateJotFuncs(ctx),
})
```

**Pros:** 200+ functions, data sources, proven library
**Cons:** +5MB binary, learning curve for users

### Verdict

For jot's use case (simple note templates), **Path A is sufficient**. Gomplate is overkill unless users need:
- Remote data sources in templates
- Complex data transformations
- Cryptographic functions

Add gomplate later if user demand exists.

## References

- [gomplate pkg.go.dev](https://pkg.go.dev/github.com/hairyhenderson/gomplate/v4)
- [gomplate source: render.go](https://github.com/hairyhenderson/gomplate/blob/main/render.go)
- [gomplate source: funcs.go](https://github.com/hairyhenderson/gomplate/blob/main/funcs.go)
- [gomplate source: strings.go](https://github.com/hairyhenderson/gomplate/blob/main/internal/funcs/strings.go)
- [Go text/template FuncMap](https://pkg.go.dev/text/template#FuncMap)
- [gosimple/slug](https://github.com/gosimple/slug)
- [go-nanoid](https://github.com/jaevor/go-nanoid)
