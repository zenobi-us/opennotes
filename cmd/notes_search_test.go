package cmd

import (
	"strings"
	"testing"
)

func TestIsDSLStyleQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{name: "field query", query: "type:epic", want: true},
		{name: "pipe query", query: "status:todo | limit:5", want: true},
		{name: "plain text", query: "meeting notes", want: false},
		{name: "empty", query: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDSLStyleQuery(tt.query)
			if got != tt.want {
				t.Fatalf("isDSLStyleQuery(%q)=%v want %v", tt.query, got, tt.want)
			}
		})
	}
}

func TestNormalizeFieldlessToTitleQuery(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "single word", input: "epic", want: "title:epic"},
		{name: "multi word phrase", input: "fieldless text", want: `title:"fieldless text"`},
		{name: "trims whitespace", input: "  roadmap  ", want: "title:roadmap"},
		{name: "empty remains empty", input: "   ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeFieldlessToTitleQuery(tt.input)
			if got != tt.want {
				t.Fatalf("normalizeFieldlessToTitleQuery(%q)=%q want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeFieldlessToTitleQuery_EscapesQuotes(t *testing.T) {
	input := `say "hello"`
	got := normalizeFieldlessToTitleQuery(input)
	if !strings.HasPrefix(got, `title:"`) {
		t.Fatalf("expected title quoted query, got %q", got)
	}
	if !strings.Contains(got, `\"hello\"`) {
		t.Fatalf("expected escaped quotes in normalized query, got %q", got)
	}
}
