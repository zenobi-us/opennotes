package services

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions

// createTestNotebook creates a notebook directory with config for testing.
func createTestNotebook(t *testing.T, dir, name string) string {
	t.Helper()

	notebookDir := filepath.Join(dir, name)
	notesDir := filepath.Join(notebookDir, ".notes")

	require.NoError(t, os.MkdirAll(notesDir, 0755))

	config := StoredNotebookConfig{
		Name:     name,
		Root:     ".notes",
		Contexts: []string{notebookDir},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err)

	configPath := filepath.Join(notebookDir, NotebookConfigFile)
	require.NoError(t, os.WriteFile(configPath, data, 0644))

	return notebookDir
}

// createTestConfigService creates a ConfigService with a test config file.
func createTestConfigService(t *testing.T, tmpDir string, notebooks []string) *ConfigService {
	t.Helper()

	configPath := filepath.Join(tmpDir, "jot", "config.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))

	config := Config{
		Notebooks:    notebooks,
		NotebookPath: "",
	}

	data, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, data, 0644))

	svc, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	return svc
}

// HasNotebook tests

func TestNotebookService_HasNotebook_ExistsTrue(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "test-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	assert.True(t, svc.HasNotebook(notebookDir))
}

func TestNotebookService_HasNotebook_NotExistsFalse(t *testing.T) {
	tmpDir := t.TempDir()

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nonExistent := filepath.Join(tmpDir, "non-existent")
	assert.False(t, svc.HasNotebook(nonExistent))
}

func TestNotebookService_HasNotebook_EmptyPath(t *testing.T) {
	tmpDir := t.TempDir()

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	assert.False(t, svc.HasNotebook(""))
}

// LoadConfig tests

func TestNotebookService_LoadConfig_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "test-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	config, err := svc.LoadConfig(notebookDir)
	require.NoError(t, err)

	assert.Equal(t, "test-notebook", config.Name)
	assert.Equal(t, filepath.Join(notebookDir, ".notes"), config.Root)
	assert.Equal(t, []string{notebookDir}, config.Contexts)
}

func TestNotebookService_LoadConfig_DefaultsLegacyConfigVersion(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "legacy-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	config, err := svc.LoadConfig(notebookDir)
	require.NoError(t, err)

	assert.Equal(t, NotebookConfigVersionBootstrap, config.ConfigVersion)
}

func TestNotebookService_LoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "invalid")

	require.NoError(t, os.MkdirAll(notebookDir, 0755))
	configPath := filepath.Join(notebookDir, NotebookConfigFile)
	require.NoError(t, os.WriteFile(configPath, []byte("{ invalid json }"), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	_, err := svc.LoadConfig(notebookDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid notebook config")
}

func TestNotebookService_LoadConfig_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "missing")

	require.NoError(t, os.MkdirAll(notebookDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	_, err := svc.LoadConfig(notebookDir)
	assert.Error(t, err)
}

func TestNotebookService_LoadConfig_CreatesRootIfMissing(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notebook")

	require.NoError(t, os.MkdirAll(notebookDir, 0755))

	// Create config pointing to non-existent root
	config := StoredNotebookConfig{
		Name: "test",
		Root: "notes-missing",
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	configPath := filepath.Join(notebookDir, NotebookConfigFile)
	require.NoError(t, os.WriteFile(configPath, data, 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	loadedConfig, err := svc.LoadConfig(notebookDir)
	require.NoError(t, err)

	// Root directory should have been created
	rootPath := filepath.Join(notebookDir, "notes-missing")
	_, err = os.Stat(rootPath)
	assert.NoError(t, err)
	assert.Equal(t, rootPath, loadedConfig.Root)
}

// Open tests

func TestNotebookService_Open_Success(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "test-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	assert.Equal(t, "test-notebook", notebook.Config.Name)
}

func TestNotebookService_Open_LoadsNoteService(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "test-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	assert.NotNil(t, notebook.Notes)
}

func TestNotebookService_Open_PersistsIndexStateAndReusesWhenUnchanged(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "test-notebook")
	notesDir := filepath.Join(notebookDir, ".notes")
	notePath := filepath.Join(notesDir, "first.md")
	require.NoError(t, os.WriteFile(notePath, []byte("# First\n\nhello indexing"), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	notebook2, err := svc.Open(notebookDir)
	require.NoError(t, err)
	assert.Same(t, notebook, notebook2, "open should reuse cached notebook when files are unchanged")

	results, err := notebook2.Notes.SearchNotes(context.Background(), "hello")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "first.md", results[0].File.Filepath)
}

func TestNotebookService_Open_RebuildsIndexStateWhenNotesChange(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "test-notebook")
	notesDir := filepath.Join(notebookDir, ".notes")
	notePath := filepath.Join(notesDir, "first.md")
	require.NoError(t, os.WriteFile(notePath, []byte("# First\n\nhello indexing"), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	notebookBefore, err := svc.Open(notebookDir)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)
	require.NoError(t, os.WriteFile(notePath, []byte("# First\n\nhello updated token"), 0644))
	now := time.Now().Add(20 * time.Millisecond)
	require.NoError(t, os.Chtimes(notePath, now, now))

	notebookAfterChange, err := svc.Open(notebookDir)
	require.NoError(t, err)
	assert.NotSame(t, notebookBefore, notebookAfterChange, "open should rebuild when notes change")

	results, err := notebookAfterChange.Notes.SearchNotes(context.Background(), "updated")
	require.NoError(t, err)
	require.Len(t, results, 1)
}

// Create tests

func TestNotebookService_Create_CreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "new-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Create("new-notebook", notebookDir, false)
	require.NoError(t, err)

	// Check notebook dir exists
	_, err = os.Stat(notebookDir)
	assert.NoError(t, err)

	// Check notes dir exists
	notesDir := filepath.Join(notebookDir, ".notes")
	_, err = os.Stat(notesDir)
	assert.NoError(t, err)

	assert.Equal(t, "new-notebook", notebook.Config.Name)
}

func TestNotebookService_Create_WritesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "new-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	_, err := svc.Create("new-notebook", notebookDir, false)
	require.NoError(t, err)

	// Check config file exists
	configPath := filepath.Join(notebookDir, NotebookConfigFile)
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Verify config content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var stored StoredNotebookConfig
	require.NoError(t, json.Unmarshal(data, &stored))

	assert.Equal(t, "new-notebook", stored.Name)
	assert.Equal(t, ".notes", stored.Root) // Should be relative
}

func TestNotebookService_Create_RegistersGlobally(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "new-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	_, err := svc.Create("new-notebook", notebookDir, true)
	require.NoError(t, err)

	// Verify notebook was registered
	assert.Contains(t, configSvc.Store.Notebooks, notebookDir)
}

func TestNotebookService_Create_WithoutRegister(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "new-notebook")

	initialNotebooks := []string{"/existing/notebook"}
	configSvc := createTestConfigService(t, tmpDir, initialNotebooks)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	_, err := svc.Create("new-notebook", notebookDir, false)
	require.NoError(t, err)

	// Verify notebook was NOT registered
	assert.NotContains(t, configSvc.Store.Notebooks, notebookDir)
	assert.Equal(t, initialNotebooks, configSvc.Store.Notebooks)
}

// Infer tests

func TestNotebookService_Infer_CurrentDirectoryPriority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a notebook in current directory
	currentNotebook := createTestNotebook(t, tmpDir, "current")

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	// Infer from current directory should find the notebook immediately
	notebook, err := svc.Infer(currentNotebook)
	require.NoError(t, err)
	require.NotNil(t, notebook)

	assert.Equal(t, "current", notebook.Config.Name)
}

func TestNotebookService_Infer_ContextMatchPriority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a notebook with context matching workDir
	workDir := filepath.Join(tmpDir, "work", "project")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	notebookDir := filepath.Join(tmpDir, "notebooks", "work-notebook")
	notesDir := filepath.Join(notebookDir, ".notes")
	require.NoError(t, os.MkdirAll(notesDir, 0755))

	config := StoredNotebookConfig{
		Name:     "work-notebook",
		Root:     ".notes",
		Contexts: []string{filepath.Join(tmpDir, "work")}, // Parent of workDir
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	configPath := filepath.Join(notebookDir, NotebookConfigFile)
	require.NoError(t, os.WriteFile(configPath, data, 0644))

	// Register the notebook
	configSvc := createTestConfigService(t, tmpDir, []string{notebookDir})
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	// Infer should find via context match
	notebook, err := svc.Infer(workDir)
	require.NoError(t, err)
	require.NotNil(t, notebook)

	assert.Equal(t, "work-notebook", notebook.Config.Name)
}

func TestNotebookService_Infer_AncestorSearchPriority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notebook in ancestor
	ancestorNotebook := createTestNotebook(t, tmpDir, "project")

	// Work from a subdirectory
	subDir := filepath.Join(ancestorNotebook, "src", "deep")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	// Infer from subdirectory should find ancestor notebook
	notebook, err := svc.Infer(subDir)
	require.NoError(t, err)
	require.NotNil(t, notebook)

	assert.Equal(t, "project", notebook.Config.Name)
}

func TestNotebookService_Infer_NoneFound(t *testing.T) {
	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Infer(workDir)
	require.NoError(t, err)
	assert.Nil(t, notebook)
}

// List tests

func TestNotebookService_List_FromRegistered(t *testing.T) {
	tmpDir := t.TempDir()

	nb1 := createTestNotebook(t, tmpDir, "notebook1")
	nb2 := createTestNotebook(t, tmpDir, "notebook2")

	configSvc := createTestConfigService(t, tmpDir, []string{nb1, nb2})
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	workDir := filepath.Join(tmpDir, "work")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	notebooks, err := svc.List(workDir)
	require.NoError(t, err)

	assert.Len(t, notebooks, 2)
}

func TestNotebookService_List_FromAncestors(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notebook in ancestor directory
	ancestorNb := createTestNotebook(t, tmpDir, "ancestor-notebook")

	// Work from subdirectory
	subDir := filepath.Join(ancestorNb, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	notebooks, err := svc.List(subDir)
	require.NoError(t, err)

	assert.Len(t, notebooks, 1)
	assert.Equal(t, "ancestor-notebook", notebooks[0].Config.Name)
}

func TestNotebookService_List_Deduplicated(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notebook
	nbDir := createTestNotebook(t, tmpDir, "notebook")

	// Register and also be an ancestor
	subDir := filepath.Join(nbDir, "src")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, []string{nbDir})
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	// List from subdir - should find via registered AND ancestor, but dedup
	notebooks, err := svc.List(subDir)
	require.NoError(t, err)

	assert.Len(t, notebooks, 1)
}

func TestNotebookService_List_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	notebooks, err := svc.List(workDir)
	require.NoError(t, err)

	assert.Empty(t, notebooks)
}

// Notebook method tests

func TestNotebook_MatchContext_Match(t *testing.T) {
	notebook := &Notebook{
		Config: NotebookConfig{
			StoredNotebookConfig: StoredNotebookConfig{
				Contexts: []string{"/home/user/projects", "/home/user/work"},
			},
		},
	}

	result := notebook.MatchContext("/home/user/projects/myapp/src")
	assert.Equal(t, "/home/user/projects", result)
}

func TestNotebook_MatchContext_NoMatch(t *testing.T) {
	notebook := &Notebook{
		Config: NotebookConfig{
			StoredNotebookConfig: StoredNotebookConfig{
				Contexts: []string{"/home/user/projects"},
			},
		},
	}

	result := notebook.MatchContext("/home/user/documents")
	assert.Equal(t, "", result)
}

func TestNotebook_AddContext_NewContext(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	newContext := "/new/context/path"
	err = notebook.AddContext(newContext, configSvc)
	require.NoError(t, err)

	assert.Contains(t, notebook.Config.Contexts, newContext)
}

func TestNotebook_AddContext_DuplicateIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	// Add same context twice
	existingContext := notebook.Config.Contexts[0]
	originalLen := len(notebook.Config.Contexts)

	err = notebook.AddContext(existingContext, configSvc)
	require.NoError(t, err)

	// Should not have been added again
	assert.Equal(t, originalLen, len(notebook.Config.Contexts))
}

func TestNotebook_SaveConfig_LocalOnly(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	// Modify and save without registering
	notebook.Config.Name = "renamed-notebook"
	err = notebook.SaveConfig(false, configSvc)
	require.NoError(t, err)

	// Verify local config was updated
	data, err := os.ReadFile(notebook.Config.Path)
	require.NoError(t, err)

	var stored StoredNotebookConfig
	require.NoError(t, json.Unmarshal(data, &stored))
	assert.Equal(t, "renamed-notebook", stored.Name)

	// Verify not registered globally
	assert.NotContains(t, configSvc.Store.Notebooks, notebookDir)
}

func TestNotebook_SaveConfig_WithRegistration(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	err = notebook.SaveConfig(true, configSvc)
	require.NoError(t, err)

	// Verify was registered globally
	assert.Contains(t, configSvc.Store.Notebooks, notebookDir)
}

func TestNotebook_SaveConfig_AvoidsDuplicateRegistration(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "notebook")

	// Already registered
	configSvc := createTestConfigService(t, tmpDir, []string{notebookDir})
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	// Save with register flag
	err = notebook.SaveConfig(true, configSvc)
	require.NoError(t, err)

	// Should still only have one entry
	count := 0
	for _, p := range configSvc.Store.Notebooks {
		if p == notebookDir {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

func TestNotebookService_LoadConfig_IncludesWorkflows(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "workflow-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "workflow-notebook",
		"root": ".notes",
		"workflows": {
			"project_story": {
				"description": "Project flow",
				"initial_state": "proposed",
				"mode": "enforce",
				"states": {
					"proposed": {
						"schema": {"type": "object"},
						"transitions": ["planned"]
					}
				}
			}
		}
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	cfg, err := svc.LoadConfig(notebookDir)
	require.NoError(t, err)
	require.Contains(t, cfg.Workflows, "project_story")
}

func TestNotebook_SaveConfig_PreservesWorkflows(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "workflow-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "workflow-notebook",
		"root": ".notes",
		"contexts": ["` + notebookDir + `"],
		"workflows": {
			"project_story": {
				"description": "Project flow",
				"initial_state": "proposed",
				"mode": "enforce",
				"field": "status",
				"states": {
					"proposed": {
						"schema": {"type": "object"},
						"transitions": ["planned"]
					}
				}
			}
		}
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	notebook.Config.Name = "renamed"
	err = notebook.SaveConfig(false, configSvc)
	require.NoError(t, err)

	updated, err := os.ReadFile(filepath.Join(notebookDir, NotebookConfigFile))
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(updated, &parsed))
	require.Contains(t, parsed, "workflows")
}

func TestNotebookService_LoadConfig_GroupWorkflowID(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "workflow-group-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "workflow-group-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Stories",
				"globs": ["stories/**/*.md"],
				"metadata": {"type": "story"},
				"workflow_id": "project_story"
			}
		],
		"workflows": {
			"project_story": {
				"description": "Project flow",
				"initial_state": "proposed",
				"mode": "enforce",
				"field": "status",
				"states": {
					"proposed": {"schema": {"type": "object"}, "transitions": ["planned"]}
				}
			}
		}
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	cfg, err := svc.LoadConfig(notebookDir)
	require.NoError(t, err)
	require.Len(t, cfg.Groups, 1)
	assert.Equal(t, "project_story", cfg.Groups[0].WorkflowID)
}

func TestNotebookService_LoadConfig_LegacyGroupWorkflowObject_NormalizesWorkflowID(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "legacy-workflow-group-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "legacy-workflow-group-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Stories",
				"globs": ["stories/**/*.md"],
				"metadata": {"type": "story"},
				"workflow": {
					"id": "project_story",
					"field": "status",
					"on_create": true,
					"on_edit": true
				}
			}
		],
		"workflows": {
			"project_story": {
				"description": "Project flow",
				"initial_state": "proposed",
				"mode": "enforce",
				"field": "status",
				"states": {
					"proposed": {"schema": {"type": "object"}, "transitions": ["planned"]}
				}
			}
		}
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	cfg, err := svc.LoadConfig(notebookDir)
	require.NoError(t, err)
	require.Len(t, cfg.Groups, 1)
	assert.Equal(t, "project_story", cfg.Groups[0].WorkflowID)
}

func TestNotebook_SaveConfig_WritesWorkflowIDOnly(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "save-workflow-group-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "save-workflow-group-notebook",
		"root": ".notes",
		"contexts": ["` + notebookDir + `"],
		"groups": [
			{
				"name": "Stories",
				"globs": ["stories/**/*.md"],
				"metadata": {"type": "story"},
				"workflow_id": "project_story"
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)
	err = notebook.SaveConfig(false, configSvc)
	require.NoError(t, err)

	updated, err := os.ReadFile(filepath.Join(notebookDir, NotebookConfigFile))
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(updated, &parsed))
	groups, ok := parsed["groups"].([]any)
	require.True(t, ok)
	require.Len(t, groups, 1)
	group, ok := groups[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "project_story", group["workflow_id"])
	assert.NotContains(t, group, "workflow")
}

// requireNotebook priority tests
// Note: These test the priority behavior, actual requireNotebook function is in cmd/notes_list.go
// We test the priority here by verifying Infer() behavior and manually simulating requireNotebook logic

func TestNotebookService_Infer_CurrentDirectoryWinsOverAncestor(t *testing.T) {
	tmpDir := t.TempDir()

	// Create current directory notebook
	currentNotebook := createTestNotebook(t, tmpDir, "current")
	currentDir := currentNotebook

	// Create ancestor notebook in tmpDir (parent of current) - this should NOT be found
	_ = createTestNotebook(t, tmpDir, "ancestor")

	configSvc := createTestConfigService(t, tmpDir, nil)
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	// Infer from currentDir should find current (not ancestor)
	notebook, err := svc.Infer(currentDir)
	require.NoError(t, err)
	require.NotNil(t, notebook)

	assert.Equal(t, "current", notebook.Config.Name)
}

func TestNotebookService_Infer_ContextMatchWinsOverAncestor(t *testing.T) {
	tmpDir := t.TempDir()

	// Create work directory
	workDir := filepath.Join(tmpDir, "work", "project")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	// Create context-matching notebook (not in ancestor chain)
	contextNotebook := filepath.Join(tmpDir, "notebooks", "context-nb")
	contextNotesDir := filepath.Join(contextNotebook, ".notes")
	require.NoError(t, os.MkdirAll(contextNotesDir, 0755))

	contextConfig := StoredNotebookConfig{
		Name:     "context-notebook",
		Root:     ".notes",
		Contexts: []string{filepath.Join(tmpDir, "work")}, // Matches workDir parent
	}
	contextData, _ := json.MarshalIndent(contextConfig, "", "  ")
	contextConfigPath := filepath.Join(contextNotebook, NotebookConfigFile)
	require.NoError(t, os.WriteFile(contextConfigPath, contextData, 0644))

	// Create ancestor notebook (in tmpDir) - this should NOT be found
	_ = createTestNotebook(t, tmpDir, "ancestor")

	// Register context notebook
	configSvc := createTestConfigService(t, tmpDir, []string{contextNotebook})
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	// Infer from workDir should find context notebook (not ancestor)
	notebook, err := svc.Infer(workDir)
	require.NoError(t, err)
	require.NotNil(t, notebook)

	assert.Equal(t, "context-notebook", notebook.Config.Name)
}

// TestNotebookService_Infer_CompleteResolutionOrder verifies the complete priority order:
// 1. Current directory (.jot.json)
// 2. Context match (registered notebooks)
// 3. Ancestor search
func TestNotebookService_Infer_CompleteResolutionOrder(t *testing.T) {
	tmpDir := t.TempDir()

	// Create work directory structure
	workDir := filepath.Join(tmpDir, "projects", "myproject", "src")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	// Create ancestor notebook (should have lowest priority)
	ancestorNotebook := createTestNotebook(t, tmpDir, "ancestor-notebook")

	// Create context-matching notebook (should have medium priority)
	contextNotebook := filepath.Join(tmpDir, "context-nb")
	contextNotesDir := filepath.Join(contextNotebook, ".notes")
	require.NoError(t, os.MkdirAll(contextNotesDir, 0755))
	contextConfig := StoredNotebookConfig{
		Name:     "context-notebook",
		Root:     ".notes",
		Contexts: []string{filepath.Join(tmpDir, "projects")}, // Parent context
	}
	contextData, _ := json.MarshalIndent(contextConfig, "", "  ")
	contextConfigPath := filepath.Join(contextNotebook, NotebookConfigFile)
	require.NoError(t, os.WriteFile(contextConfigPath, contextData, 0644))

	// Create current directory notebook (should have highest priority)
	currentNotebook := filepath.Join(workDir, ".jot.json")
	currentConfig := StoredNotebookConfig{
		Name:     "current-directory-notebook",
		Root:     ".notes",
		Contexts: []string{workDir},
	}
	currentDir := filepath.Join(workDir, ".notes")
	require.NoError(t, os.MkdirAll(currentDir, 0755))
	currentData, _ := json.MarshalIndent(currentConfig, "", "  ")
	require.NoError(t, os.WriteFile(currentNotebook, currentData, 0644))

	// Register both context and ancestor notebooks
	configSvc := createTestConfigService(t, tmpDir, []string{contextNotebook, ancestorNotebook})
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	// Should find current directory notebook (highest priority)
	notebook, err := svc.Infer(workDir)
	require.NoError(t, err)
	require.NotNil(t, notebook)
	assert.Equal(t, "current-directory-notebook", notebook.Config.Name)
}

// TestNotebookService_Infer_ContextBeforeAncestorWithoutCurrentDir verifies context priority without current dir
func TestNotebookService_Infer_ContextBeforeAncestorWithoutCurrentDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create work directory (no notebook here)
	workDir := filepath.Join(tmpDir, "work", "project")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	// Create ancestor notebook (should be lower priority)
	ancestorNotebook := createTestNotebook(t, tmpDir, "ancestor-notebook")

	// Create context-matching notebook (should win over ancestor)
	contextNotebook := filepath.Join(tmpDir, "context-nb")
	contextNotesDir := filepath.Join(contextNotebook, ".notes")
	require.NoError(t, os.MkdirAll(contextNotesDir, 0755))
	contextConfig := StoredNotebookConfig{
		Name:     "context-notebook",
		Root:     ".notes",
		Contexts: []string{filepath.Join(tmpDir, "work")},
	}
	contextData, _ := json.MarshalIndent(contextConfig, "", "  ")
	contextConfigPath := filepath.Join(contextNotebook, NotebookConfigFile)
	require.NoError(t, os.WriteFile(contextConfigPath, contextData, 0644))

	// Register both
	configSvc := createTestConfigService(t, tmpDir, []string{contextNotebook, ancestorNotebook})
	t.Cleanup(func() {
	})
	svc := NewNotebookService(configSvc)

	// Should find context notebook (not ancestor)
	notebook, err := svc.Infer(workDir)
	require.NoError(t, err)
	require.NotNil(t, notebook)
	assert.Equal(t, "context-notebook", notebook.Config.Name)
}

// ResolveGroupByType tests

func TestNotebookService_ResolveGroupByType_ExactTypeMatch(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "type-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "type-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["tasks/*.md"],
				"type": "task",
				"metadata": {"type": "task"}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	group, err := svc.ResolveGroupByType(nb, "task")
	require.NoError(t, err)
	assert.Equal(t, "Tasks", group.Name)
	assert.Equal(t, "task", group.Type)
}

func TestNotebookService_ResolveGroupByType_CaseInsensitiveMatch(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "case-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "case-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["tasks/*.md"],
				"type": "Task",
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	// Test different case variations
	for _, typeName := range []string{"task", "TASK", "Task", "tAsK"} {
		group, err := svc.ResolveGroupByType(nb, typeName)
		require.NoError(t, err, "failed for type name: %s", typeName)
		assert.Equal(t, "Tasks", group.Name)
	}
}

func TestNotebookService_ResolveGroupByType_AliasMatch(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "alias-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "alias-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["tasks/*.md"],
				"type": "task",
				"aliases": ["todo", "item", "work"],
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	// Test alias matching
	for _, alias := range []string{"todo", "item", "work", "TODO", "WORK"} {
		group, err := svc.ResolveGroupByType(nb, alias)
		require.NoError(t, err, "failed for alias: %s", alias)
		assert.Equal(t, "Tasks", group.Name)
	}

	// Primary type should also work
	group, err := svc.ResolveGroupByType(nb, "task")
	require.NoError(t, err)
	assert.Equal(t, "Tasks", group.Name)
}

func TestNotebookService_ResolveGroupByType_NotFoundError(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notfound-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "notfound-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["tasks/*.md"],
				"type": "task",
				"aliases": ["todo"],
				"metadata": {}
			},
			{
				"name": "Meetings",
				"globs": ["meetings/*.md"],
				"type": "meeting",
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	_, err = svc.ResolveGroupByType(nb, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown type \"nonexistent\"")
	assert.Contains(t, err.Error(), "available types are:")
	assert.Contains(t, err.Error(), "task")
	assert.Contains(t, err.Error(), "meeting")
	assert.Contains(t, err.Error(), "todo")
}

func TestNotebookService_ResolveGroupByType_EmptyTypeName(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "empty-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "empty-notebook",
		"root": ".notes",
		"groups": []
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	_, err = svc.ResolveGroupByType(nb, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type name cannot be empty")
}

func TestNotebookService_ResolveGroupByType_NoTypesDefinedError(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notypes-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "notypes-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Notes",
				"globs": ["*.md"],
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	_, err = svc.ResolveGroupByType(nb, "task")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no types defined in this notebook")
}

func TestNotebookService_ResolveGroupByType_TypePrecedenceOverAlias(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "precedence-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	// Create a scenario where "item" is a primary type for one group
	// and an alias for another - primary type should win
	configJSON := `{
		"name": "precedence-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["tasks/*.md"],
				"type": "task",
				"aliases": ["item"],
				"metadata": {}
			},
			{
				"name": "Items",
				"globs": ["items/*.md"],
				"type": "item",
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	// "item" should resolve to Items group (exact type match, not alias)
	group, err := svc.ResolveGroupByType(nb, "item")
	require.NoError(t, err)
	assert.Equal(t, "Items", group.Name)

	// "task" should resolve to Tasks group
	group, err = svc.ResolveGroupByType(nb, "task")
	require.NoError(t, err)
	assert.Equal(t, "Tasks", group.Name)
}

// ListAvailableTypes tests

func TestNotebookService_ListAvailableTypes(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "list-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "list-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["tasks/*.md"],
				"type": "task",
				"aliases": ["todo", "item"],
				"metadata": {}
			},
			{
				"name": "Meetings",
				"globs": ["meetings/*.md"],
				"type": "meeting",
				"aliases": ["mtg"],
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	types := svc.ListAvailableTypes(nb)
	assert.Contains(t, types, "task")
	assert.Contains(t, types, "todo")
	assert.Contains(t, types, "item")
	assert.Contains(t, types, "meeting")
	assert.Contains(t, types, "mtg")
}

func TestNotebookService_ListAvailableTypes_NoDuplicates(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "dedup-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	// Both groups use "shared" as an alias (shouldn't duplicate)
	configJSON := `{
		"name": "dedup-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["tasks/*.md"],
				"type": "task",
				"aliases": ["shared", "todo"],
				"metadata": {}
			},
			{
				"name": "Notes",
				"globs": ["notes/*.md"],
				"type": "note",
				"aliases": ["shared"],
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	types := svc.ListAvailableTypes(nb)
	// Count occurrences of "shared" - should only appear once
	count := strings.Count(types, "shared")
	assert.Equal(t, 1, count, "shared should only appear once")
}

func TestNotebookService_ListAvailableTypes_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "empty-types-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "empty-types-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Notes",
				"globs": ["*.md"],
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	types := svc.ListAvailableTypes(nb)
	assert.Equal(t, "", types)
}

// GetGroupDirectory tests

func TestNotebookService_GetGroupDirectory_SimpleGlob(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "dir-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "dir-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["tasks/*.md"],
				"type": "task",
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	group := &nb.Config.Groups[0]
	dir := svc.GetGroupDirectory(nb, group)
	assert.Equal(t, "tasks", dir)
}

func TestNotebookService_GetGroupDirectory_NestedGlob(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "nested-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "nested-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["projects/tasks/*.md"],
				"type": "task",
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	group := &nb.Config.Groups[0]
	dir := svc.GetGroupDirectory(nb, group)
	assert.Equal(t, filepath.Join("projects", "tasks"), dir)
}

func TestNotebookService_GetGroupDirectory_RecursiveGlob(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "recursive-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "recursive-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["tasks/**/*.md"],
				"type": "task",
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	group := &nb.Config.Groups[0]
	dir := svc.GetGroupDirectory(nb, group)
	assert.Equal(t, "tasks", dir)
}

func TestNotebookService_GetGroupDirectory_RootGlob(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "root-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "root-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "All Notes",
				"globs": ["*.md"],
				"type": "note",
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	group := &nb.Config.Groups[0]
	dir := svc.GetGroupDirectory(nb, group)
	assert.Equal(t, "", dir)
}

func TestNotebookService_GetGroupDirectory_NoGlobs(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "noglob-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "noglob-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "Empty",
				"globs": [],
				"type": "empty",
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	group := &nb.Config.Groups[0]
	dir := svc.GetGroupDirectory(nb, group)
	assert.Equal(t, "", dir)
}

func TestNotebookService_GetGroupDirectory_DoubleStarOnly(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "doublestar-notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "doublestar-notebook",
		"root": ".notes",
		"groups": [
			{
				"name": "All",
				"globs": ["**/*.md"],
				"type": "all",
				"metadata": {}
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	group := &nb.Config.Groups[0]
	dir := svc.GetGroupDirectory(nb, group)
	assert.Equal(t, "", dir)
}

// NotebookGroup FilenameFormat tests

func TestNotebookGroup_GetFilenameFormat_EmptyReturnsDefault(t *testing.T) {
	group := NotebookGroup{
		Name:  "Test",
		Globs: []string{"**/*.md"},
	}
	assert.Equal(t, DefaultFilenameFormat, group.GetFilenameFormat())
}

func TestNotebookGroup_GetFilenameFormat_CustomFormatPreserved(t *testing.T) {
	customFormat := "{{ now | date \"2006-01-02\" }}-{{ .title | slug }}.md"
	group := NotebookGroup{
		Name:           "Test",
		Globs:          []string{"**/*.md"},
		FilenameFormat: customFormat,
	}
	assert.Equal(t, customFormat, group.GetFilenameFormat())
}

func TestNotebookGroup_FilenameFormat_ParsesFromJSON(t *testing.T) {
	jsonData := `{
		"name": "Test Group",
		"globs": ["**/*.md"],
		"filename_format": "{{ .title | slug }}-custom.md"
	}`

	var group NotebookGroup
	err := json.Unmarshal([]byte(jsonData), &group)
	require.NoError(t, err)
	assert.Equal(t, "{{ .title | slug }}-custom.md", group.FilenameFormat)
}

func TestNotebookGroup_FilenameFormat_EmptyWhenOmittedFromJSON(t *testing.T) {
	jsonData := `{
		"name": "Test Group",
		"globs": ["**/*.md"]
	}`

	var group NotebookGroup
	err := json.Unmarshal([]byte(jsonData), &group)
	require.NoError(t, err)
	assert.Empty(t, group.FilenameFormat)
	assert.Equal(t, DefaultFilenameFormat, group.GetFilenameFormat())
}

func TestNotebookGroup_ValidateFilenameFormat_EmptyIsValid(t *testing.T) {
	group := NotebookGroup{
		Name:  "Test",
		Globs: []string{"**/*.md"},
	}
	assert.NoError(t, group.ValidateFilenameFormat())
}

func TestNotebookGroup_ValidateFilenameFormat_ValidFormat(t *testing.T) {
	group := NotebookGroup{
		Name:           "Test",
		Globs:          []string{"**/*.md"},
		FilenameFormat: "{{ .title | slug }}.md",
	}
	assert.NoError(t, group.ValidateFilenameFormat())
}

func TestNotebookGroup_ValidateFilenameFormat_MustEndWithMd(t *testing.T) {
	group := NotebookGroup{
		Name:           "Test",
		Globs:          []string{"**/*.md"},
		FilenameFormat: "{{ .title | slug }}.txt",
	}
	err := group.ValidateFilenameFormat()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must end with .md")
}

func TestNotebookGroup_ValidateFilenameFormat_NoForwardSlash(t *testing.T) {
	group := NotebookGroup{
		Name:           "Test",
		Globs:          []string{"**/*.md"},
		FilenameFormat: "subdir/{{ .title | slug }}.md",
	}
	err := group.ValidateFilenameFormat()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not contain path separators")
}

func TestNotebookGroup_ValidateFilenameFormat_NoBackslash(t *testing.T) {
	group := NotebookGroup{
		Name:           "Test",
		Globs:          []string{"**/*.md"},
		FilenameFormat: "subdir\\{{ .title | slug }}.md",
	}
	err := group.ValidateFilenameFormat()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not contain path separators")
}

func TestNotebookService_LoadConfig_IncludesFilenameFormat(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notebook")

	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "Test",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["tasks/**/*.md"],
				"filename_format": "{{ now | date \"2006-01-02\" }}-{{ .title | slug }}.md"
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	cfg, err := svc.LoadConfig(notebookDir)
	require.NoError(t, err)
	require.Len(t, cfg.Groups, 1)
	assert.Equal(t, "{{ now | date \"2006-01-02\" }}-{{ .title | slug }}.md", cfg.Groups[0].FilenameFormat)
}

// NotebookGroup GetTemplate tests

func TestNotebookGroup_GetTemplate_EmptyReturnsDefault(t *testing.T) {
	group := NotebookGroup{
		Name:  "Test",
		Globs: []string{"**/*.md"},
	}
	assert.Equal(t, DefaultContentTemplate, group.GetTemplate())
}

func TestNotebookGroup_GetTemplate_CustomTemplatePreserved(t *testing.T) {
	customTemplate := `---
title: {{ .title }}
type: custom
---

Custom content here.
`
	group := NotebookGroup{
		Name:     "Test",
		Globs:    []string{"**/*.md"},
		Template: customTemplate,
	}
	assert.Equal(t, customTemplate, group.GetTemplate())
}

func TestNotebookGroup_Template_ParsesFromJSON(t *testing.T) {
	configJSON := `{
		"name": "Tasks",
		"globs": ["tasks/**/*.md"],
		"template": "---\ntitle: {{ .title }}\n---\n"
	}`

	var group NotebookGroup
	err := json.Unmarshal([]byte(configJSON), &group)
	require.NoError(t, err)
	assert.Equal(t, "---\ntitle: {{ .title }}\n---\n", group.Template)
}

func TestNotebookGroup_Template_EmptyWhenOmittedFromJSON(t *testing.T) {
	configJSON := `{
		"name": "Tasks",
		"globs": ["tasks/**/*.md"]
	}`

	var group NotebookGroup
	err := json.Unmarshal([]byte(configJSON), &group)
	require.NoError(t, err)
	assert.Equal(t, DefaultContentTemplate, group.GetTemplate())
}

func TestNotebookService_LoadConfig_IncludesTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notebook")

	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "Test",
		"root": ".notes",
		"groups": [
			{
				"name": "Tasks",
				"globs": ["tasks/**/*.md"],
				"template": "---\ntitle: {{ .title }}\ntype: task\n---\n# {{ .title }}\n"
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	cfg, err := svc.LoadConfig(notebookDir)
	require.NoError(t, err)
	require.Len(t, cfg.Groups, 1)
	assert.Equal(t, "---\ntitle: {{ .title }}\ntype: task\n---\n# {{ .title }}\n", cfg.Groups[0].Template)
}

// GetDefaultGroup tests

func TestNotebookService_GetDefaultGroup_ReturnsConfiguredDefault(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "Test",
		"root": ".notes",
		"default_group": "Tasks",
		"groups": [
			{"name": "Tasks", "globs": ["tasks/**/*.md"], "type": "task"},
			{"name": "Notes", "globs": ["notes/**/*.md"], "type": "note"}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	group, err := svc.GetDefaultGroup(nb)
	require.NoError(t, err)
	assert.Equal(t, "Tasks", group.Name)
	assert.Equal(t, "task", group.Type)
}

func TestNotebookService_GetDefaultGroup_MatchesByType(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "Test",
		"root": ".notes",
		"default_group": "task",
		"groups": [
			{"name": "Daily Tasks", "globs": ["tasks/**/*.md"], "type": "task"},
			{"name": "Notes", "globs": ["notes/**/*.md"], "type": "note"}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	group, err := svc.GetDefaultGroup(nb)
	require.NoError(t, err)
	assert.Equal(t, "Daily Tasks", group.Name)
	assert.Equal(t, "task", group.Type)
}

func TestNotebookService_GetDefaultGroup_ErrorsWithoutDefault(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "Test",
		"root": ".notes",
		"groups": [
			{"name": "Tasks", "globs": ["tasks/**/*.md"], "type": "task"},
			{"name": "Notes", "globs": ["notes/**/*.md"], "type": "note"}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	_, err = svc.GetDefaultGroup(nb)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no group specified and interactive mode disabled")
	assert.Contains(t, err.Error(), "Use --type flag or set default_group")
}

func TestNotebookService_GetDefaultGroup_ErrorsWithInvalidDefault(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "Test",
		"root": ".notes",
		"default_group": "NonExistent",
		"groups": [
			{"name": "Tasks", "globs": ["tasks/**/*.md"], "type": "task"},
			{"name": "Notes", "globs": ["notes/**/*.md"], "type": "note"}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	_, err = svc.GetDefaultGroup(nb)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "default_group \"NonExistent\" not found")
	assert.Contains(t, err.Error(), "Available groups: Tasks, Notes")
}

func TestNotebookService_GetDefaultGroup_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notebook")
	require.NoError(t, os.MkdirAll(filepath.Join(notebookDir, ".notes"), 0755))

	configJSON := `{
		"name": "Test",
		"root": ".notes",
		"default_group": "TASKS",
		"groups": [
			{"name": "Tasks", "globs": ["tasks/**/*.md"], "type": "task"},
			{"name": "Notes", "globs": ["notes/**/*.md"], "type": "note"}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), []byte(configJSON), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	nb, err := svc.Open(notebookDir)
	require.NoError(t, err)

	group, err := svc.GetDefaultGroup(nb)
	require.NoError(t, err)
	assert.Equal(t, "Tasks", group.Name)
}

// Integration test: GetFilenameFormat with GenerateFilename
// This tests the full flow used by notes add command when creating notes with groups

func TestNotebookGroup_FilenameFormat_IntegrationWithGenerateFilename(t *testing.T) {
	tests := []struct {
		name           string
		filenameFormat string
		title          string
		expectedName   string
	}{
		{
			name:           "default format produces slugified filename",
			filenameFormat: "", // Will use DefaultFilenameFormat
			title:          "My Task Title",
			expectedName:   "my-task-title.md",
		},
		{
			name:           "custom format with prefix",
			filenameFormat: "task-{{ .title | slug }}.md",
			title:          "Fix Login Bug",
			expectedName:   "task-fix-login-bug.md",
		},
		{
			name:           "custom format with uppercase",
			filenameFormat: "{{ .title | upper | slug }}.md",
			title:          "Hello World",
			expectedName:   "hello-world.md",
		},
		{
			name:           "custom format with max length",
			filenameFormat: "{{ slugmax .title 10 }}.md",
			title:          "This Is A Very Long Title That Should Be Truncated",
			expectedName:   "this-is-a.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := NotebookGroup{
				Name:           "Tasks",
				Globs:          []string{"tasks/**/*.md"},
				FilenameFormat: tt.filenameFormat,
			}

			// Get the format (falls back to default if empty)
			format := group.GetFilenameFormat()

			// Generate the filename (simulates what notes add does)
			generatedFilename, err := GenerateFilename(format, tt.title)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedName, generatedFilename,
				"filename format %q with title %q", format, tt.title)
		})
	}
}

func TestNotebookGroup_FilenameFormat_ErrorHandling(t *testing.T) {
	// Test that invalid templates are handled gracefully
	group := NotebookGroup{
		Name:           "Tasks",
		Globs:          []string{"tasks/**/*.md"},
		FilenameFormat: "{{ .title | unknownfunction }}.md",
	}

	format := group.GetFilenameFormat()
	_, err := GenerateFilename(format, "Test Title")
	assert.Error(t, err, "should error on invalid template function")
}
