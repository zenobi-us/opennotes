package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/zenobi-us/jot/internal/search"
	"github.com/zenobi-us/jot/internal/search/bleve"
	"gopkg.in/yaml.v3"
)

// DefaultFilenameFormat is the default gomplate template for generating note filenames.
const DefaultFilenameFormat = "{{ .title | slug }}.md"

// DefaultContentTemplate is the default template for generating initial note content.
// Uses Go template syntax with jot functions available.
const DefaultContentTemplate = `---
title: {{ .title }}
created_at: {{ jot.Now "2006-01-02T15:04:05Z07:00" }}
---

# {{ .title }}
`

// NotebookGroup defines a group of notes with shared properties.
type NotebookGroup struct {
	Name           string         `json:"name"`
	Globs          []string       `json:"globs"`
	Metadata       map[string]any `json:"metadata"`
	Template       string         `json:"template,omitempty"`
	WorkflowID     string         `json:"workflow_id,omitempty"`
	Type           string         `json:"type,omitempty"`
	Aliases        []string       `json:"aliases,omitempty"`
	FilenameFormat string         `json:"filename_format,omitempty"`
}

// GetFilenameFormat returns the filename format template, falling back to DefaultFilenameFormat if empty.
func (g *NotebookGroup) GetFilenameFormat() string {
	if g.FilenameFormat == "" {
		return DefaultFilenameFormat
	}
	return g.FilenameFormat
}

// GetTemplate returns the content template, falling back to DefaultContentTemplate if empty.
func (g *NotebookGroup) GetTemplate() string {
	if g.Template == "" {
		return DefaultContentTemplate
	}
	return g.Template
}

// ValidateFilenameFormat checks that the filename format is valid.
// It must end with .md and not contain path separators.
func (g *NotebookGroup) ValidateFilenameFormat() error {
	if g.FilenameFormat == "" {
		return nil // Empty is valid, will use default
	}
	if !strings.HasSuffix(g.FilenameFormat, ".md") {
		return fmt.Errorf("filename_format must end with .md, got: %s", g.FilenameFormat)
	}
	if strings.Contains(g.FilenameFormat, "/") || strings.Contains(g.FilenameFormat, "\\") {
		return fmt.Errorf("filename_format must not contain path separators, got: %s", g.FilenameFormat)
	}
	return nil
}

type legacyGroupWorkflowBinding struct {
	ID string `json:"id"`
}

type notebookGroupRaw struct {
	Name           string                      `json:"name"`
	Globs          []string                    `json:"globs"`
	Metadata       map[string]any              `json:"metadata"`
	Template       string                      `json:"template,omitempty"`
	WorkflowID     string                      `json:"workflow_id,omitempty"`
	Workflow       *legacyGroupWorkflowBinding `json:"workflow,omitempty"`
	Type           string                      `json:"type,omitempty"`
	Aliases        []string                    `json:"aliases,omitempty"`
	FilenameFormat string                      `json:"filename_format,omitempty"`
}

// UnmarshalJSON supports the canonical workflow_id shape and legacy workflow object shape.
func (g *NotebookGroup) UnmarshalJSON(data []byte) error {
	var raw notebookGroupRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	workflowID := raw.WorkflowID
	if workflowID == "" && raw.Workflow != nil {
		workflowID = raw.Workflow.ID
	}

	g.Name = raw.Name
	g.Globs = raw.Globs
	g.Metadata = raw.Metadata
	g.Template = raw.Template
	g.WorkflowID = workflowID
	g.Type = raw.Type
	g.Aliases = raw.Aliases
	g.FilenameFormat = raw.FilenameFormat

	return nil
}

// Notebook config version constants.
const (
	// NotebookConfigVersionBootstrap is assigned to legacy notebooks that do not
	// yet declare a config_version value.
	NotebookConfigVersionBootstrap Version = 1
)

// StoredNotebookConfig is what's stored in .jot.json.
type StoredNotebookConfig struct {
	ConfigVersion Version                       `json:"config_version,omitempty"`
	Root          string                        `json:"root"`
	Name          string                        `json:"name"`
	Contexts      []string                      `json:"contexts,omitempty"`
	Templates     map[string]string             `json:"templates,omitempty"`
	Groups        []NotebookGroup               `json:"groups,omitempty"`
	Workflows     map[string]WorkflowDefinition `json:"workflows,omitempty"`
	DefaultGroup  string                        `json:"default_group,omitempty"`
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
type cachedNotebook struct {
	state    notebookIndexState
	notebook *Notebook
}

type NotebookService struct {
	configService   *ConfigService
	log             zerolog.Logger
	cachedNotebooks map[string]cachedNotebook
}

// NewNotebookService creates a notebook service.
func NewNotebookService(cfg *ConfigService) *NotebookService {
	return &NotebookService{
		configService:   cfg,
		log:             Log("NotebookService"),
		cachedNotebooks: make(map[string]cachedNotebook),
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

	if stored.ConfigVersion == 0 {
		stored.ConfigVersion = NotebookConfigVersionBootstrap
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

	for name, wf := range stored.Workflows {
		if wf.InitialState == "" {
			return nil, fmt.Errorf("workflow %q missing initial_state", name)
		}

		if _, ok := wf.States[wf.InitialState]; !ok {
			return nil, fmt.Errorf("workflow %q initial_state %q not in states", name, wf.InitialState)
		}
	}

	return &NotebookConfig{
		StoredNotebookConfig: StoredNotebookConfig{
			ConfigVersion: stored.ConfigVersion,
			Root:          rootPath, // Now absolute
			Name:          stored.Name,
			Contexts:      stored.Contexts,
			Templates:     stored.Templates,
			Groups:        stored.Groups,
			Workflows:     stored.Workflows,
			DefaultGroup:  stored.DefaultGroup,
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

	state, err := buildNotebookIndexState(config.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to scan notebook state: %w", err)
	}

	if cached, ok := s.cachedNotebooks[config.Root]; ok && notebookIndexStateEqual(cached.state, state) {
		return cached.notebook, nil
	}

	// Create Bleve index for this notebook
	idx, err := s.createIndex(config.Root, state.Files)
	if err != nil {
		return nil, fmt.Errorf("failed to create search index: %w", err)
	}

	noteService := NewNoteService(s.configService, idx, config.Root)

	semanticIdx, err := s.createSemanticIndex(config.Root)
	if err != nil {
		s.log.Warn().Err(err).Msg("failed to initialize semantic backend; using noop fallback")
		semanticIdx = NewNoopSemanticIndex()
	}
	noteService.SetSemanticIndex(semanticIdx)

	notebook := &Notebook{
		Config: *config,
		Notes:  noteService,
	}
	s.cachedNotebooks[config.Root] = cachedNotebook{state: state, notebook: notebook}
	return notebook, nil
}

type notebookIndexedFile struct {
	Path        string
	Size        int64
	ModUnixNano int64
}

type notebookIndexState struct {
	Files []notebookIndexedFile
}

func notebookIndexStateEqual(a, b notebookIndexState) bool {
	if len(a.Files) != len(b.Files) {
		return false
	}
	for i := range a.Files {
		if a.Files[i] != b.Files[i] {
			return false
		}
	}
	return true
}

// createIndex creates and populates a Bleve index for the notebook.
func (s *NotebookService) createIndex(notebookRoot string, files []notebookIndexedFile) (search.Index, error) {
	storage := bleve.MemStorage()
	idx, err := bleve.NewIndex(storage, bleve.Options{InMemory: true})
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	fs := afero.NewOsFs()
	ctx := context.Background()
	for _, file := range files {
		fullPath := filepath.Join(notebookRoot, file.Path)
		content, readErr := afero.ReadFile(fs, fullPath)
		if readErr != nil {
			s.log.Warn().Err(readErr).Str("path", file.Path).Msg("failed to read file")
			continue
		}

		metadata, body := parseFrontmatter(content)
		modTime := time.Unix(0, file.ModUnixNano)
		doc := search.Document{
			Path:     file.Path,
			Title:    extractTitle(metadata),
			Body:     body,
			Lead:     extractLead(body),
			Tags:     extractTags(metadata),
			Metadata: metadata,
			Created:  extractTime(metadata, "created", modTime),
			Modified: extractTime(metadata, "modified", modTime),
		}

		if addErr := idx.Add(ctx, doc); addErr != nil {
			s.log.Warn().Err(addErr).Str("path", file.Path).Msg("failed to index document")
		}
	}

	return idx, nil
}

func buildNotebookIndexState(notebookRoot string) (notebookIndexState, error) {
	fs := afero.NewOsFs()
	files := make([]notebookIndexedFile, 0, 64)

	err := afero.Walk(fs, notebookRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".jot" {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}

		relPath, relErr := filepath.Rel(notebookRoot, path)
		if relErr != nil {
			return relErr
		}

		files = append(files, notebookIndexedFile{
			Path:        relPath,
			Size:        info.Size(),
			ModUnixNano: info.ModTime().UnixNano(),
		})
		return nil
	})
	if err != nil {
		return notebookIndexState{}, err
	}

	slices.SortFunc(files, func(a, b notebookIndexedFile) int {
		return strings.Compare(a.Path, b.Path)
	})

	return notebookIndexState{Files: files}, nil
}

// createSemanticIndex initializes semantic retrieval backend for a notebook.
// Phase 3 starts with a safe noop backend and can be swapped with a real
// semantic backend implementation without changing callers.
func (s *NotebookService) createSemanticIndex(notebookRoot string) (SemanticIndex, error) {
	s.log.Debug().Str("notebookRoot", notebookRoot).Msg("semantic backend not configured; using noop fallback")
	return NewNoopSemanticIndex(), nil
}

// Helper functions for extracting metadata

func extractTitle(metadata map[string]any) string {
	if title, ok := metadata["title"].(string); ok && title != "" {
		return title
	}
	return ""
}

func extractTags(metadata map[string]any) []string {
	// Handle both "tag" and "tags" fields
	if tag, ok := metadata["tag"].(string); ok && tag != "" {
		return []string{tag}
	}
	if tags, ok := metadata["tags"].([]any); ok {
		result := make([]string, 0, len(tags))
		for _, t := range tags {
			if s, ok := t.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	if tags, ok := metadata["tags"].([]string); ok {
		return tags
	}
	return nil
}

func extractTime(metadata map[string]any, field string, defaultTime time.Time) time.Time {
	if t, ok := metadata[field].(time.Time); ok {
		return t
	}
	if s, ok := metadata[field].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, s); err == nil {
			return parsed
		}
	}
	return defaultTime
}

func parseFrontmatter(content []byte) (map[string]any, string) {
	// Check for frontmatter delimiter
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return make(map[string]any), string(content)
	}

	// Find the end of frontmatter
	rest := content[4:] // Skip first "---\n"
	endIdx := bytes.Index(rest, []byte("\n---\n"))
	if endIdx == -1 {
		// No closing delimiter, treat as no frontmatter
		return make(map[string]any), string(content)
	}

	// Extract frontmatter and body
	frontmatterBytes := rest[:endIdx]
	bodyBytes := rest[endIdx+5:] // Skip "\n---\n"

	// Parse YAML frontmatter
	var metadata map[string]any
	if err := yaml.Unmarshal(frontmatterBytes, &metadata); err != nil {
		// Failed to parse, return empty metadata
		return make(map[string]any), string(content)
	}

	return metadata, string(bodyBytes)
}

func extractLead(body string) string {
	lines := strings.Split(body, "\n")
	var lead strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines at the start
		if lead.Len() == 0 && line == "" {
			continue
		}

		// Skip headings
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Stop at first empty line after content
		if lead.Len() > 0 && line == "" {
			break
		}

		// Add line to lead
		if line != "" {
			if lead.Len() > 0 {
				lead.WriteString(" ")
			}
			lead.WriteString(line)
		}
	}

	result := lead.String()
	if len(result) > 200 {
		return result[:200] + "..."
	}
	return result
}

// Create creates a new notebook.
func (s *NotebookService) Create(name, path string, register bool) (*Notebook, error) {
	if path == "" {
		path, _ = os.Getwd()
	}

	// If the directory exists, use "." as root (use existing notes)
	// If it doesn't exist, create a ".notes" subdirectory for new notes
	var notesDir string
	if _, err := os.Stat(path); err == nil {
		// Directory exists - use it as the root
		notesDir = path
	} else {
		// Directory doesn't exist - create ".notes" subdirectory
		notesDir = filepath.Join(path, ".notes")
	}

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

	// Create notes directory if it doesn't exist
	// (it already exists for existing notebook directories)
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return nil, err
	}

	// Create notebook directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	state, err := buildNotebookIndexState(notesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan notebook state: %w", err)
	}

	// Create Bleve index for this notebook
	idx, err := s.createIndex(notesDir, state.Files)
	if err != nil {
		return nil, fmt.Errorf("failed to create search index: %w", err)
	}

	noteService := NewNoteService(s.configService, idx, notesDir)

	semanticIdx, semErr := s.createSemanticIndex(notesDir)
	if semErr != nil {
		s.log.Warn().Err(semErr).Msg("failed to initialize semantic backend; using noop fallback")
		semanticIdx = NewNoopSemanticIndex()
	}
	noteService.SetSemanticIndex(semanticIdx)

	notebook := &Notebook{
		Config: config,
		Notes:  noteService,
	}
	s.cachedNotebooks[notesDir] = cachedNotebook{state: state, notebook: notebook}

	// Save config
	if err := notebook.SaveConfig(register, s.configService); err != nil {
		return nil, err
	}

	return notebook, nil
}

// Infer discovers notebook from current context (auto-detection only).
// Note: JOT_NOTEBOOK envvar and --notebook flag are handled upstream in requireNotebook().
// Infer() handles auto-detection priority:
// 1. .jot.json in current directory
// 2. Context matching (registered notebooks with path context)
// 3. Ancestor search (walk up tree for .jot.json)
func (s *NotebookService) Infer(cwd string) (*Notebook, error) {
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// Step 1: Check .jot.json in current directory (direct check)
	if s.HasNotebook(cwd) {
		return s.Open(cwd)
	}

	// Step 2: Check registered notebooks for context match
	notebooks, _ := s.List(cwd)
	for _, nb := range notebooks {
		if nb.MatchContext(cwd) != "" {
			return nb, nil
		}
	}

	// Step 3: Search ancestor directories (start from parent, not current)
	current := filepath.Dir(cwd)
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
	if slices.Contains(n.Config.Contexts, contextPath) {
		return nil // Already exists
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
		Workflows: n.Config.Workflows,
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
		if slices.Contains(notebooks, notebookDir) {
			return nil // Already registered
		}
		configService.Store.Notebooks = append(notebooks, notebookDir)
		return configService.Write(configService.Store)
	}

	return nil
}

// GetViews returns all views defined in a notebook's .jot.json
// Returns an empty map if no views are defined
func (s *NotebookService) GetViews(notebookPath string) (map[string]json.RawMessage, error) {
	if notebookPath == "" {
		return map[string]json.RawMessage{}, nil
	}

	configPath := configFilePath(notebookPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]json.RawMessage{}, nil
		}
		return nil, fmt.Errorf("failed to read notebook config: %w", err)
	}

	var configData map[string]interface{}
	if err := json.Unmarshal(data, &configData); err != nil {
		return nil, fmt.Errorf("failed to parse notebook config: %w", err)
	}

	views, ok := configData["views"].(map[string]interface{})
	if !ok {
		return map[string]json.RawMessage{}, nil
	}

	// Convert to json.RawMessage for consistency
	result := make(map[string]json.RawMessage)
	for name, viewData := range views {
		rawData, err := json.Marshal(viewData)
		if err != nil {
			s.log.Warn().Str("view", name).Err(err).Msg("failed to marshal view definition")
			continue
		}
		result[name] = rawData
	}

	return result, nil
}

// ResolveGroupByType resolves a type name to a group configuration.
// It performs case-insensitive matching, first checking the Type field,
// then the Aliases array.
func (s *NotebookService) ResolveGroupByType(nb *Notebook, typeName string) (*NotebookGroup, error) {
	if typeName == "" {
		return nil, fmt.Errorf("type name cannot be empty")
	}

	typeNameLower := strings.ToLower(typeName)

	// First pass: check Type field (exact match, case-insensitive)
	for i := range nb.Config.Groups {
		group := &nb.Config.Groups[i]
		if group.Type != "" && strings.ToLower(group.Type) == typeNameLower {
			return group, nil
		}
	}

	// Second pass: check Aliases array (case-insensitive)
	for i := range nb.Config.Groups {
		group := &nb.Config.Groups[i]
		for _, alias := range group.Aliases {
			if strings.ToLower(alias) == typeNameLower {
				return group, nil
			}
		}
	}

	// Not found - return error with available types
	available := s.ListAvailableTypes(nb)
	if available == "" {
		return nil, fmt.Errorf("unknown type %q: no types defined in this notebook", typeName)
	}
	return nil, fmt.Errorf("unknown type %q: available types are: %s", typeName, available)
}

// GetDefaultGroup returns the default group configured in the notebook.
// Returns an error if no default_group is configured or if the configured
// group doesn't exist. This is used as a fallback when interactive mode is disabled.
func (s *NotebookService) GetDefaultGroup(nb *Notebook) (*NotebookGroup, error) {
	if nb.Config.DefaultGroup == "" {
		return nil, fmt.Errorf(
			"no group specified and interactive mode disabled. " +
				"Use --type flag or set default_group in notebook config")
	}

	defaultNameLower := strings.ToLower(nb.Config.DefaultGroup)
	for i := range nb.Config.Groups {
		group := &nb.Config.Groups[i]
		// Match by name
		if strings.ToLower(group.Name) == defaultNameLower {
			return group, nil
		}
		// Match by type
		if group.Type != "" && strings.ToLower(group.Type) == defaultNameLower {
			return group, nil
		}
	}

	return nil, fmt.Errorf(
		"default_group %q not found in notebook groups. "+
			"Available groups: %s", nb.Config.DefaultGroup, s.listGroupNames(nb))
}

// listGroupNames returns a comma-separated list of group names.
func (s *NotebookService) listGroupNames(nb *Notebook) string {
	var names []string
	for _, g := range nb.Config.Groups {
		names = append(names, g.Name)
	}
	return strings.Join(names, ", ")
}

// ListAvailableTypes returns a comma-separated list of all types and aliases
// defined in the notebook's groups.
func (s *NotebookService) ListAvailableTypes(nb *Notebook) string {
	var types []string
	seen := make(map[string]bool)

	for _, group := range nb.Config.Groups {
		// Add the primary type
		if group.Type != "" && !seen[strings.ToLower(group.Type)] {
			types = append(types, group.Type)
			seen[strings.ToLower(group.Type)] = true
		}

		// Add aliases
		for _, alias := range group.Aliases {
			if !seen[strings.ToLower(alias)] {
				types = append(types, alias)
				seen[strings.ToLower(alias)] = true
			}
		}
	}

	return strings.Join(types, ", ")
}

// GetGroupDirectory returns the directory path for a group based on its globs.
// If the group has globs, it extracts the directory from the first glob pattern.
// Returns empty string if no directory can be determined.
func (s *NotebookService) GetGroupDirectory(nb *Notebook, group *NotebookGroup) string {
	if len(group.Globs) == 0 {
		return ""
	}

	// Use the first glob pattern to determine directory
	glob := group.Globs[0]

	// Extract directory from glob (e.g., "tasks/*.md" -> "tasks", "meetings/**/*.md" -> "meetings")
	dir := filepath.Dir(glob)

	// Handle patterns like "**/*.md" which should return empty
	if dir == "." || dir == "**" {
		return ""
	}

	// Remove any ** from the path
	parts := strings.Split(dir, string(filepath.Separator))
	var cleanParts []string
	for _, part := range parts {
		if part != "**" && part != "*" {
			cleanParts = append(cleanParts, part)
		}
	}

	if len(cleanParts) == 0 {
		return ""
	}

	return filepath.Join(cleanParts...)
}
