package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/zenobi-us/jot/internal/search"
)

// QueryCondition represents a single search condition for boolean queries.
type QueryCondition struct {
	Type     string // "and", "or", "not"
	Field    string // "data.tag", "data.status", "path", "title", "links-to", "linked-by"
	Operator string // "=" (currently only equality is supported)
	Value    string // user-provided value
}

// AllowedFields is the whitelist of valid fields for security.
// Only these fields can be queried via boolean conditions.
var AllowedFields = map[string]bool{
	"data.tag":      true,
	"data.tags":     true,
	"data.status":   true,
	"data.priority": true,
	"data.assignee": true,
	"data.author":   true,
	"data.type":     true,
	"data.category": true,
	"data.project":  true,
	"data.sprint":   true,
	"path":          true,
	"title":         true,
	"links-to":      true,
	"linked-by":     true,
}

// MaxValueLength is the maximum allowed length for condition values (security).
const MaxValueLength = 1000

// SearchService provides search operations for notes.
type SearchService struct {
	log zerolog.Logger
}

// NewSearchService creates a new search service.
func NewSearchService() *SearchService {
	return &SearchService{
		log: Log("SearchService"),
	}
}

// TextSearch performs exact text matching on notes.
// Searches both content and filepath (case-insensitive).
func (s *SearchService) TextSearch(query string, notes []Note) []Note {
	if query == "" {
		return notes
	}

	var matches []Note
	queryLower := strings.ToLower(query)

	for _, note := range notes {
		// Check content
		if strings.Contains(strings.ToLower(note.Content), queryLower) {
			matches = append(matches, note)
			continue
		}

		// Check filepath
		if strings.Contains(strings.ToLower(note.File.Filepath), queryLower) {
			matches = append(matches, note)
			continue
		}
	}

	s.log.Debug().
		Str("query", query).
		Int("total_notes", len(notes)).
		Int("matches", len(matches)).
		Msg("text search completed")

	return matches
}

// ParseConditions parses CLI flags into QueryConditions.
// andFlags, orFlags, notFlags are arrays of "field=value" strings.
func (s *SearchService) ParseConditions(andFlags, orFlags, notFlags []string) ([]QueryCondition, error) {
	var conditions []QueryCondition

	for _, flag := range andFlags {
		cond, err := s.parseCondition("and", flag)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, cond)
	}

	for _, flag := range orFlags {
		cond, err := s.parseCondition("or", flag)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, cond)
	}

	for _, flag := range notFlags {
		cond, err := s.parseCondition("not", flag)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, cond)
	}

	return conditions, nil
}

// parseCondition parses a single "field=value" string into a QueryCondition.
func (s *SearchService) parseCondition(condType, flag string) (QueryCondition, error) {
	parts := strings.SplitN(flag, "=", 2)
	if len(parts) != 2 {
		return QueryCondition{}, fmt.Errorf("invalid condition format: %s (expected field=value)", flag)
	}

	field, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

	// Validate field (security - whitelist only)
	if !AllowedFields[field] {
		return QueryCondition{}, fmt.Errorf("invalid field: %s (allowed: data.tag, data.status, data.priority, data.assignee, data.author, data.type, data.category, data.project, data.sprint, path, title, links-to, linked-by)", field)
	}

	// Validate value length (security)
	if len(value) > MaxValueLength {
		return QueryCondition{}, fmt.Errorf("value too long (max %d chars)", MaxValueLength)
	}

	// Validate value is not empty
	if value == "" {
		return QueryCondition{}, fmt.Errorf("value cannot be empty for field: %s", field)
	}

	return QueryCondition{
		Type:     condType,
		Field:    field,
		Operator: "=",
		Value:    value,
	}, nil
}

// BuildWhereClause constructs a parameterized SQL WHERE clause from conditions.
// Returns the WHERE clause (without "WHERE"), the parameters, and any error.
// SECURITY: Always use parameterized queries - never concatenate values into SQL.
// Deprecated: Use BuildWhereClauseWithGlob for queries that may include linked-by conditions.
func (s *SearchService) BuildWhereClause(conditions []QueryCondition) (string, []interface{}, error) {
	return s.BuildWhereClauseWithGlob(conditions, "")
}

// BuildWhereClauseWithGlob constructs a parameterized SQL WHERE clause from conditions.
// The notebookGlob parameter is required for linked-by queries to access the source document.
// Returns the WHERE clause (without "WHERE"), the parameters, and any error.
// SECURITY: Always use parameterized queries - never concatenate values into SQL.
func (s *SearchService) BuildWhereClauseWithGlob(conditions []QueryCondition, notebookGlob string) (string, []interface{}, error) {
	if len(conditions) == 0 {
		return "", nil, nil
	}

	var andParts []string
	var orParts []string
	var notParts []string
	var params []interface{}

	for _, cond := range conditions {
		sqlPart, condParams, err := s.buildConditionSQL(cond, notebookGlob)
		if err != nil {
			return "", nil, err
		}

		switch cond.Type {
		case "and":
			andParts = append(andParts, sqlPart)
		case "or":
			orParts = append(orParts, sqlPart)
		case "not":
			notParts = append(notParts, fmt.Sprintf("NOT (%s)", sqlPart))
		}
		params = append(params, condParams...)
	}

	var whereParts []string

	// AND conditions are joined with AND
	if len(andParts) > 0 {
		whereParts = append(whereParts, strings.Join(andParts, " AND "))
	}

	// OR conditions are grouped together with parentheses
	if len(orParts) > 0 {
		whereParts = append(whereParts, fmt.Sprintf("(%s)", strings.Join(orParts, " OR ")))
	}

	// NOT conditions are each negated and joined with AND
	if len(notParts) > 0 {
		whereParts = append(whereParts, strings.Join(notParts, " AND "))
	}

	whereClause := strings.Join(whereParts, " AND ")

	s.log.Debug().
		Str("whereClause", whereClause).
		Int("paramCount", len(params)).
		Int("conditionCount", len(conditions)).
		Msg("built WHERE clause")

	return whereClause, params, nil
}

// buildConditionSQL builds the SQL fragment for a single condition.
// Returns the SQL fragment (with ? placeholders) and the parameter values.
// The notebookGlob is used for linked-by queries that need to access source documents.
func (s *SearchService) buildConditionSQL(cond QueryCondition, notebookGlob string) (string, []interface{}, error) {
	switch {
	case strings.HasPrefix(cond.Field, "data."):
		// Frontmatter field - use metadata map access
		// For DuckDB, we need to access the metadata MAP column
		fieldName := strings.TrimPrefix(cond.Field, "data.")
		// Using DuckDB MAP syntax: metadata[fieldName] = value
		// Note: We use COALESCE to handle NULL gracefully
		sqlPart := "COALESCE(metadata[?], '') = ?"
		return sqlPart, []interface{}{fieldName, cond.Value}, nil

	case cond.Field == "path":
		// Path field - use filepath with wildcard prefix to match anywhere in path
		// This allows users to specify relative paths like "epics/*" that match
		// the full absolute paths like "/home/user/notebook/epics/epic1.md"
		sqlPart := "file_path LIKE ?"
		// Convert glob-like patterns to LIKE patterns
		likePattern := GlobToLike(cond.Value)
		// Add % prefix to match from any point in the path
		if !strings.HasPrefix(likePattern, "%") {
			likePattern = "%" + likePattern
		}
		return sqlPart, []interface{}{likePattern}, nil

	case cond.Field == "title":
		// Title field - check metadata title or use filename
		sqlPart := "(COALESCE(metadata['title'], '') = ? OR file_path LIKE ?)"
		return sqlPart, []interface{}{cond.Value, "%" + cond.Value + "%"}, nil

	case cond.Field == "links-to":
		// Links-to: find documents whose links array contains the target
		// This uses DuckDB's UNNEST and LIKE for glob support
		likePattern := GlobToLike(cond.Value)
		sqlPart := `EXISTS (
			SELECT 1 FROM (
				SELECT unnest(COALESCE(TRY_CAST(metadata['links'] AS VARCHAR[]), ARRAY[]::VARCHAR[])) AS link
			) AS links_table
			WHERE link LIKE ?
		)`
		return sqlPart, []interface{}{likePattern}, nil

	case cond.Field == "linked-by":
		// Linked-by: find documents that the specified source links TO
		// Returns documents that appear in the source's links array
		// Graph: if source --links--> target, linked-by=source returns target
		//
		// We use a subquery that reads the source document and extracts its links,
		// then check if the current file_path matches any of those links.
		if notebookGlob == "" {
			return "", nil, fmt.Errorf("linked-by queries require notebook context")
		}
		likePattern := GlobToLike(cond.Value)
		// Use LIKE with the link to allow for relative path matching
		// The file_path might be absolute while links might be relative
		sqlPart := `EXISTS (
			SELECT 1 FROM (
				SELECT unnest(COALESCE(TRY_CAST(src.metadata['links'] AS VARCHAR[]), ARRAY[]::VARCHAR[])) AS link
				FROM read_markdown(?, include_filepath:=true) AS src
				WHERE src.file_path LIKE ?
			) AS source_links
			WHERE file_path LIKE '%' || source_links.link OR file_path LIKE '%/' || source_links.link
		)`
		return sqlPart, []interface{}{notebookGlob, "%" + likePattern}, nil

	default:
		return "", nil, fmt.Errorf("unsupported field: %s", cond.Field)
	}
}

// GlobToLike converts glob patterns to SQL LIKE patterns.
// ** -> %
// * -> %
// ? -> _
// Also escapes SQL special characters (%, _) before conversion.
func GlobToLike(pattern string) string {
	// First escape SQL special characters that are NOT glob characters
	// We need to be careful here - if user has literal % or _ in their pattern,
	// we need to escape them first before converting glob chars
	result := strings.ReplaceAll(pattern, "%", "\\%")
	result = strings.ReplaceAll(result, "_", "\\_")

	// Now convert glob patterns to SQL LIKE patterns
	// Handle ** first (matches any path including subdirectories)
	result = strings.ReplaceAll(result, "**", "{{DOUBLESTAR}}")
	// Then handle * (matches any characters)
	result = strings.ReplaceAll(result, "*", "%")
	// Replace the placeholder back
	result = strings.ReplaceAll(result, "{{DOUBLESTAR}}", "%")
	// Handle ? (matches single character)
	result = strings.ReplaceAll(result, "?", "_")

	return result
}

// BuildQuery converts QueryCondition structs to search.Query AST.
// This mirrors the pattern of BuildWhereClauseWithGlob but outputs
// a search.Query instead of SQL.
//
// Supported fields:
//   - data.* (metadata fields: tag, status, priority, etc.)
//   - path (with glob pattern support)
//   - title
//
// Unsupported fields (return error):
//   - links-to (requires Phase 5.3 link graph index)
//   - linked-by (requires Phase 5.3 link graph index)
//
// Boolean logic:
//   - AND conditions: All must match (implicit AND)
//   - OR conditions: Any can match (nested OrExpr)
//   - NOT conditions: Must not match (NotExpr wrapper)
//
// Example:
//
//	conditions := []QueryCondition{
//	    {Type: "and", Field: "data.tag", Operator: "=", Value: "work"},
//	    {Type: "and", Field: "data.status", Operator: "=", Value: "active"},
//	}
//	query, err := searchService.BuildQuery(ctx, conditions)
//	// query.Expressions = [FieldExpr{...}, FieldExpr{...}]
func (s *SearchService) BuildQuery(ctx context.Context, conditions []QueryCondition) (*search.Query, error) {
	if len(conditions) == 0 {
		return &search.Query{}, nil
	}

	var andExprs []search.Expr
	var orExprs []search.Expr
	var notExprs []search.Expr

	// Convert each condition to an expression
	for _, cond := range conditions {
		// Check for unsupported link queries
		if cond.Field == "links-to" || cond.Field == "linked-by" {
			return nil, s.buildLinkQueryError(cond.Field)
		}

		// Convert to expression
		expr, err := s.conditionToExpr(cond)
		if err != nil {
			return nil, err
		}

		// Group by type
		switch cond.Type {
		case "and":
			andExprs = append(andExprs, expr)
		case "or":
			orExprs = append(orExprs, expr)
		case "not":
			notExprs = append(notExprs, search.NotExpr{Expr: expr})
		default:
			return nil, fmt.Errorf("unsupported condition type: %s", cond.Type)
		}
	}

	// Build final expression tree
	var allExprs []search.Expr

	// Add AND expressions directly
	allExprs = append(allExprs, andExprs...)

	// Group OR expressions into nested OrExpr
	if len(orExprs) > 0 {
		if len(orExprs) == 1 {
			allExprs = append(allExprs, orExprs[0])
		} else {
			// Build left-to-right OR tree: (a OR (b OR c))
			orExpr := orExprs[0]
			for i := 1; i < len(orExprs); i++ {
				orExpr = search.OrExpr{Left: orExpr, Right: orExprs[i]}
			}
			allExprs = append(allExprs, orExpr)
		}
	}

	// Add NOT expressions
	allExprs = append(allExprs, notExprs...)

	s.log.Debug().
		Int("and_count", len(andExprs)).
		Int("or_count", len(orExprs)).
		Int("not_count", len(notExprs)).
		Int("total_exprs", len(allExprs)).
		Msg("built query from conditions")

	return &search.Query{Expressions: allExprs}, nil
}

// conditionToExpr converts a single QueryCondition to a search.Expr.
func (s *SearchService) conditionToExpr(cond QueryCondition) (search.Expr, error) {
	switch {
	case strings.HasPrefix(cond.Field, "data."):
		// Metadata field: data.tag -> metadata.tag
		return s.buildMetadataExpr(cond)

	case cond.Field == "path":
		// Path field with glob support
		return s.buildPathExpr(cond)

	case cond.Field == "title":
		// Title field
		return search.FieldExpr{
			Field: "title",
			Op:    search.OpEquals,
			Value: cond.Value,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported field: %s (allowed: data.*, path, title)", cond.Field)
	}
}

// buildMetadataExpr builds expression for metadata fields (data.*).
func (s *SearchService) buildMetadataExpr(cond QueryCondition) (search.Expr, error) {
	// Extract metadata field name: data.tag -> tag
	field := strings.TrimPrefix(cond.Field, "data.")

	// Normalize field name
	// data.tags -> data.tag (alias)
	if field == "tags" {
		field = "tag"
	}

	// Build field expression
	// Note: Bleve document has Metadata map[string]any
	// We need to search in metadata.field
	return search.FieldExpr{
		Field: "metadata." + field,
		Op:    search.OpEquals,
		Value: cond.Value,
	}, nil
}

// buildPathExpr builds expression for path field with glob support.
func (s *SearchService) buildPathExpr(cond QueryCondition) (search.Expr, error) {
	value := cond.Value

	// Detect glob patterns
	hasWildcard := strings.Contains(value, "*") || strings.Contains(value, "?")

	if !hasWildcard {
		// Exact path or prefix
		// If ends with /, it's a prefix: projects/ -> projects/*
		if strings.HasSuffix(value, "/") {
			return search.FieldExpr{
				Field: "path",
				Op:    search.OpPrefix,
				Value: value,
			}, nil
		}

		// Exact path match
		return search.FieldExpr{
			Field: "path",
			Op:    search.OpEquals,
			Value: value,
		}, nil
	}

	// Has wildcards - determine type
	// Simple prefix: projects/* -> prefix query (fast)
	// Complex: **/tasks/*.md -> wildcard query (slower)

	if strings.HasSuffix(value, "/*") && !strings.Contains(strings.TrimSuffix(value, "/*"), "*") {
		// Simple prefix pattern: projects/* -> projects/
		prefix := strings.TrimSuffix(value, "*")
		return search.FieldExpr{
			Field: "path",
			Op:    search.OpPrefix,
			Value: prefix,
		}, nil
	}

	// Complex wildcard pattern
	wildcardType := detectWildcardType(value)
	return search.WildcardExpr{
		Field:   "path",
		Pattern: value,
		Type:    wildcardType,
	}, nil
}

// detectWildcardType determines wildcard expression type.
func detectWildcardType(pattern string) search.WildcardType {
	hasPrefix := strings.HasPrefix(pattern, "*")
	hasSuffix := strings.HasSuffix(pattern, "*")

	if hasPrefix && hasSuffix {
		return search.WildcardBoth
	} else if hasPrefix {
		return search.WildcardSuffix
	} else if hasSuffix {
		return search.WildcardPrefix
	}

	// Mid-pattern wildcard: foo*bar
	return search.WildcardBoth
}

// buildLinkQueryError returns a clear error for unsupported link queries.
func (s *SearchService) buildLinkQueryError(field string) error {
	return fmt.Errorf(
		"link queries are not yet supported\n\n"+
			"Field '%s' requires a dedicated link graph index, which is planned for Phase 5.3.\n\n"+
			"Temporary workaround: Use text search or path/title filters:\n"+
			"  jot notes search \"project\"\n"+
			"  jot notes search query --and path=docs/*.md\n\n"+
			"Track implementation progress:\n"+
			"  https://github.com/zenobi-us/jot/issues/XXX\n\n"+
			"Supported fields:\n"+
			"  - Metadata: data.tag, data.status, data.priority, data.assignee,\n"+
			"              data.author, data.type, data.category, data.project, data.sprint\n"+
			"  - Path: path (with glob support: *, **, ?)\n"+
			"  - Title: title",
		field,
	)
}
