package services

import (
	"bytes"
	"text/template"

	"github.com/charmbracelet/glamour"
)

// Display handles terminal rendering with glamour.
type Display struct {
	renderer *glamour.TermRenderer
}

// NewDisplay creates a new display service with glamour rendering.
func NewDisplay() (*Display, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return nil, err
	}

	return &Display{renderer: renderer}, nil
}

// Render renders markdown content to the terminal.
func (d *Display) Render(markdown string) (string, error) {
	return d.renderer.Render(markdown)
}

// RenderTemplate renders a Go template with context, then renders as markdown.
func (d *Display) RenderTemplate(tmpl string, ctx any) (string, error) {
	// Parse and execute Go template
	t, err := template.New("display").Parse(tmpl)
	if err != nil {
		// Fallback: return template as-is if parsing fails
		return tmpl, nil
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		// Fallback: return template as-is if execution fails
		return tmpl, nil
	}

	// Render the result as markdown
	rendered, err := d.renderer.Render(buf.String())
	if err != nil {
		// Fallback: return unrendered content
		return buf.String(), nil
	}

	return rendered, nil
}
