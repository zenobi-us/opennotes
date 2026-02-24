package parser

import (
	"testing"

	"github.com/zenobi-us/jot/internal/search"
)

func TestParser_Parse_SimpleTerm(t *testing.T) {
	p := New()

	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantType string
		wantVal  string
	}{
		{
			name:     "single word",
			input:    "meeting",
			wantLen:  1,
			wantType: "TermExpr",
			wantVal:  "meeting",
		},
		{
			name:     "quoted phrase",
			input:    `"project meeting"`,
			wantLen:  1,
			wantType: "TermExpr",
			wantVal:  "project meeting",
		},
		{
			name:     "multiple terms",
			input:    "meeting notes",
			wantLen:  2,
			wantType: "TermExpr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(query.Expressions) != tt.wantLen {
				t.Errorf("got %d expressions, want %d", len(query.Expressions), tt.wantLen)
			}

			if tt.wantLen > 0 && tt.wantVal != "" {
				term, ok := query.Expressions[0].(search.TermExpr)
				if !ok {
					t.Errorf("expected TermExpr, got %T", query.Expressions[0])
					return
				}
				if term.Value != tt.wantVal {
					t.Errorf("got value %q, want %q", term.Value, tt.wantVal)
				}
			}
		})
	}
}

func TestParser_Parse_FieldExpr(t *testing.T) {
	p := New()

	tests := []struct {
		name      string
		input     string
		wantField string
		wantOp    search.CompareOp
		wantValue string
	}{
		{
			name:      "tag filter",
			input:     "tag:work",
			wantField: "tag",
			wantOp:    search.OpEquals,
			wantValue: "work",
		},
		{
			name:      "title filter",
			input:     "title:meeting",
			wantField: "title",
			wantOp:    search.OpEquals,
			wantValue: "meeting",
		},
		{
			name:      "path filter",
			input:     "path:projects/",
			wantField: "path",
			wantOp:    search.OpEquals,
			wantValue: "projects/",
		},
		{
			name:      "quoted value",
			input:     `title:"project meeting"`,
			wantField: "title",
			wantOp:    search.OpEquals,
			wantValue: "project meeting",
		},
		{
			name:      "body search",
			input:     "body:important",
			wantField: "body",
			wantOp:    search.OpEquals,
			wantValue: "important",
		},
		{
			name:      "metadata type field",
			input:     "type:epic",
			wantField: "type",
			wantOp:    search.OpEquals,
			wantValue: "epic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(query.Expressions) != 1 {
				t.Fatalf("got %d expressions, want 1", len(query.Expressions))
			}

			field, ok := query.Expressions[0].(search.FieldExpr)
			if !ok {
				t.Fatalf("expected FieldExpr, got %T", query.Expressions[0])
			}

			if field.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", field.Field, tt.wantField)
			}
			if field.Op != tt.wantOp {
				t.Errorf("Op = %q, want %q", field.Op, tt.wantOp)
			}
			if field.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", field.Value, tt.wantValue)
			}
		})
	}
}

func TestParser_Parse_DateExpr(t *testing.T) {
	p := New()

	tests := []struct {
		name      string
		input     string
		wantField string
		wantOp    search.CompareOp
		wantValue string
	}{
		{
			name:      "created equals",
			input:     "created:2024-01-01",
			wantField: "created",
			wantOp:    search.OpEquals,
			wantValue: "2024-01-01",
		},
		{
			name:      "created after",
			input:     "created:>2024-01-01",
			wantField: "created",
			wantOp:    search.OpGt,
			wantValue: "2024-01-01",
		},
		{
			name:      "created before",
			input:     "created:<2024-06-30",
			wantField: "created",
			wantOp:    search.OpLt,
			wantValue: "2024-06-30",
		},
		{
			name:      "modified gte",
			input:     "modified:>=2024-01-01",
			wantField: "modified",
			wantOp:    search.OpGte,
			wantValue: "2024-01-01",
		},
		{
			name:      "modified lte",
			input:     "modified:<=2024-12-31",
			wantField: "modified",
			wantOp:    search.OpLte,
			wantValue: "2024-12-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(query.Expressions) != 1 {
				t.Fatalf("got %d expressions, want 1", len(query.Expressions))
			}

			date, ok := query.Expressions[0].(search.DateExpr)
			if !ok {
				t.Fatalf("expected DateExpr, got %T", query.Expressions[0])
			}

			if date.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", date.Field, tt.wantField)
			}
			if date.Op != tt.wantOp {
				t.Errorf("Op = %q, want %q", date.Op, tt.wantOp)
			}
			if date.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", date.Value, tt.wantValue)
			}
		})
	}
}

func TestParser_Parse_NotExpr(t *testing.T) {
	p := New()

	tests := []struct {
		name      string
		input     string
		wantInner string // Type name of inner expression
	}{
		{
			name:      "negated term",
			input:     "-archived",
			wantInner: "TermExpr",
		},
		{
			name:      "negated tag",
			input:     "-tag:done",
			wantInner: "FieldExpr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(query.Expressions) != 1 {
				t.Fatalf("got %d expressions, want 1", len(query.Expressions))
			}

			not, ok := query.Expressions[0].(search.NotExpr)
			if !ok {
				t.Fatalf("expected NotExpr, got %T", query.Expressions[0])
			}

			switch tt.wantInner {
			case "TermExpr":
				if _, ok := not.Expr.(search.TermExpr); !ok {
					t.Errorf("inner expr is %T, want TermExpr", not.Expr)
				}
			case "FieldExpr":
				if _, ok := not.Expr.(search.FieldExpr); !ok {
					t.Errorf("inner expr is %T, want FieldExpr", not.Expr)
				}
			}
		})
	}
}

func TestParser_Parse_Combined(t *testing.T) {
	p := New()

	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantDesc string // Description of what we expect
	}{
		{
			name:     "tag and title",
			input:    "tag:work title:meeting",
			wantLen:  2,
			wantDesc: "two FieldExpr",
		},
		{
			name:     "term and negation",
			input:    "meeting -archived",
			wantLen:  2,
			wantDesc: "TermExpr and NotExpr",
		},
		{
			name:     "complex query",
			input:    `tag:work created:>2024-01-01 -tag:done "important meeting"`,
			wantLen:  4,
			wantDesc: "FieldExpr, DateExpr, NotExpr, TermExpr",
		},
		{
			name:     "all field types",
			input:    "tag:work title:meeting path:projects/ created:>2024-01-01",
			wantLen:  4,
			wantDesc: "various field expressions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(query.Expressions) != tt.wantLen {
				t.Errorf("got %d expressions, want %d (%s)",
					len(query.Expressions), tt.wantLen, tt.wantDesc)
			}

			if query.Raw != tt.input {
				t.Errorf("Raw = %q, want %q", query.Raw, tt.input)
			}
		})
	}
}

func TestParser_Parse_EmptyInput(t *testing.T) {
	p := New()

	tests := []string{"", "   ", "\t\n"}

	for _, input := range tests {
		t.Run("empty", func(t *testing.T) {
			query, err := p.Parse(input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if !query.IsEmpty() {
				t.Errorf("expected empty query, got %d expressions", len(query.Expressions))
			}
		})
	}
}

func TestParser_Validate(t *testing.T) {
	p := New()

	valid := []string{
		"meeting",
		"tag:work",
		"created:>2024-01-01",
		"-archived",
		`"exact phrase"`,
		"tag:work title:meeting -archived",
	}

	for _, input := range valid {
		t.Run(input, func(t *testing.T) {
			if err := p.Validate(input); err != nil {
				t.Errorf("Validate(%q) = %v, want nil", input, err)
			}
		})
	}
}

func TestParser_Help(t *testing.T) {
	p := New()
	help := p.Help()

	// Check that help contains expected sections
	expected := []string{
		"Basic Search",
		"Field Filters",
		"Date Filters",
		"Negation",
		"tag:",
		"title:",
		"created:",
	}

	for _, s := range expected {
		if !containsString(help, s) {
			t.Errorf("Help() missing %q", s)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || containsString(s[1:], substr)))
}

func TestParser_Interface(t *testing.T) {
	// Verify Parser implements search.Parser
	var _ search.Parser = New()
}
func TestParser_Parse_OrExpressions(t *testing.T) {
	p := New()

	t.Run("simple or", func(t *testing.T) {
		query, err := p.Parse("tag:work OR tag:personal")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if len(query.Expressions) != 1 {
			t.Fatalf("expected single expression, got %d", len(query.Expressions))
		}

		orExpr, ok := query.Expressions[0].(search.OrExpr)
		if !ok {
			t.Fatalf("expected OrExpr, got %T", query.Expressions[0])
		}

		left, ok := orExpr.Left.(search.FieldExpr)
		if !ok || left.Field != "tag" || left.Value != "work" {
			t.Fatalf("unexpected left operand: %#v", orExpr.Left)
		}

		right, ok := orExpr.Right.(search.FieldExpr)
		if !ok || right.Field != "tag" || right.Value != "personal" {
			t.Fatalf("unexpected right operand: %#v", orExpr.Right)
		}
	})

	t.Run("and precedence over or", func(t *testing.T) {
		query, err := p.Parse("tag:work status:todo OR tag:personal")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if len(query.Expressions) != 1 {
			t.Fatalf("expected single expression, got %d", len(query.Expressions))
		}

		orExpr, ok := query.Expressions[0].(search.OrExpr)
		if !ok {
			t.Fatalf("expected OrExpr, got %T", query.Expressions[0])
		}

		leftAnd, ok := orExpr.Left.(search.AndExpr)
		if !ok {
			t.Fatalf("expected AndExpr on left, got %T", orExpr.Left)
		}

		if len(leftAnd.Expressions) != 2 {
			t.Fatalf("expected 2 AND expressions, got %d", len(leftAnd.Expressions))
		}
	})

	t.Run("multiple or chain", func(t *testing.T) {
		query, err := p.Parse("tag:work OR tag:personal OR status:todo")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if len(query.Expressions) != 1 {
			t.Fatalf("expected single expression, got %d", len(query.Expressions))
		}

		outer, ok := query.Expressions[0].(search.OrExpr)
		if !ok {
			t.Fatalf("expected OrExpr, got %T", query.Expressions[0])
		}

		_, leftIsOr := outer.Left.(search.OrExpr)
		if !leftIsOr {
			t.Fatalf("expected nested OrExpr on left")
		}

		rightField, ok := outer.Right.(search.FieldExpr)
		if !ok || rightField.Field != "status" {
			t.Fatalf("unexpected right operand: %#v", outer.Right)
		}
	})
}
