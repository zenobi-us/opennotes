package parser

import (
	"fmt"
	"strings"

	"github.com/zenobi-us/jot/internal/search"
)

// Parser implements the search.Parser interface using Participle.
type Parser struct{}

// New creates a new Parser instance.
func New() *Parser {
	return &Parser{}
}

// Parse parses a query string into a Query AST.
func (p *Parser) Parse(input string) (*search.Query, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return &search.Query{Raw: input}, nil
	}

	ast, err := queryParser.ParseString("", input)
	if err != nil {
		return nil, formatParseError(input, err)
	}

	query := convert(ast)
	query.Raw = input
	return query, nil
}

// Validate checks if a query string is syntactically valid.
func (p *Parser) Validate(input string) error {
	_, err := p.Parse(input)
	return err
}

// Help returns syntax help for the query language.
func (p *Parser) Help() string {
	return `Jot Query Syntax
======================

Basic Search:
  meeting              Search for "meeting" in all fields
  "exact phrase"       Search for exact phrase

Field Filters:
  tag:work             Notes with tag "work"
  title:meeting        Notes with "meeting" in title
  path:projects/       Notes in projects/ directory
  body:important       Search only in body text

Date Filters:
  created:2024-01-01   Created on specific date
  created:>2024-01-01  Created after date
  created:<2024-01-01  Created before date
  modified:>=2024-06   Modified on or after date

Negation:
  -archived            Exclude notes containing "archived"
  -tag:done            Exclude notes with tag "done"

Combining (implicit AND):
  tag:work status:todo Notes with tag "work" AND status "todo"
  meeting -archived    Contains "meeting" but not "archived"

Supported Fields:
  tag       - Filter by tag
  title     - Search in title
  body      - Search in body only
  path      - Filter by path prefix
  created   - Filter by creation date
  modified  - Filter by modification date
  status    - Filter by status field
  type      - Filter by type field
  project   - Filter by project field

Examples:
  tag:work                      All work-tagged notes
  tag:work title:meeting        Work notes about meetings
  created:>2024-01-01 -archived Recent notes, not archived
  "project plan" tag:urgent     Exact phrase with tag filter
`
}

// formatParseError creates a user-friendly error from a parse error.
func formatParseError(input string, err error) *search.ParseError {
	// Try to extract position from Participle error
	errStr := err.Error()

	// Basic error formatting
	parseErr := &search.ParseError{
		Message: "invalid query syntax",
		Input:   input,
	}

	// Try to provide helpful suggestions
	if strings.Contains(errStr, "unexpected") {
		parseErr.Message = "unexpected character or token"
		if strings.Contains(input, "::") {
			parseErr.Suggestion = "use single colon for field:value"
		}
	}

	if strings.Contains(errStr, "expected") {
		parseErr.Message = "incomplete query"
		if strings.HasSuffix(strings.TrimSpace(input), ":") {
			parseErr.Suggestion = "add a value after the colon (e.g., tag:work)"
		}
	}

	// Include original error for debugging
	parseErr.Message = fmt.Sprintf("%s: %v", parseErr.Message, err)

	return parseErr
}

// Ensure Parser implements search.Parser interface.
var _ search.Parser = (*Parser)(nil)
