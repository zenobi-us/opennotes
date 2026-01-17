package services

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/rs/zerolog"

	// DuckDB driver
	_ "github.com/duckdb/duckdb-go/v2"
)

// DbService manages DuckDB database connections.
type DbService struct {
	db       *sql.DB
	readOnly *sql.DB
	once     sync.Once
	roOnce   sync.Once
	mu       sync.Mutex
	log      zerolog.Logger
}

// NewDbService creates a new database service.
func NewDbService() *DbService {
	return &DbService{
		log: Log("DbService"),
	}
}

// GetDB returns an initialized database connection.
// The connection is lazily initialized on first call and reused thereafter.
func (d *DbService) GetDB(ctx context.Context) (*sql.DB, error) {
	var initErr error

	d.once.Do(func() {
		d.log.Debug().Msg("initializing database")

		// Open in-memory database
		db, err := sql.Open("duckdb", "")
		if err != nil {
			initErr = fmt.Errorf("failed to open database: %w", err)
			return
		}
		d.db = db

		// Install and load markdown extension
		d.log.Debug().Msg("installing markdown extension")
		if _, err := db.ExecContext(ctx, "INSTALL markdown FROM community"); err != nil {
			initErr = fmt.Errorf("failed to install markdown extension: %w", err)
			return
		}

		d.log.Debug().Msg("loading markdown extension")
		if _, err := db.ExecContext(ctx, "LOAD markdown"); err != nil {
			initErr = fmt.Errorf("failed to load markdown extension: %w", err)
			return
		}

		d.log.Debug().Msg("database initialized")
	})

	if initErr != nil {
		return nil, initErr
	}

	return d.db, nil
}

// GetReadOnlyDB returns a separate read-only database connection.
// This is used for executing user-provided SQL queries safely.
// The connection is lazily initialized on first call and reused thereafter.
func (d *DbService) GetReadOnlyDB(ctx context.Context) (*sql.DB, error) {
	var initErr error

	d.roOnce.Do(func() {
		d.log.Debug().Msg("initializing read-only database connection")

		// Open separate in-memory database
		db, err := sql.Open("duckdb", "")
		if err != nil {
			initErr = fmt.Errorf("failed to open read-only database: %w", err)
			return
		}

		// Install and load markdown extension
		d.log.Debug().Msg("installing markdown extension on read-only connection")
		if _, err := db.ExecContext(ctx, "INSTALL markdown FROM community"); err != nil {
			initErr = fmt.Errorf("failed to install markdown extension on read-only connection: %w", err)
			db.Close()
			return
		}

		d.log.Debug().Msg("loading markdown extension on read-only connection")
		if _, err := db.ExecContext(ctx, "LOAD markdown"); err != nil {
			initErr = fmt.Errorf("failed to load markdown extension on read-only connection: %w", err)
			db.Close()
			return
		}

		d.readOnly = db
		d.log.Debug().Msg("read-only database initialized")
	})

	if initErr != nil {
		return nil, initErr
	}

	return d.readOnly, nil
}

// Query executes a query and returns results as maps.
func (d *DbService) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	db, err := d.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			d.log.Warn().Err(err).Msg("failed to close rows")
		}
	}()

	return rowsToMaps(rows)
}

// rowsToMaps converts sql.Rows to a slice of maps.
func rowsToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create slice of interface{} to hold values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Create map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Close closes both database connections.
func (d *DbService) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var errs []error

	if d.db != nil {
		d.log.Debug().Msg("closing main database")
		if err := d.db.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if d.readOnly != nil {
		d.log.Debug().Msg("closing read-only database")
		if err := d.readOnly.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close database(s): %v", errs)
	}

	return nil
}
