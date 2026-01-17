package services

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
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

// RenderSQLResults renders SQL query results as an ASCII table.
// Results is a slice of maps where each map represents a row.
// Columns are extracted from the first result map and sorted alphabetically.
func (d *Display) RenderSQLResults(results []map[string]interface{}) error {
	// Handle empty results
	if len(results) == 0 {
		fmt.Println("No results")
		return nil
	}

	// Extract columns from first result
	// Use a map to get unique columns, then sort
	columnSet := make(map[string]bool)
	for key := range results[0] {
		columnSet[key] = true
	}

	// Convert to slice and sort
	columns := make([]string, 0, len(columnSet))
	for col := range columnSet {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	// Calculate column widths
	widths := make(map[string]int)

	// Initialize with column header widths
	for _, col := range columns {
		widths[col] = len(col)
	}

	// Update widths with data widths
	for _, row := range results {
		for _, col := range columns {
			val := fmt.Sprintf("%v", row[col])
			if len(val) > widths[col] {
				widths[col] = len(val)
			}
		}
	}

	// Print header row
	for i, col := range columns {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Printf("%-*s", widths[col], col)
	}
	fmt.Println()

	// Print separator row
	for i, col := range columns {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Print(strings.Repeat("-", widths[col]))
	}
	fmt.Println()

	// Print data rows
	for _, row := range results {
		for i, col := range columns {
			if i > 0 {
				fmt.Print("  ")
			}
			val := fmt.Sprintf("%v", row[col])
			fmt.Printf("%-*s", widths[col], val)
		}
		fmt.Println()
	}

	// Print summary
	fmt.Printf("\n%d row", len(results))
	if len(results) != 1 {
		fmt.Print("s")
	}
	fmt.Println()

	return nil
}
