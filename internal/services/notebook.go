package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

// NotebookGroup defines a group of notes with shared properties.
type NotebookGroup struct {
	Name     string         `json:"name"`
	Globs    []string       `json:"globs"`
	Metadata map[string]any `json:"metadata"`
	Template string         `json:"template,omitempty"`
}

// StoredNotebookConfig is what's stored in .opennotes.json.
type StoredNotebookConfig struct {
	Root      string            `json:"root"`
	Name      string            `json:"name"`
	Contexts  []string          `json:"contexts,omitempty"`
	Templates map[string]string `json:"templates,omitempty"`
	Groups    []NotebookGroup   `json:"groups,omitempty"`
}

// NotebookConfig includes runtime-resolved paths.
type NotebookConfig struct {
	StoredNotebookConfig
	Path string `json:"-"` // Path to the config file (not stored)
}

// Notebook represents a loaded notebook with its services.
type Notebook struct {
	Config NotebookConfig
	Notes  *NoteService
}

// NotebookService manages notebook operations.
type NotebookService struct {
	configService *ConfigService
	dbService     *DbService
	log           zerolog.Logger
}

// NewNotebookService creates a notebook service.
func NewNotebookService(cfg *ConfigService, db *DbService) *NotebookService {
	return &NotebookService{
		configService: cfg,
		dbService:     db,
		log:           Log("NotebookService"),
	}
}

// configFilePath returns the config file path for a notebook directory.
func configFilePath(notebookDir string) string {
	return filepath.Join(notebookDir, NotebookConfigFile)
}

// HasNotebook checks if a directory contains a notebook.
func (s *NotebookService) HasNotebook(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(configFilePath(path))
	return err == nil
}

// LoadConfig loads notebook configuration from a directory.
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
		// Create root directory if it doesn't exist
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(rootPath, 0755); mkErr != nil {
				return nil, fmt.Errorf("notes path not found and could not create: %s", rootPath)
			}
		} else {
			return nil, fmt.Errorf("notes path error: %w", err)
		}
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

// Open loads a notebook from the given path.
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

// Create creates a new notebook.
func (s *NotebookService) Create(name, path string, register bool) (*Notebook, error) {
	if path == "" {
		path, _ = os.Getwd()
	}

	notesDir := filepath.Join(path, ".notes")

	config := NotebookConfig{
		StoredNotebookConfig: StoredNotebookConfig{
			Root:     notesDir, // Store absolute path; SaveConfig will convert to relative
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

// Infer discovers notebook from current context.
// Priority: 1. Declared path, 2. Context matching, 3. Ancestor search.
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

// List returns all known notebooks.
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

// MatchContext checks if a path matches any notebook context.
func (n *Notebook) MatchContext(path string) string {
	for _, ctx := range n.Config.Contexts {
		if strings.HasPrefix(path, ctx) {
			return ctx
		}
	}
	return ""
}

// AddContext adds a context path to the notebook.
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

// SaveConfig writes the notebook config to disk.
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
		notebookDir := filepath.Dir(n.Config.Path)
		for _, p := range notebooks {
			if p == notebookDir {
				return nil // Already registered
			}
		}
		configService.Store.Notebooks = append(notebooks, notebookDir)
		return configService.Write(configService.Store)
	}

	return nil
}
