# OpenNotes Go Rewrite Specification

## Overview

Rewrite the OpenNotes CLI from Node.js/TypeScript to Go while preserving the existing TypeScript implementation. The Go version will be a complete port maintaining feature parity, architectural patterns, and user-facing behavior.

**Module**: `github.com/zenobi-us/opennotes`

## Directory Structure

```
opennotes/
├── cmd/                           # CLI commands (mirrors src/cmds/)
│   ├── root.go                    # Root command + interceptor logic
│   ├── init.go                    # init command
│   ├── notebook.go                # notebook command group
│   ├── notebook_list.go           # notebook list subcommand
│   ├── notebook_create.go         # notebook create subcommand
│   ├── notebook_register.go       # notebook register subcommand
│   ├── notebook_addcontext.go     # notebook add-context subcommand
│   ├── notes.go                   # notes command group
│   ├── notes_list.go              # notes list subcommand
│   ├── notes_add.go               # notes add subcommand
│   ├── notes_remove.go            # notes remove subcommand
│   └── notes_search.go            # notes search subcommand
├── internal/
│   ├── core/
│   │   ├── schema.go              # Validation helpers
│   │   └── strings.go             # String utilities (dedent, etc.)
│   ├── middleware/
│   │   └── notebook.go            # requireNotebookMiddleware equivalent
│   └── services/
│       ├── config.go              # ConfigService
│       ├── db.go                  # DbService interface + DuckDB impl
│       ├── display.go             # TuiRender equivalent
│       ├── logger.go              # LoggerService
│       ├── notebook.go            # NotebookService
│       └── note.go                # NoteService
├── main.go                        # Entry point
├── go.mod
├── go.sum
└── src/                           # PRESERVED - existing TypeScript
```

## Dependencies

### Required Go Packages

```go
// go.mod additions
require (
    github.com/spf13/cobra v1.8.0           // CLI framework
    github.com/knadh/koanf/v2 v2.1.0        // Configuration
    github.com/knadh/koanf/parsers/json v0.1.0
    github.com/knadh/koanf/providers/file v0.1.0
    github.com/knadh/koanf/providers/env v0.1.0
    github.com/duckdb/duckdb-go/v2 v2.5.4   // Already present
    github.com/rs/zerolog v1.32.0           // Structured logging
    github.com/charmbracelet/glamour v0.6.0 // Markdown rendering
)
```

## Service Implementations

---

### 1. ConfigService (`internal/services/config.go`)

Manages global user configuration with environment variable support.

#### Types

```go
package services

import (
    "os"
    "path/filepath"
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/parsers/json"
    "github.com/knadh/koanf/providers/file"
    "github.com/knadh/koanf/providers/env"
)

// Config represents the global configuration schema
type Config struct {
    // Notebook paths are directories containing .opennotes.json
    Notebooks    []string `koanf:"notebooks" json:"notebooks"`
    // Current notebook path (from env, flag, or stored)
    NotebookPath string   `koanf:"notebookpath" json:"notebookPath,omitempty"`
}

// ConfigService manages configuration loading and persistence
type ConfigService struct {
    k     *koanf.Koanf
    Store Config
    path  string
}
```

#### Implementation Details

```go
// GlobalConfigFile returns platform-specific config path
func GlobalConfigFile() string {
    configDir, _ := os.UserConfigDir()
    return filepath.Join(configDir, "opennotes", "config.json")
}

// NotebookConfigFile is the config filename in notebook directories
const NotebookConfigFile = ".opennotes.json"

// NewConfigService creates and initializes the config service
func NewConfigService() (*ConfigService, error) {
    k := koanf.New(".")
    configPath := GlobalConfigFile()
    
    // 1. Load defaults
    defaultNotebooksDir := filepath.Join(filepath.Dir(configPath), "notebooks")
    defaults := map[string]interface{}{
        "notebooks": []string{defaultNotebooksDir},
    }
    // Load defaults using confmap provider
    
    // 2. Load from config file (if exists)
    if _, err := os.Stat(configPath); err == nil {
        if err := k.Load(file.Provider(configPath), json.Parser()); err != nil {
            // Log warning but continue
        }
    }
    
    // 3. Load environment variables with OPENNOTES_ prefix
    // Transform: OPENNOTES_NOTEBOOK_PATH -> notebookpath
    k.Load(env.Provider("OPENNOTES_", ".", func(s string) string {
        return strings.ToLower(strings.TrimPrefix(s, "OPENNOTES_"))
    }), nil)
    
    // 4. Unmarshal to struct
    var cfg Config
    if err := k.Unmarshal("", &cfg); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    return &ConfigService{
        k:     k,
        Store: cfg,
        path:  configPath,
    }, nil
}

// Write persists the configuration to disk
func (c *ConfigService) Write(cfg Config) error {
    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
        return err
    }
    
    data, err := json.Marshal(cfg)
    if err != nil {
        return err
    }
    
    return os.WriteFile(c.path, data, 0644)
}
```

#### Config Priority (matches TypeScript)

1. Environment variables (`OPENNOTES_*`)
2. Config file (`~/.config/opennotes/config.json`)
3. Defaults

---

### 2. DbService (`internal/services/db.go`)

Manages DuckDB connections with markdown extension.

#### Types

```go
package services

import (
    "sync"
    "github.com/duckdb/duckdb-go/v2"
)

// DbService manages DuckDB database connections
type DbService struct {
    db   *duckdb.DB
    once sync.Once
    mu   sync.Mutex
}

// PreparedStatement wraps a DuckDB prepared statement
type PreparedStatement struct {
    stmt *duckdb.Stmt
}

// StatementResult represents query results
type StatementResult struct {
    rows *duckdb.Rows
}
```

#### Implementation

```go
// NewDbService creates a new database service
func NewDbService() *DbService {
    return &DbService{}
}

// GetDB returns an initialized database connection
func (d *DbService) GetDB() (*duckdb.Conn, error) {
    var initErr error
    
    d.once.Do(func() {
        // Open in-memory database
        db, err := duckdb.Open("")
        if err != nil {
            initErr = err
            return
        }
        d.db = db
        
        // Get connection to install extensions
        conn, err := db.Connect()
        if err != nil {
            initErr = err
            return
        }
        defer conn.Close()
        
        // Install and load markdown extension
        if _, err := conn.Exec("INSTALL markdown FROM community"); err != nil {
            initErr = fmt.Errorf("failed to install markdown: %w", err)
            return
        }
        if _, err := conn.Exec("LOAD markdown"); err != nil {
            initErr = fmt.Errorf("failed to load markdown: %w", err)
            return
        }
    })
    
    if initErr != nil {
        return nil, initErr
    }
    
    return d.db.Connect()
}

// Prepare creates a prepared statement
func (d *DbService) Prepare(conn *duckdb.Conn, query string) (*PreparedStatement, error) {
    stmt, err := conn.Prepare(query)
    if err != nil {
        return nil, err
    }
    return &PreparedStatement{stmt: stmt}, nil
}

// Query executes a query and returns results as maps
func (d *DbService) Query(conn *duckdb.Conn, query string) ([]map[string]interface{}, error) {
    rows, err := conn.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    return rowsToMaps(rows)
}

// Close closes the database
func (d *DbService) Close() error {
    if d.db != nil {
        return d.db.Close()
    }
    return nil
}
```

---

### 3. NotebookService (`internal/services/notebook.go`)

Manages notebook discovery, loading, and operations.

#### Types

```go
package services

// NotebookGroup defines a group of notes with shared properties
type NotebookGroup struct {
    Name     string            `json:"name"`
    Globs    []string          `json:"globs"`
    Metadata map[string]any    `json:"metadata"`
    Template string            `json:"template,omitempty"`
}

// StoredNotebookConfig is what's stored in .opennotes.json
type StoredNotebookConfig struct {
    Root      string            `json:"root"`
    Name      string            `json:"name"`
    Contexts  []string          `json:"contexts,omitempty"`
    Templates map[string]string `json:"templates,omitempty"`
    Groups    []NotebookGroup   `json:"groups,omitempty"`
}

// NotebookConfig includes runtime-resolved paths
type NotebookConfig struct {
    StoredNotebookConfig
    Path string `json:"-"` // Path to the config file (not stored)
}

// Notebook represents a loaded notebook with its services
type Notebook struct {
    Config NotebookConfig
    Notes  *NoteService
}

// NotebookService manages notebook operations
type NotebookService struct {
    configService *ConfigService
    dbService     *DbService
    log           zerolog.Logger
}
```

#### Implementation

```go
// NewNotebookService creates a notebook service
func NewNotebookService(cfg *ConfigService, db *DbService) *NotebookService {
    return &NotebookService{
        configService: cfg,
        dbService:     db,
        log:           log.With().Str("service", "NotebookService").Logger(),
    }
}

// configFilePath returns the config file path for a notebook directory
func configFilePath(notebookDir string) string {
    return filepath.Join(notebookDir, NotebookConfigFile)
}

// HasNotebook checks if a directory contains a notebook
func (s *NotebookService) HasNotebook(path string) bool {
    if path == "" {
        return false
    }
    _, err := os.Stat(configFilePath(path))
    return err == nil
}

// LoadConfig loads notebook configuration from a directory
func (s *NotebookService) LoadConfig(path string) (*NotebookConfig, error) {
    configPath := configFilePath(path)
    
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    var stored StoredNotebookConfig
    if err := json.Unmarshal(data, &stored); err != nil {
        return nil, fmt.Errorf("invalid notebook config: %w", err)
    }
    
    // Resolve root path relative to config location
    rootPath := filepath.Join(path, stored.Root)
    if _, err := os.Stat(rootPath); err != nil {
        return nil, fmt.Errorf("notes path not found: %s", rootPath)
    }
    
    return &NotebookConfig{
        StoredNotebookConfig: StoredNotebookConfig{
            Root:      rootPath, // Now absolute
            Name:      stored.Name,
            Contexts:  stored.Contexts,
            Templates: stored.Templates,
            Groups:    stored.Groups,
        },
        Path: configPath,
    }, nil
}

// Open loads a notebook from the given path
func (s *NotebookService) Open(notebookPath string) (*Notebook, error) {
    config, err := s.LoadConfig(notebookPath)
    if err != nil {
        return nil, err
    }
    
    noteService := NewNoteService(s.configService, s.dbService, config.Root)
    
    return &Notebook{
        Config: *config,
        Notes:  noteService,
    }, nil
}

// Create creates a new notebook
func (s *NotebookService) Create(name, path string, register bool) (*Notebook, error) {
    if path == "" {
        path, _ = os.Getwd()
    }
    
    notesDir := filepath.Join(path, ".notes")
    
    config := NotebookConfig{
        StoredNotebookConfig: StoredNotebookConfig{
            Root:     ".notes",
            Name:     name,
            Contexts: []string{path},
            Groups: []NotebookGroup{
                {
                    Name:     "Default",
                    Globs:    []string{"**/*.md"},
                    Metadata: map[string]any{},
                },
            },
            Templates: map[string]string{},
        },
        Path: configFilePath(path),
    }
    
    // Create notes directory
    if err := os.MkdirAll(notesDir, 0755); err != nil {
        return nil, err
    }
    
    noteService := NewNoteService(s.configService, s.dbService, notesDir)
    notebook := &Notebook{
        Config: config,
        Notes:  noteService,
    }
    
    // Save config
    if err := notebook.SaveConfig(register, s.configService); err != nil {
        return nil, err
    }
    
    return notebook, nil
}

// Infer discovers notebook from current context
// Priority: 1. Declared path, 2. Context matching, 3. Ancestor search
func (s *NotebookService) Infer(cwd string) (*Notebook, error) {
    if cwd == "" {
        cwd, _ = os.Getwd()
    }
    
    // Step 1: Check declared notebook path
    if declaredPath := s.configService.Store.NotebookPath; declaredPath != "" {
        if s.HasNotebook(declaredPath) {
            return s.Open(declaredPath)
        }
    }
    
    // Step 2: Check registered notebooks for context match
    notebooks, _ := s.List(cwd)
    for _, nb := range notebooks {
        if nb.MatchContext(cwd) != "" {
            return nb, nil
        }
    }
    
    // Step 3: Search ancestor directories
    current := cwd
    for current != "/" && current != "" {
        if s.HasNotebook(current) {
            return s.Open(current)
        }
        current = filepath.Dir(current)
    }
    
    return nil, nil // No notebook found
}

// List returns all known notebooks
func (s *NotebookService) List(cwd string) ([]*Notebook, error) {
    var notebooks []*Notebook
    
    // From registered paths
    for _, path := range s.configService.Store.Notebooks {
        if s.HasNotebook(path) {
            if nb, err := s.Open(path); err == nil {
                notebooks = append(notebooks, nb)
            }
        }
    }
    
    // From ancestor directories
    if cwd == "" {
        cwd, _ = os.Getwd()
    }
    current := cwd
    for current != "/" && current != "" {
        if s.HasNotebook(current) {
            if nb, err := s.Open(current); err == nil {
                // Avoid duplicates
                found := false
                for _, existing := range notebooks {
                    if existing.Config.Path == nb.Config.Path {
                        found = true
                        break
                    }
                }
                if !found {
                    notebooks = append(notebooks, nb)
                }
            }
        }
        current = filepath.Dir(current)
    }
    
    return notebooks, nil
}

// Notebook methods

// MatchContext checks if a path matches any notebook context
func (n *Notebook) MatchContext(path string) string {
    for _, ctx := range n.Config.Contexts {
        if strings.HasPrefix(path, ctx) {
            return ctx
        }
    }
    return ""
}

// AddContext adds a context path to the notebook
func (n *Notebook) AddContext(contextPath string, configService *ConfigService) error {
    if contextPath == "" {
        contextPath, _ = os.Getwd()
    }
    
    // Check if already exists
    for _, ctx := range n.Config.Contexts {
        if ctx == contextPath {
            return nil // Already exists
        }
    }
    
    n.Config.Contexts = append(n.Config.Contexts, contextPath)
    return n.SaveConfig(false, configService)
}

// SaveConfig writes the notebook config to disk
func (n *Notebook) SaveConfig(register bool, configService *ConfigService) error {
    configDir := filepath.Dir(n.Config.Path)
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return err
    }
    
    // Calculate relative root for storage
    relRoot, _ := filepath.Rel(configDir, n.Config.Root)
    if relRoot == "" {
        relRoot = "."
    }
    
    stored := StoredNotebookConfig{
        Root:      relRoot,
        Name:      n.Config.Name,
        Contexts:  n.Config.Contexts,
        Templates: n.Config.Templates,
        Groups:    n.Config.Groups,
    }
    
    data, err := json.MarshalIndent(stored, "", "  ")
    if err != nil {
        return err
    }
    
    if err := os.WriteFile(n.Config.Path, data, 0644); err != nil {
        return err
    }
    
    // Register globally if requested
    if register {
        notebooks := configService.Store.Notebooks
        for _, p := range notebooks {
            if p == n.Config.Root {
                return nil // Already registered
            }
        }
        configService.Store.Notebooks = append(notebooks, n.Config.Root)
        return configService.Write(configService.Store)
    }
    
    return nil
}
```

---

### 4. NoteService (`internal/services/note.go`)

Manages note queries via DuckDB.

#### Types

```go
package services

// Note represents a markdown note
type Note struct {
    File struct {
        Filepath string `json:"filepath"`
        Relative string `json:"relative"`
    } `json:"file"`
    Content  string         `json:"content"`
    Metadata map[string]any `json:"metadata"`
}

// NoteService provides note query operations
type NoteService struct {
    configService *ConfigService
    dbService     *DbService
    notebookPath  string
    log           zerolog.Logger
}
```

#### Implementation

```go
// NewNoteService creates a note service for a notebook
func NewNoteService(cfg *ConfigService, db *DbService, notebookPath string) *NoteService {
    return &NoteService{
        configService: cfg,
        dbService:     db,
        notebookPath:  notebookPath,
        log:           log.With().Str("service", "NoteService").Logger(),
    }
}

// SearchNotes returns all notes in the notebook
func (s *NoteService) SearchNotes(query string) ([]Note, error) {
    if s.notebookPath == "" {
        return nil, fmt.Errorf("no notebook selected")
    }
    
    conn, err := s.dbService.GetDB()
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    
    glob := filepath.Join(s.notebookPath, "**", "*.md")
    
    sqlQuery := `SELECT * FROM read_markdown($1, include_filepath:=true)`
    rows, err := conn.Query(sqlQuery, glob)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var notes []Note
    for rows.Next() {
        var filePath string
        var metadata map[string]any
        var content string
        
        if err := rows.Scan(&filePath, &metadata, &content); err != nil {
            continue
        }
        
        relative := strings.TrimPrefix(filePath, s.notebookPath+"/")
        
        notes = append(notes, Note{
            File: struct {
                Filepath string `json:"filepath"`
                Relative string `json:"relative"`
            }{
                Filepath: filePath,
                Relative: relative,
            },
            Content:  content,
            Metadata: metadata,
        })
    }
    
    return notes, nil
}

// Count returns the number of notes in the notebook
func (s *NoteService) Count() (int, error) {
    if s.notebookPath == "" {
        return 0, nil
    }
    
    conn, err := s.dbService.GetDB()
    if err != nil {
        return 0, err
    }
    defer conn.Close()
    
    glob := filepath.Join(s.notebookPath, "**", "*.md")
    
    var count int
    row := conn.QueryRow(`SELECT COUNT(*) FROM read_markdown($1)`, glob)
    if err := row.Scan(&count); err != nil {
        return 0, err
    }
    
    return count, nil
}

// Query executes a raw SQL query
func (s *NoteService) Query(sql string) ([]map[string]any, error) {
    conn, err := s.dbService.GetDB()
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    
    return s.dbService.Query(conn, sql)
}
```

---

### 5. LoggerService (`internal/services/logger.go`)

Structured logging with zerolog.

```go
package services

import (
    "os"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger
func InitLogger() {
    // Check DEBUG env var
    debug := os.Getenv("DEBUG") != ""
    level := zerolog.InfoLevel
    if debug {
        level = zerolog.DebugLevel
    }
    
    // Check LOG_LEVEL env var
    if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
        if parsed, err := zerolog.ParseLevel(lvl); err == nil {
            level = parsed
        }
    }
    
    zerolog.SetGlobalLevel(level)
    
    // Pretty console output
    log.Logger = zerolog.New(zerolog.ConsoleWriter{
        Out: os.Stderr,
    }).With().Timestamp().Logger()
}

// Log returns a child logger with namespace
func Log(namespace string) zerolog.Logger {
    return log.With().Str("namespace", namespace).Logger()
}
```

---

### 6. Display Service (`internal/services/display.go`)

Markdown rendering for terminal output.

```go
package services

import (
    "github.com/charmbracelet/glamour"
    "text/template"
    "bytes"
)

// TuiRender renders a template with context and formats as markdown
func TuiRender(tmpl string, ctx interface{}) (string, error) {
    // Parse and execute Go template
    t, err := template.New("tui").Parse(tmpl)
    if err != nil {
        return "", err
    }
    
    var buf bytes.Buffer
    if err := t.Execute(&buf, ctx); err != nil {
        return "", err
    }
    
    // Render markdown for terminal
    renderer, err := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(100),
    )
    if err != nil {
        return buf.String(), nil // Fallback to plain
    }
    
    rendered, err := renderer.Render(buf.String())
    if err != nil {
        return buf.String(), nil
    }
    
    return rendered, nil
}
```

---

## CLI Commands (`cmd/`)

### Root Command (`cmd/root.go`)

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/zenobi-us/opennotes/internal/services"
)

var (
    cfgService      *services.ConfigService
    dbService       *services.DbService
    notebookService *services.NotebookService
)

var rootCmd = &cobra.Command{
    Use:   "opennotes",
    Short: "A CLI for managing markdown-based notes",
    Long:  `OpenNotes is a CLI tool for managing your markdown-based notes organized in notebooks.`,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Initialize services (interceptor equivalent)
        services.InitLogger()
        
        var err error
        cfgService, err = services.NewConfigService()
        if err != nil {
            return err
        }
        
        dbService = services.NewDbService()
        notebookService = services.NewNotebookService(cfgService, dbService)
        
        return nil
    },
    PersistentPostRun: func(cmd *cobra.Command, args []string) {
        // Cleanup
        if dbService != nil {
            dbService.Close()
        }
    },
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    // Global flags
    rootCmd.PersistentFlags().String("notebook", "", "Path to notebook")
}
```

### Init Command (`cmd/init.go`)

```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize opennotes configuration",
    RunE: func(cmd *cobra.Command, args []string) error {
        if err := cfgService.Write(cfgService.Store); err != nil {
            return fmt.Errorf("failed to initialize: %w", err)
        }
        
        fmt.Printf("OpenNotes initialized at %s\n", services.GlobalConfigFile())
        return nil
    },
}

func init() {
    rootCmd.AddCommand(initCmd)
}
```

### Notebook Commands (`cmd/notebook.go`, etc.)

```go
// cmd/notebook.go
package cmd

import "github.com/spf13/cobra"

var notebookCmd = &cobra.Command{
    Use:   "notebook",
    Short: "Manage notebooks",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Default: show current notebook info
        nb, err := notebookService.Infer("")
        if err != nil || nb == nil {
            return displayCreateFirstNotebook()
        }
        return displayNotebookInfo(nb)
    },
}

func init() {
    rootCmd.AddCommand(notebookCmd)
}

// cmd/notebook_list.go
var notebookListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all notebooks",
    RunE: func(cmd *cobra.Command, args []string) error {
        notebooks, err := notebookService.List("")
        if err != nil {
            return err
        }
        return displayNotebookList(notebooks)
    },
}

func init() {
    notebookCmd.AddCommand(notebookListCmd)
}

// cmd/notebook_create.go
var notebookCreateCmd = &cobra.Command{
    Use:   "create",
    Short: "Create a new notebook",
    RunE: func(cmd *cobra.Command, args []string) error {
        name, _ := cmd.Flags().GetString("name")
        path, _ := cmd.Flags().GetString("path")
        global, _ := cmd.Flags().GetBool("global")
        
        nb, err := notebookService.Create(name, path, global)
        if err != nil {
            return err
        }
        
        return displayNotebookCreated(nb)
    },
}

func init() {
    notebookCreateCmd.Flags().String("name", "", "Notebook name (required)")
    notebookCreateCmd.Flags().String("path", "", "Notebook path (default: cwd)")
    notebookCreateCmd.Flags().Bool("global", false, "Register globally")
    notebookCreateCmd.MarkFlagRequired("name")
    notebookCmd.AddCommand(notebookCreateCmd)
}
```

### Notes Commands (`cmd/notes.go`, etc.)

```go
// cmd/notes.go
package cmd

import "github.com/spf13/cobra"

var notesCmd = &cobra.Command{
    Use:   "notes",
    Short: "Manage notes",
}

func init() {
    rootCmd.AddCommand(notesCmd)
}

// cmd/notes_list.go
var notesListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all notes in the notebook",
    RunE: func(cmd *cobra.Command, args []string) error {
        nb, err := requireNotebook(cmd)
        if err != nil {
            return err
        }
        
        notes, err := nb.Notes.SearchNotes("")
        if err != nil {
            return err
        }
        
        return displayNoteList(notes)
    },
}

func init() {
    notesCmd.AddCommand(notesListCmd)
}

// cmd/notes_search.go
var notesSearchCmd = &cobra.Command{
    Use:   "search [query]",
    Short: "Search notes",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        nb, err := requireNotebook(cmd)
        if err != nil {
            return err
        }
        
        notes, err := nb.Notes.SearchNotes(args[0])
        if err != nil {
            return err
        }
        
        return displayNoteList(notes)
    },
}

func init() {
    notesCmd.AddCommand(notesSearchCmd)
}
```

---

## Middleware (`internal/middleware/notebook.go`)

```go
package middleware

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/zenobi-us/opennotes/internal/services"
)

// RequireNotebook ensures a notebook is available
func RequireNotebook(cmd *cobra.Command, notebookService *services.NotebookService) (*services.Notebook, error) {
    // Check --notebook flag first
    notebookPath, _ := cmd.Flags().GetString("notebook")
    
    if notebookPath != "" {
        return notebookService.Open(notebookPath)
    }
    
    // Try to infer from context
    nb, err := notebookService.Infer("")
    if err != nil {
        return nil, err
    }
    
    if nb == nil {
        return nil, fmt.Errorf("no notebook found. Create one with: opennotes notebook create --name \"My Notebook\"")
    }
    
    return nb, nil
}
```

---

## Entry Point (`main.go`)

```go
package main

import (
    "fmt"
    "os"
    "github.com/zenobi-us/opennotes/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

---

## Implementation Phases

### Phase 1: Core Infrastructure
- [ ] Set up Go project structure
- [ ] Implement LoggerService
- [ ] Implement ConfigService with Koanf
- [ ] Implement DbService with duckdb-go
- [ ] Add root command with Cobra

### Phase 2: Notebook Management
- [ ] Implement NotebookService (types, load, save)
- [ ] Implement notebook discovery (Infer, List)
- [ ] Add notebook commands (notebook, list, create, register, add-context)
- [ ] Implement requireNotebook middleware

### Phase 3: Note Operations
- [ ] Implement NoteService (search, count, query)
- [ ] Add notes commands (list, search, add, remove)
- [ ] Implement display/TUI templates

### Phase 4: Polish
- [ ] Error handling and validation
- [ ] Help text and documentation
- [ ] Integration tests
- [ ] Build configuration (Makefile/mise tasks)

---

## Testing Strategy

```go
// Example test structure
// internal/services/config_test.go
package services_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/zenobi-us/opennotes/internal/services"
)

func TestConfigService_LoadDefaults(t *testing.T) {
    cfg, err := services.NewConfigService()
    assert.NoError(t, err)
    assert.NotEmpty(t, cfg.Store.Notebooks)
}

func TestConfigService_EnvOverride(t *testing.T) {
    t.Setenv("OPENNOTES_NOTEBOOK_PATH", "/tmp/test")
    
    cfg, err := services.NewConfigService()
    assert.NoError(t, err)
    assert.Equal(t, "/tmp/test", cfg.Store.NotebookPath)
}
```

---

## Build & Run

```bash
# Build
go build -o dist/opennotes .

# Run
./dist/opennotes --help
./dist/opennotes init
./dist/opennotes notebook create --name "Test"
./dist/opennotes notes list

# Test
go test ./...
```

---

## Notes

1. **Preserve TypeScript**: The `src/` directory remains untouched
2. **Feature parity**: All existing commands must work identically
3. **Config compatibility**: Go version must read/write same config format
4. **DuckDB markdown**: Uses same `read_markdown()` extension queries
5. **Error messages**: Match TypeScript output format for consistency
