package bleve

import (
	"fmt"
	"strings"
	"time"

	bquery "github.com/blevesearch/bleve/v2/search/query"

	"github.com/zenobi-us/jot/internal/search"
)

// TranslateQuery converts a search.Query AST to a Bleve query.
func TranslateQuery(q *search.Query) (bquery.Query, error) {
	if q == nil || q.IsEmpty() {
		// Empty query matches all documents
		return bquery.NewMatchAllQuery(), nil
	}

	queries := make([]bquery.Query, 0, len(q.Expressions))
	for _, expr := range q.Expressions {
		bq, err := translateExpr(expr)
		if err != nil {
			return nil, err
		}
		queries = append(queries, bq)
	}

	if len(queries) == 1 {
		return queries[0], nil
	}

	// Multiple expressions are ANDed together
	return bquery.NewConjunctionQuery(queries), nil
}

// translateExpr translates a single expression to a Bleve query.
func translateExpr(expr search.Expr) (bquery.Query, error) {
	switch e := expr.(type) {
	case search.TermExpr:
		return translateTermExpr(e)
	case search.FieldExpr:
		return translateFieldExpr(e)
	case search.NotExpr:
		return translateNotExpr(e)
	case search.OrExpr:
		return translateOrExpr(e)
	case search.AndExpr:
		return translateAndExpr(e)
	case search.DateExpr:
		return translateDateExpr(e)
	case search.RangeExpr:
		return translateRangeExpr(e)
	case search.WildcardExpr:
		return translateWildcardExpr(e)
	case search.ExistsExpr:
		return translateExistsExpr(e)
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// translateTermExpr translates a simple search term.
func translateTermExpr(e search.TermExpr) (bquery.Query, error) {
	// Match against all text fields with boosting
	// Use a disjunction with max boost
	queries := []bquery.Query{
		boostQuery(bquery.NewMatchQuery(e.Value), WeightTitle, FieldTitle),
		boostQuery(bquery.NewMatchQuery(e.Value), WeightLead, FieldLead),
		boostQuery(bquery.NewMatchQuery(e.Value), WeightBody, FieldBody),
	}
	return bquery.NewDisjunctionQuery(queries), nil
}

// boostQuery wraps a query to match on a specific field with a boost.
func boostQuery(q *bquery.MatchQuery, boost float64, field string) bquery.Query {
	q.SetField(field)
	q.SetBoost(boost)
	return q
}

// translateFieldExpr translates a field-qualified search.
func translateFieldExpr(e search.FieldExpr) (bquery.Query, error) {
	field := normalizeField(e.Field)

	switch e.Op {
	case search.OpEquals, "":
		// Default: match or term query depending on field
		if field == FieldTags {
			// Tags use term query for exact match
			tq := bquery.NewTermQuery(strings.ToLower(e.Value))
			tq.SetField(field)
			return tq, nil
		}
		mq := bquery.NewMatchQuery(e.Value)
		mq.SetField(field)
		return mq, nil

	case search.OpPrefix:
		pq := bquery.NewPrefixQuery(strings.ToLower(e.Value))
		pq.SetField(field)
		return pq, nil

	case search.OpGt, search.OpGte, search.OpLt, search.OpLte:
		// Numeric/date comparison - delegate to range query
		return translateDateExpr(search.DateExpr(e))

	default:
		return nil, fmt.Errorf("unsupported operator %q for field %q", e.Op, e.Field)
	}
}

// translateNotExpr translates a negated expression.
func translateNotExpr(e search.NotExpr) (bquery.Query, error) {
	inner, err := translateExpr(e.Expr)
	if err != nil {
		return nil, err
	}

	// Boolean query with must and mustNot
	must := []bquery.Query{bquery.NewMatchAllQuery()}
	mustNot := []bquery.Query{inner}
	return bquery.NewBooleanQuery(must, nil, mustNot), nil
}

// translateOrExpr translates an OR expression.
func translateOrExpr(e search.OrExpr) (bquery.Query, error) {
	left, err := translateExpr(e.Left)
	if err != nil {
		return nil, err
	}
	right, err := translateExpr(e.Right)
	if err != nil {
		return nil, err
	}
	return bquery.NewDisjunctionQuery([]bquery.Query{left, right}), nil
}

func translateAndExpr(e search.AndExpr) (bquery.Query, error) {
	queries := make([]bquery.Query, 0, len(e.Expressions))
	for _, expr := range e.Expressions {
		q, err := translateExpr(expr)
		if err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}

	switch len(queries) {
	case 0:
		return bquery.NewMatchAllQuery(), nil
	case 1:
		return queries[0], nil
	default:
		return bquery.NewConjunctionQuery(queries), nil
	}
}

// translateDateExpr translates a date-based filter.
func translateDateExpr(e search.DateExpr) (bquery.Query, error) {
	field := normalizeField(e.Field)

	// Parse the date value
	t, err := parseDate(e.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: %w", e.Value, err)
	}

	var minTime, maxTime time.Time
	inclusive := true

	switch e.Op {
	case search.OpGt:
		minTime = t
		inclusive = false
	case search.OpGte:
		minTime = t
	case search.OpLt:
		maxTime = t
		inclusive = false
	case search.OpLte:
		maxTime = t
	case search.OpEquals, "":
		// Exact date match - match the whole day
		minTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		maxTime = minTime.Add(24*time.Hour - time.Nanosecond)
	default:
		return nil, fmt.Errorf("unsupported date operator %q", e.Op)
	}

	drq := bquery.NewDateRangeInclusiveQuery(minTime, maxTime, &inclusive, &inclusive)
	drq.SetField(field)
	return drq, nil
}

// translateRangeExpr translates a range query.
func translateRangeExpr(e search.RangeExpr) (bquery.Query, error) {
	field := normalizeField(e.Field)

	// Assume dates for now
	startTime, err := parseDate(e.Start)
	if err != nil {
		return nil, fmt.Errorf("invalid range start %q: %w", e.Start, err)
	}
	endTime, err := parseDate(e.End)
	if err != nil {
		return nil, fmt.Errorf("invalid range end %q: %w", e.End, err)
	}

	inclusive := true
	drq := bquery.NewDateRangeInclusiveQuery(startTime, endTime, &inclusive, &inclusive)
	drq.SetField(field)
	return drq, nil
}

// translateWildcardExpr translates a wildcard search.
func translateWildcardExpr(e search.WildcardExpr) (bquery.Query, error) {
	field := normalizeField(e.Field)
	if field == "" {
		field = FieldBody // Default to body
	}

	var pattern string
	switch e.Type {
	case search.WildcardPrefix:
		pattern = strings.TrimSuffix(e.Pattern, "*") + "*"
	case search.WildcardSuffix:
		pattern = "*" + strings.TrimPrefix(e.Pattern, "*")
	case search.WildcardBoth:
		pattern = "*" + strings.Trim(e.Pattern, "*") + "*"
	default:
		pattern = e.Pattern
	}

	wq := bquery.NewWildcardQuery(strings.ToLower(pattern))
	wq.SetField(field)
	return wq, nil
}

// translateExistsExpr translates an existence check to a Bleve query.
// has:<field> (Negated=false) uses a regexp query to match any non-empty value.
// missing:<field> (Negated=true) wraps the exists query in a boolean NOT.
func translateExistsExpr(e search.ExistsExpr) (bquery.Query, error) {
	field := normalizeField(e.Field)

	// Use a regexp that matches any non-empty string to check field existence.
	existsQ := bquery.NewRegexpQuery(".+")
	existsQ.SetField(field)

	if e.Negated {
		// missing:<field> — match all documents that do NOT have this field
		must := []bquery.Query{bquery.NewMatchAllQuery()}
		mustNot := []bquery.Query{existsQ}
		return bquery.NewBooleanQuery(must, nil, mustNot), nil
	}

	// has:<field> — match documents where the field exists
	return existsQ, nil
}

// normalizeField maps query field names to Bleve field names.
func normalizeField(field string) string {
	switch strings.ToLower(field) {
	case "path", "p":
		return FieldPath
	case "title", "t":
		return FieldTitle
	case "body", "content", "b":
		return FieldBody
	case "lead":
		return FieldLead
	case "tag", "tags":
		return FieldTags
	case "created", "date":
		return FieldCreated
	case "modified", "updated":
		return FieldModified
	case "status":
		return FieldMetadata + ".status"
	case "priority", "assignee", "author", "type", "category", "project", "sprint":
		return FieldMetadata + "." + strings.ToLower(field)
	default:
		// Check if it's a metadata field
		if strings.HasPrefix(field, "meta.") || strings.HasPrefix(field, "metadata.") {
			return field
		}
		return field
	}
}

// parseDate parses various date formats.
func parseDate(s string) (time.Time, error) {
	// Try common formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006/01/02",
		"02-01-2006",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	// Try relative dates
	switch strings.ToLower(s) {
	case "today":
		return time.Now().Truncate(24 * time.Hour), nil
	case "yesterday":
		return time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour), nil
	case "this-week":
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return now.Add(-time.Duration(weekday-1) * 24 * time.Hour).Truncate(24 * time.Hour), nil
	case "this-month":
		now := time.Now()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), nil
	}

	return time.Time{}, fmt.Errorf("unrecognized date format: %s", s)
}

// TranslateFindOpts converts FindOpts to a Bleve query.
// This handles both the parsed Query and the convenience filters.
func TranslateFindOpts(opts search.FindOpts) (bquery.Query, error) {
	var queries []bquery.Query

	// Add the main query if present
	if opts.Query != nil && !opts.Query.IsEmpty() {
		q, err := TranslateQuery(opts.Query)
		if err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}

	// Add tag filters - use MatchQuery for analyzed fields
	for _, tag := range opts.Tags {
		mq := bquery.NewMatchQuery(tag)
		mq.SetField(FieldTags)
		queries = append(queries, mq)
	}

	// Add exclude tag filters - use MatchQuery for analyzed fields
	for _, tag := range opts.ExcludeTags {
		mq := bquery.NewMatchQuery(tag)
		mq.SetField(FieldTags)
		must := []bquery.Query{bquery.NewMatchAllQuery()}
		mustNot := []bquery.Query{mq}
		queries = append(queries, bquery.NewBooleanQuery(must, nil, mustNot))
	}

	// Add path prefix filter
	if opts.PathPrefix != "" {
		pq := bquery.NewPrefixQuery(opts.PathPrefix)
		pq.SetField(FieldPath)
		queries = append(queries, pq)
	}

	// Add exclude path filters
	for _, path := range opts.ExcludePaths {
		pq := bquery.NewPrefixQuery(path)
		pq.SetField(FieldPath)
		must := []bquery.Query{bquery.NewMatchAllQuery()}
		mustNot := []bquery.Query{pq}
		queries = append(queries, bquery.NewBooleanQuery(must, nil, mustNot))
	}

	// Add date range filters
	if !opts.CreatedAfter.IsZero() || !opts.CreatedBefore.IsZero() {
		q := buildDateRangeQuery(FieldCreated, opts.CreatedAfter, opts.CreatedBefore)
		queries = append(queries, q)
	}

	if !opts.ModifiedAfter.IsZero() || !opts.ModifiedBefore.IsZero() {
		q := buildDateRangeQuery(FieldModified, opts.ModifiedAfter, opts.ModifiedBefore)
		queries = append(queries, q)
	}

	// If no queries, match all
	if len(queries) == 0 {
		return bquery.NewMatchAllQuery(), nil
	}

	if len(queries) == 1 {
		return queries[0], nil
	}

	return bquery.NewConjunctionQuery(queries), nil
}

// buildDateRangeQuery creates a date range query.
func buildDateRangeQuery(field string, after, before time.Time) bquery.Query {
	inclusive := true
	drq := bquery.NewDateRangeInclusiveQuery(after, before, &inclusive, &inclusive)
	drq.SetField(field)
	return drq
}
