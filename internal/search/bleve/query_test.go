package bleve

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zenobi-us/jot/internal/search"
)

func TestTranslateQuery_Empty(t *testing.T) {
	q, err := TranslateQuery(nil)
	require.NoError(t, err)
	assert.NotNil(t, q)
}

func TestTranslateQuery_SingleTerm(t *testing.T) {
	query := &search.Query{
		Expressions: []search.Expr{
			search.TermExpr{Value: "meeting"},
		},
	}

	q, err := TranslateQuery(query)
	require.NoError(t, err)
	assert.NotNil(t, q)
}

func TestTranslateQuery_FieldExpr(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value string
	}{
		{"tag", "tag", "work"},
		{"title", "title", "meeting"},
		{"path", "path", "projects/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &search.Query{
				Expressions: []search.Expr{
					search.FieldExpr{Field: tt.field, Value: tt.value},
				},
			}

			q, err := TranslateQuery(query)
			require.NoError(t, err)
			assert.NotNil(t, q)
		})
	}
}

func TestTranslateQuery_NotExpr(t *testing.T) {
	query := &search.Query{
		Expressions: []search.Expr{
			search.NotExpr{
				Expr: search.TermExpr{Value: "archived"},
			},
		},
	}

	q, err := TranslateQuery(query)
	require.NoError(t, err)
	assert.NotNil(t, q)
}

func TestTranslateQuery_OrExpr(t *testing.T) {
	query := &search.Query{
		Expressions: []search.Expr{
			search.OrExpr{
				Left:  search.FieldExpr{Field: "tag", Value: "work"},
				Right: search.FieldExpr{Field: "tag", Value: "personal"},
			},
		},
	}

	q, err := TranslateQuery(query)
	require.NoError(t, err)
	assert.NotNil(t, q)
}

func TestTranslateQuery_DateExpr(t *testing.T) {
	tests := []struct {
		name  string
		op    search.CompareOp
		value string
	}{
		{"greater than", search.OpGt, "2024-01-01"},
		{"less than", search.OpLt, "2024-06-30"},
		{"equals", search.OpEquals, "2024-03-15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &search.Query{
				Expressions: []search.Expr{
					search.DateExpr{Field: "created", Op: tt.op, Value: tt.value},
				},
			}

			q, err := TranslateQuery(query)
			require.NoError(t, err)
			assert.NotNil(t, q)
		})
	}
}

func TestTranslateQuery_MultipleExpressions(t *testing.T) {
	query := &search.Query{
		Expressions: []search.Expr{
			search.FieldExpr{Field: "tag", Value: "work"},
			search.TermExpr{Value: "meeting"},
			search.NotExpr{Expr: search.TermExpr{Value: "archived"}},
		},
	}

	q, err := TranslateQuery(query)
	require.NoError(t, err)
	assert.NotNil(t, q)
}

func TestTranslateExpr_ExistsExpr(t *testing.T) {
	t.Run("has:tag translates to regexp exists query", func(t *testing.T) {
		query := &search.Query{
			Expressions: []search.Expr{
				search.ExistsExpr{Field: "tag", Negated: false},
			},
		}

		q, err := TranslateQuery(query)
		require.NoError(t, err)
		assert.NotNil(t, q)
	})

	t.Run("missing:tag translates to boolean NOT exists query", func(t *testing.T) {
		query := &search.Query{
			Expressions: []search.Expr{
				search.ExistsExpr{Field: "tag", Negated: true},
			},
		}

		q, err := TranslateQuery(query)
		require.NoError(t, err)
		assert.NotNil(t, q)
	})

	t.Run("has:status translates correctly", func(t *testing.T) {
		query := &search.Query{
			Expressions: []search.Expr{
				search.ExistsExpr{Field: "status", Negated: false},
			},
		}

		q, err := TranslateQuery(query)
		require.NoError(t, err)
		assert.NotNil(t, q)
	})

	t.Run("exists combined with other expressions", func(t *testing.T) {
		query := &search.Query{
			Expressions: []search.Expr{
				search.ExistsExpr{Field: "tag", Negated: false},
				search.FieldExpr{Field: "status", Value: "todo"},
			},
		}

		q, err := TranslateQuery(query)
		require.NoError(t, err)
		assert.NotNil(t, q)
	})
}

func TestTranslateFindOpts_Empty(t *testing.T) {
	opts := search.FindOpts{}

	q, err := TranslateFindOpts(opts)
	require.NoError(t, err)
	assert.NotNil(t, q)
}

func TestTranslateFindOpts_WithTags(t *testing.T) {
	opts := search.FindOpts{}.WithTags("work", "urgent")

	q, err := TranslateFindOpts(opts)
	require.NoError(t, err)
	assert.NotNil(t, q)
}

func TestTranslateFindOpts_WithExcludeTags(t *testing.T) {
	opts := search.FindOpts{}.ExcludingTags("archived")

	q, err := TranslateFindOpts(opts)
	require.NoError(t, err)
	assert.NotNil(t, q)
}

func TestTranslateFindOpts_WithPath(t *testing.T) {
	opts := search.FindOpts{}.WithPath("projects/")

	q, err := TranslateFindOpts(opts)
	require.NoError(t, err)
	assert.NotNil(t, q)
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2024-01-15", false},
		{"2024/01/15", false},
		{"today", false},
		{"yesterday", false},
		{"this-week", false},
		{"this-month", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseDate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizeField(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"path", FieldPath},
		{"p", FieldPath},
		{"title", FieldTitle},
		{"t", FieldTitle},
		{"body", FieldBody},
		{"content", FieldBody},
		{"tag", FieldTags},
		{"tags", FieldTags},
		{"created", FieldCreated},
		{"date", FieldCreated},
		{"modified", FieldModified},
		{"updated", FieldModified},
		{"status", FieldMetadata + ".status"},
		{"type", FieldMetadata + ".type"},
		{"custom", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeField(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestTranslateQuery_AndExpr(t *testing.T) {
	andExpr := search.AndExpr{
		Expressions: []search.Expr{
			search.FieldExpr{Field: "tag", Value: "work"},
			search.FieldExpr{Field: "status", Value: "todo"},
		},
	}

	query := &search.Query{Expressions: []search.Expr{andExpr}}

	q, err := TranslateQuery(query)
	require.NoError(t, err)
	assert.NotNil(t, q)
}
