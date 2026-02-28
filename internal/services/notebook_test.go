package services

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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
