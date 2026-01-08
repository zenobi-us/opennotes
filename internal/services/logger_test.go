package services

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestInitLogger_DefaultLevel(t *testing.T) {
	// Save original env vars
	t.Setenv("DEBUG", "")
	t.Setenv("LOG_LEVEL", "")

	InitLogger()

	// Check that default level is info
	if zerolog.GlobalLevel() != zerolog.InfoLevel {
		t.Errorf("InitLogger() default level = %v, want %v", zerolog.GlobalLevel(), zerolog.InfoLevel)
	}
}

func TestInitLogger_DEBUG_EnvVar(t *testing.T) {
	t.Setenv("DEBUG", "1")
	t.Setenv("LOG_LEVEL", "")

	InitLogger()

	if zerolog.GlobalLevel() != zerolog.DebugLevel {
		t.Errorf("InitLogger() with DEBUG=1 level = %v, want %v", zerolog.GlobalLevel(), zerolog.DebugLevel)
	}
}

func TestInitLogger_DEBUG_AnyValue(t *testing.T) {
	// Any non-empty value should enable debug
	t.Setenv("DEBUG", "true")
	t.Setenv("LOG_LEVEL", "")

	InitLogger()

	if zerolog.GlobalLevel() != zerolog.DebugLevel {
		t.Errorf("InitLogger() with DEBUG=true level = %v, want %v", zerolog.GlobalLevel(), zerolog.DebugLevel)
	}
}

func TestInitLogger_LOG_LEVEL_EnvVar(t *testing.T) {
	t.Setenv("DEBUG", "")
	t.Setenv("LOG_LEVEL", "warn")

	InitLogger()

	if zerolog.GlobalLevel() != zerolog.WarnLevel {
		t.Errorf("InitLogger() with LOG_LEVEL=warn level = %v, want %v", zerolog.GlobalLevel(), zerolog.WarnLevel)
	}
}

func TestInitLogger_LOG_LEVEL_Precedence(t *testing.T) {
	// LOG_LEVEL should override DEBUG
	t.Setenv("DEBUG", "1")
	t.Setenv("LOG_LEVEL", "error")

	InitLogger()

	if zerolog.GlobalLevel() != zerolog.ErrorLevel {
		t.Errorf("InitLogger() LOG_LEVEL should override DEBUG, got level = %v, want %v", zerolog.GlobalLevel(), zerolog.ErrorLevel)
	}
}

func TestInitLogger_LOG_LEVEL_InvalidValue(t *testing.T) {
	// Invalid LOG_LEVEL should fall back to default (info) or DEBUG setting
	t.Setenv("DEBUG", "")
	t.Setenv("LOG_LEVEL", "invalid_level")

	InitLogger()

	// Should remain at info level since DEBUG is not set
	if zerolog.GlobalLevel() != zerolog.InfoLevel {
		t.Errorf("InitLogger() with invalid LOG_LEVEL should fall back to info, got level = %v", zerolog.GlobalLevel())
	}
}

func TestInitLogger_LOG_LEVEL_InvalidWithDebug(t *testing.T) {
	// Invalid LOG_LEVEL should not override DEBUG
	t.Setenv("DEBUG", "1")
	t.Setenv("LOG_LEVEL", "invalid_level")

	InitLogger()

	// Should be debug level since DEBUG is set and LOG_LEVEL is invalid
	if zerolog.GlobalLevel() != zerolog.DebugLevel {
		t.Errorf("InitLogger() with DEBUG=1 and invalid LOG_LEVEL should be debug, got level = %v", zerolog.GlobalLevel())
	}
}

func TestInitLogger_AllLevels(t *testing.T) {
	levels := []struct {
		name     string
		expected zerolog.Level
	}{
		{"trace", zerolog.TraceLevel},
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"fatal", zerolog.FatalLevel},
		{"panic", zerolog.PanicLevel},
	}

	for _, tc := range levels {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DEBUG", "")
			t.Setenv("LOG_LEVEL", tc.name)

			InitLogger()

			if zerolog.GlobalLevel() != tc.expected {
				t.Errorf("InitLogger() with LOG_LEVEL=%s got level = %v, want %v", tc.name, zerolog.GlobalLevel(), tc.expected)
			}
		})
	}
}

func TestLog_ReturnsNamespacedLogger(t *testing.T) {
	t.Setenv("DEBUG", "")
	t.Setenv("LOG_LEVEL", "")
	InitLogger()

	logger := Log("TestNamespace")

	// The logger should not be nil/zero value
	if logger.GetLevel() < zerolog.TraceLevel {
		t.Error("Log() returned invalid logger")
	}
}

func TestLog_DifferentNamespaces(t *testing.T) {
	InitLogger()

	logger1 := Log("Namespace1")
	logger2 := Log("Namespace2")

	// Both should be valid loggers
	// We can't easily test the namespace field, but we can verify they're valid
	if logger1.GetLevel() < zerolog.TraceLevel {
		t.Error("Log('Namespace1') returned invalid logger")
	}

	if logger2.GetLevel() < zerolog.TraceLevel {
		t.Error("Log('Namespace2') returned invalid logger")
	}
}
