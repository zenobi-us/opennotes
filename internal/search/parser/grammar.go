package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Grammar AST types for Participle parser.
// These are intermediate types that get converted to search.Expr types.

// queryAST is the root grammar node.
type queryAST struct {
	Clause *clauseAST     `parser:"@@?"`
	Or     []*orClauseAST `parser:"(@@)*"`
}

// clauseAST represents a chain of expressions combined via implicit AND.
type clauseAST struct {
	Expressions []*expressionAST `parser:"@@+"`
}

// orClauseAST represents an OR clause: OR <clause>
type orClauseAST struct {
	Operator string     `parser:"@OrKeyword"`
	Clause   *clauseAST `parser:"@@"`
}

// expressionAST represents a single expression in the query.
type expressionAST struct {
	Existence *existenceExprAST `parser:"  @@"`
	Not       *notExprAST       `parser:"| @@"`
	Field     *fieldExprAST     `parser:"| @@"`
	Term      *termAST          `parser:"| @@"`
}

// existenceExprAST represents an existence check: has:field or missing:field
type existenceExprAST struct {
	Keyword string `parser:"@ExistenceKeyword ':'"`
	Field   string `parser:"( @Field | @Word )"`
}

// notExprAST represents a negated expression: -term or -field:value
type notExprAST struct {
	Field *fieldExprAST `parser:"'-' ( @@"`
	Term  *termAST      `parser:"    | @@ )"`
}

// fieldExprAST represents a field-qualified expression: field:value or field:>value
type fieldExprAST struct {
	Field    string `parser:"@Field ':'"`
	Operator string `parser:"@( '>''=' | '<''=' | '>' | '<' )?"`
	Value    string `parser:"( @String | @Date | @Field | @Word )"`
}

// termAST represents a simple search term.
type termAST struct {
	Value string `parser:"@String | @Word"`
}

// queryLexer defines the token types for the query language.
var queryLexer = lexer.MustSimple([]lexer.SimpleRule{
	// Existence keywords must come before Field to be matched first
	{Name: "ExistenceKeyword", Pattern: `(has|missing)`},
	{Name: "OrKeyword", Pattern: `(?i)OR`},
	{Name: "Field", Pattern: `(tag|tags|title|path|created|modified|body|status|type)`},
	{Name: "String", Pattern: `"[^"]*"`},
	// Date patterns must come before Word to capture dates properly
	{Name: "Date", Pattern: `\d{4}-\d{2}-\d{2}`},
	{Name: "Word", Pattern: `[^\s:"\-><>=]+`},
	{Name: "Punct", Pattern: `[:\-><>=]`},
	{Name: "Whitespace", Pattern: `\s+`},
})

// queryParser is the Participle parser instance.
var queryParser = participle.MustBuild[queryAST](
	participle.Lexer(queryLexer),
	participle.Elide("Whitespace"),
	participle.UseLookahead(2),
)
