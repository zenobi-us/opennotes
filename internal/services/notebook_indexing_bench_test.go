package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkNotebookService_Open_AgentLoopIndexing(b *testing.B) {
	tmpDir := b.TempDir()
	notebookDir := filepath.Join(tmpDir, "bench-notebook")
	notesDir := filepath.Join(notebookDir, ".notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		b.Fatalf("mkdir notes: %v", err)
	}

	config := StoredNotebookConfig{Name: "bench-notebook", Root: ".notes"}
	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		b.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(notebookDir, NotebookConfigFile), configBytes, 0644); err != nil {
		b.Fatalf("write config: %v", err)
	}

	for i := 0; i < 500; i++ {
		note := fmt.Sprintf("# Note %d\n\nagent loop benchmark content %d\n", i, i)
		path := filepath.Join(notesDir, fmt.Sprintf("note-%03d.md", i))
		if err := os.WriteFile(path, []byte(note), 0644); err != nil {
			b.Fatalf("write note %d: %v", i, err)
		}
	}

	configSvc := createTestConfigServiceForBench(b, tmpDir)

	b.Run("cold-open-reindex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			svc := NewNotebookService(configSvc)
			if _, err := svc.Open(notebookDir); err != nil {
				b.Fatalf("open notebook: %v", err)
			}
		}
	})

	b.Run("warm-open-cached", func(b *testing.B) {
		svc := NewNotebookService(configSvc)
		if _, err := svc.Open(notebookDir); err != nil {
			b.Fatalf("seed open: %v", err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := svc.Open(notebookDir); err != nil {
				b.Fatalf("cached open: %v", err)
			}
		}
	})
}

func createTestConfigServiceForBench(b *testing.B, tmpDir string) *ConfigService {
	b.Helper()
	configPath := filepath.Join(tmpDir, "jot", "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		b.Fatalf("mkdir config dir: %v", err)
	}

	config := Config{Notebooks: nil, NotebookPath: ""}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		b.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		b.Fatalf("write config: %v", err)
	}

	svc, err := NewConfigServiceWithPath(configPath)
	if err != nil {
		b.Fatalf("new config service: %v", err)
	}
	return svc
}
