package services

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/zenobi-us/opennotes/internal/core"
)

// Note represents a markdown note.
type Note struct {
	File struct {
		Filepath string `json:"filepath"`
		Relative string `json:"relative"`
	} `json:"file"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
}

// DisplayName returns the display name for the note.
// Priority:
// 1. metadata["title"] if available
// 2. Slugified filename (without extension)
func (n *Note) DisplayName() string {
	// Check for title in metadata
	if title, ok := n.Metadata["title"]; ok {
		if titleStr, ok := title.(string); ok && titleStr != "" {
			return titleStr
		}
	}

	// Fallback to slugified filename
	filename := path.Base(n.File.Relative)
	// Remove .md extension
	filename = strings.TrimSuffix(filename, ".md")
	return core.Slugify(filename)
}

// NoteService provides note query operations.
type NoteService struct {
	configService *ConfigService
	dbService     *DbService
	notebookPath  string
	log           zerolog.Logger
}

// NewNoteService creates a note service for a notebook.
func NewNoteService(cfg *ConfigService, db *DbService, notebookPath string) *NoteService {
	return &NoteService{
		configService: cfg,
		dbService:     db,
		notebookPath:  notebookPath,
		log:           Log("NoteService"),
	}
}

// SearchNotes returns all notes in the notebook matching the query.
func (s *NoteService) SearchNotes(ctx context.Context, query string) ([]Note, error) {
	if s.notebookPath == "" {
		return nil, fmt.Errorf("no notebook selected")
	}

	db, err := s.dbService.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	glob := filepath.Join(s.notebookPath, "**", "*.md")
	s.log.Debug().Str("glob", glob).Str("query", query).Msg("searching notes")

	// Use DuckDB's read_markdown function with filepath included
	sqlQuery := `SELECT * FROM read_markdown(?, include_filepath:=true)`
	rows, err := db.QueryContext(ctx, sqlQuery, glob)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.log.Warn().Err(err).Msg("failed to close rows")
		}
	}()

	var notes []Note
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		// Create slice of interface{} to hold values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			s.log.Warn().Err(err).Msg("failed to scan row")
			continue
		}

		// Map columns to Note struct
		note := Note{
			Metadata: make(map[string]any),
		}

		for i, col := range columns {
			val := values[i]
			switch col {
			case "filepath", "file_path", "filename":
				if v, ok := val.(string); ok {
					note.File.Filepath = v
					note.File.Relative = strings.TrimPrefix(v, s.notebookPath+"/")
				}
			case "content", "body":
				if v, ok := val.(string); ok {
					note.Content = v
				}
			case "metadata":
				// metadata column contains a DuckDB MAP with frontmatter data
				// The type might be duckdb.Map or map[any]any
				// Try to handle it as a map type by using reflection if needed
				rv := reflect.ValueOf(val)
				if rv.Kind() == reflect.Map {
					// It's some kind of map - iterate over it
					for _, key := range rv.MapKeys() {
						if keyStr, ok := key.Interface().(string); ok {
							note.Metadata[keyStr] = rv.MapIndex(key).Interface()
						}
					}
				} else if v, ok := val.(map[any]any); ok {
					for k, val := range v {
						if keyStr, ok := k.(string); ok {
							note.Metadata[keyStr] = val
						}
					}
				} else if v, ok := val.(map[string]any); ok {
					note.Metadata = v
				}
			default:
				note.Metadata[col] = val
			}
		}

		// Filter by query if provided
		if query != "" {
			// Simple contains check on content and filepath
			if !strings.Contains(strings.ToLower(note.Content), strings.ToLower(query)) &&
				!strings.Contains(strings.ToLower(note.File.Filepath), strings.ToLower(query)) {
				continue
			}
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	s.log.Debug().Int("count", len(notes)).Msg("notes found")
	return notes, nil
}

// Count returns the number of notes in the notebook.
func (s *NoteService) Count(ctx context.Context) (int, error) {
	if s.notebookPath == "" {
		return 0, nil
	}

	db, err := s.dbService.GetDB(ctx)
	if err != nil {
		return 0, err
	}

	glob := filepath.Join(s.notebookPath, "**", "*.md")

	var count int
	row := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM read_markdown(?)`, glob)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// ValidateSQL validates a user-provided SQL query for safety.
// Only SELECT and WITH (CTE) queries are allowed.
// Dangerous keywords (DROP, DELETE, UPDATE, etc.) are blocked.
func ValidateSQL(query string) error {
	// Trim and normalize to uppercase
	normalized := strings.TrimSpace(strings.ToUpper(query))

	if normalized == "" {
		return fmt.Errorf("query cannot be empty")
	}

	// Check query type - only SELECT and WITH allowed
	if !strings.HasPrefix(normalized, "SELECT") && !strings.HasPrefix(normalized, "WITH") {
		return fmt.Errorf("only SELECT queries are allowed")
	}

	// Dangerous keywords blocklist - check with word boundaries
	// Split query by spaces and other delimiters to find keywords
	tokens := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '(' || r == ')' ||
			r == ',' || r == ';' || r == '=' || r == '<' || r == '>'
	})

	dangerous := map[string]bool{
		"DROP":     true,
		"DELETE":   true,
		"UPDATE":   true,
		"INSERT":   true,
		"ALTER":    true,
		"CREATE":   true,
		"TRUNCATE": true,
		"REPLACE":  true,
		"ATTACH":   true,
		"DETACH":   true,
		"PRAGMA":   true,
	}

	for _, token := range tokens {
		if dangerous[token] {
			return fmt.Errorf("keyword '%s' is not allowed", token)
		}
	}

	return nil
}

// ExecuteSQLSafe executes a user-provided SQL query safely.
// Validates the query, executes with a 30-second timeout on a read-only connection,
// and returns results as maps.
func (s *NoteService) ExecuteSQLSafe(ctx context.Context, query string) ([]map[string]any, error) {
	// 1. Validate query
	if err := ValidateSQL(query); err != nil {
		s.log.Warn().Err(err).Msg("SQL query validation failed")
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	// 2. Get read-only connection
	db, err := s.dbService.GetReadOnlyDB(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get read-only database connection")
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 3. Create context with 30-second timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.log.Debug().Str("query", query).Msg("executing SQL query")

	// 4. Execute query
	rows, err := db.QueryContext(timeoutCtx, query)
	if err != nil {
		s.log.Error().Err(err).Str("query", query).Msg("query execution failed")
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.log.Warn().Err(err).Msg("failed to close result rows")
		}
	}()

	// 5. Convert rows to maps
	results, err := rowsToMaps(rows)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to scan query results")
		return nil, fmt.Errorf("failed to read results: %w", err)
	}

	s.log.Debug().Int("rows", len(results)).Msg("query executed successfully")
	return results, nil
}

// Query executes a raw SQL query.
func (s *NoteService) Query(ctx context.Context, sql string) ([]map[string]any, error) {
	return s.dbService.Query(ctx, sql)
}
