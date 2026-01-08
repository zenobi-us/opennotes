package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
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
	defer rows.Close()

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

// Query executes a raw SQL query.
func (s *NoteService) Query(ctx context.Context, sql string) ([]map[string]any, error) {
	return s.dbService.Query(ctx, sql)
}
