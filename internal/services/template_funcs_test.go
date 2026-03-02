package services

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestJotNamespace_Slug(t *testing.T) {
	te := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "Basic slug conversion",
			template: `{{ jot.Slug "Hello World" }}`,
			expected: "hello-world",
		},
		{
			name:     "Slug with special characters",
			template: `{{ jot.Slug "Hello, World! How are you?" }}`,
			expected: "hello-world-how-are-you",
		},
		{
			name:     "Slug with variable",
			template: `{{ .title | jot.Slug }}`,
			expected: "my-test-title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{"title": "My Test Title"}
			result, err := te.Render(tt.template, data)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestJotNamespace_NanoID(t *testing.T) {
	te := NewTemplateEngine()

	tests := []struct {
		name   string
		length int
	}{
		{"8 character NanoID", 8},
		{"12 character NanoID", 12},
		{"21 character NanoID", 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := `{{ jot.NanoID ` + string(rune('0'+tt.length/10)) + string(rune('0'+tt.length%10)) + ` }}`
			// Build template correctly
			if tt.length < 10 {
				template = `{{ jot.NanoID ` + string(rune('0'+tt.length)) + ` }}`
			}

			result, err := te.Render(template, nil)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			if len(result) != tt.length {
				t.Errorf("expected length %d, got %d (%q)", tt.length, len(result), result)
			}

			// Verify URL-safe characters only (A-Za-z0-9_-)
			urlSafePattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
			if !urlSafePattern.MatchString(result) {
				t.Errorf("NanoID contains non-URL-safe characters: %q", result)
			}
		})
	}

	// Test uniqueness (probabilistic)
	t.Run("NanoIDs are unique", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			result, err := te.Render(`{{ jot.NanoID 12 }}`, nil)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}
			if seen[result] {
				t.Errorf("Duplicate NanoID generated: %q", result)
			}
			seen[result] = true
		}
	})
}

func TestJotNamespace_Timestamp(t *testing.T) {
	te := NewTemplateEngine()

	before := time.Now().Unix()
	result, err := te.Render(`{{ jot.Timestamp }}`, nil)
	after := time.Now().Unix()

	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Parse the result as int64
	var ts int64
	_, scanErr := regexp.MatchString(`^\d+$`, result)
	if scanErr != nil {
		t.Fatalf("Timestamp is not numeric: %q", result)
	}

	// Convert string to int for comparison
	for _, c := range result {
		ts = ts*10 + int64(c-'0')
	}

	if ts < before || ts > after {
		t.Errorf("Timestamp %d not in expected range [%d, %d]", ts, before, after)
	}
}

func TestJotNamespace_DatePath(t *testing.T) {
	te := NewTemplateEngine()

	result, err := te.Render(`{{ jot.DatePath }}`, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify format matches YYYY/MM/DD
	datePathPattern := regexp.MustCompile(`^\d{4}/\d{2}/\d{2}$`)
	if !datePathPattern.MatchString(result) {
		t.Errorf("DatePath does not match YYYY/MM/DD pattern: %q", result)
	}

	// Verify it matches today's date
	expected := time.Now().Format("2006/01/02")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestJotNamespace_UUID(t *testing.T) {
	te := NewTemplateEngine()

	result, err := te.Render(`{{ jot.UUID }}`, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidPattern.MatchString(result) {
		t.Errorf("UUID does not match v4 format: %q", result)
	}

	// Test uniqueness
	t.Run("UUIDs are unique", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			result, err := te.Render(`{{ jot.UUID }}`, nil)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}
			if seen[result] {
				t.Errorf("Duplicate UUID generated: %q", result)
			}
			seen[result] = true
		}
	})
}

func TestJotNamespace_Now(t *testing.T) {
	te := NewTemplateEngine()

	tests := []struct {
		name     string
		format   string
		validate func(string) bool
	}{
		{
			name:   "Date format",
			format: "2006-01-02",
			validate: func(s string) bool {
				return s == time.Now().Format("2006-01-02")
			},
		},
		{
			name:   "Year only",
			format: "2006",
			validate: func(s string) bool {
				return s == time.Now().Format("2006")
			},
		},
		{
			name:   "Full datetime",
			format: "2006-01-02T15:04",
			validate: func(s string) bool {
				// Allow 1 minute tolerance
				pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$`)
				return pattern.MatchString(s)
			},
		},
		{
			name:   "Custom format with text",
			format: "Jan 2, 2006",
			validate: func(s string) bool {
				pattern := regexp.MustCompile(`^[A-Za-z]+ \d+, \d{4}$`)
				return pattern.MatchString(s)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := `{{ jot.Now "` + tt.format + `" }}`
			result, err := te.Render(template, nil)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}
			if !tt.validate(result) {
				t.Errorf("Now(%q) returned invalid result: %q", tt.format, result)
			}
		})
	}
}

func TestJotNamespace_Integration(t *testing.T) {
	te := NewTemplateEngine()

	t.Run("Full filename format with date and slug", func(t *testing.T) {
		template := `{{ jot.Now "2006-01-02" }}-{{ .title | slug }}.md`
		data := map[string]interface{}{"title": "My Test Note"}

		result, err := te.Render(template, data)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}

		// Verify format: YYYY-MM-DD-slug.md
		today := time.Now().Format("2006-01-02")
		expected := today + "-my-test-note.md"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("DatePath with nanoid filename", func(t *testing.T) {
		template := `{{ jot.DatePath }}/{{ jot.NanoID 8 }}-{{ .title | slug }}.md`
		data := map[string]interface{}{"title": "Quick Note"}

		result, err := te.Render(template, data)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}

		// Verify format: YYYY/MM/DD/xxxxxxxx-slug.md
		pattern := regexp.MustCompile(`^\d{4}/\d{2}/\d{2}/[A-Za-z0-9_-]{8}-quick-note\.md$`)
		if !pattern.MatchString(result) {
			t.Errorf("Result does not match expected pattern: %q", result)
		}
	})

	t.Run("Timestamp-based filename", func(t *testing.T) {
		template := `{{ jot.Timestamp }}-{{ .title | slug }}.md`
		data := map[string]interface{}{"title": "Timestamped Note"}

		result, err := te.Render(template, data)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}

		// Verify format: timestamp-slug.md
		pattern := regexp.MustCompile(`^\d+-timestamped-note\.md$`)
		if !pattern.MatchString(result) {
			t.Errorf("Result does not match expected pattern: %q", result)
		}
	})

	t.Run("UUID filename", func(t *testing.T) {
		template := `{{ jot.UUID }}.md`

		result, err := te.Render(template, nil)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}

		// Verify format: uuid.md
		if !strings.HasSuffix(result, ".md") {
			t.Errorf("Result does not end with .md: %q", result)
		}
		if len(result) != 36+3 { // UUID + ".md"
			t.Errorf("Unexpected result length: %q", result)
		}
	})

	t.Run("Combined jot.Slug usage", func(t *testing.T) {
		template := `{{ jot.Slug .title }}.md`
		data := map[string]interface{}{"title": "Test With CAPS and Spaces"}

		result, err := te.Render(template, data)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}

		expected := "test-with-caps-and-spaces.md"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

func TestJotFuncs(t *testing.T) {
	// Verify JotFuncs returns a JotNamespace
	jot := JotFuncs()

	// Test direct method calls
	t.Run("Direct Slug call", func(t *testing.T) {
		result := jot.Slug("Hello World")
		if result != "hello-world" {
			t.Errorf("expected 'hello-world', got %q", result)
		}
	})

	t.Run("Direct NanoID call", func(t *testing.T) {
		result := jot.NanoID(10)
		if len(result) != 10 {
			t.Errorf("expected length 10, got %d", len(result))
		}
	})

	t.Run("Direct Timestamp call", func(t *testing.T) {
		before := time.Now().Unix()
		result := jot.Timestamp()
		after := time.Now().Unix()
		if result < before || result > after {
			t.Errorf("Timestamp %d not in expected range", result)
		}
	})

	t.Run("Direct DatePath call", func(t *testing.T) {
		result := jot.DatePath()
		expected := time.Now().Format("2006/01/02")
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("Direct UUID call", func(t *testing.T) {
		result := jot.UUID()
		if len(result) != 36 {
			t.Errorf("UUID length should be 36, got %d", len(result))
		}
	})

	t.Run("Direct Now call", func(t *testing.T) {
		result := jot.Now("2006-01-02")
		expected := time.Now().Format("2006-01-02")
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}
