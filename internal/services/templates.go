package services

import (
	"bytes"
	"strings"
	"text/template"
)

// Templates for rendering various outputs.
var Templates = struct {
	NoteList     string
	NoteDetail   string
	NotebookInfo string
	NotebookList string
}{
	NoteList: strings.TrimSpace(`
{{- if eq (len .Notes) 0 -}}
No notes found.
{{- else -}}
## Notes ({{ len .Notes }})

{{ range .Notes -}}
- [{{ .File.Relative }}](file://{{ .File.Filepath }})
{{ end -}}
{{- end -}}
`),

	NoteDetail: strings.TrimSpace(`
# {{ .Title }}

**File:** {{ .File.Relative }}

{{ if .Metadata -}}
**Metadata:**
{{ range $key, $value := .Metadata -}}
- {{ $key }}: {{ $value }}
{{ end }}
{{ end -}}

---

{{ .Content }}
`),

	NotebookInfo: strings.TrimSpace(`
## {{ .Config.Name }}

| Property | Value |
|----------|-------|
| Config | {{ .Config.Path }} |
| Root | {{ .Config.Root }} |

{{ if .Config.Contexts -}}
### Contexts
{{ range .Config.Contexts -}}
- {{ . }}
{{ end }}
{{ end -}}

{{ if .Config.Groups -}}
### Groups
{{ range .Config.Groups -}}
- **{{ .Name }}** ({{ range $i, $g := .Globs }}{{ if $i }}, {{ end }}{{ $g }}{{ end }})
{{ end }}
{{ end -}}
`),

	NotebookList: strings.TrimSpace(`
{{- if eq (len .Notebooks) 0 -}}
No notebooks found.
{{- else -}}
## Notebooks ({{ len .Notebooks }})

{{ range .Notebooks -}}
### {{ .Config.Name }}
- **Path:** {{ .Config.Path }}
- **Root:** {{ .Config.Root }}
{{ if .Config.Contexts -}}
- **Contexts:** {{ range $i, $c := .Config.Contexts }}{{ if $i }}, {{ end }}{{ $c }}{{ end }}
{{ end }}
{{ end -}}
{{- end -}}
`),
}

// TuiRender is a convenience function to render a template with glamour.
func TuiRender(tmpl string, ctx any) (string, error) {
	display, err := NewDisplay()
	if err != nil {
		// Fallback without glamour rendering
		t, err := template.New("tui").Parse(tmpl)
		if err != nil {
			return tmpl, nil
		}

		var buf bytes.Buffer
		if err := t.Execute(&buf, ctx); err != nil {
			return tmpl, nil
		}

		return buf.String(), nil
	}

	return display.RenderTemplate(tmpl, ctx)
}
