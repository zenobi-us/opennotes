package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLegacyQueryDeprecationTracking(t *testing.T) {
	if legacyQueryDeprecationSince == "" || !strings.Contains(legacyQueryDeprecationSince, "TODO(") {
		t.Fatalf("deprecation since placeholder must remain explicit until version is set: %q", legacyQueryDeprecationSince)
	}
	if legacyQueryDeprecationRemovalTarget == "" || !strings.Contains(legacyQueryDeprecationRemovalTarget, "TODO(") {
		t.Fatalf("deprecation removal target placeholder must remain explicit until release is scheduled: %q", legacyQueryDeprecationRemovalTarget)
	}
	if legacyQueryDeprecationTrackingEpic == "" {
		t.Fatal("deprecation tracking epic path must be set")
	}

	epicPath := filepath.Join("..", legacyQueryDeprecationTrackingEpic)
	if _, err := os.Stat(epicPath); err != nil {
		t.Fatalf("deprecation tracking epic missing at %s: %v", epicPath, err)
	}
}

func TestLegacyQueryDeprecationWarningMessage_IncludesMigrationGuidance(t *testing.T) {
	msg := legacyQueryDeprecationWarningMessage()

	if !strings.Contains(msg, "jot notes search query") {
		t.Fatalf("warning must mention deprecated command path: %q", msg)
	}
	if !strings.Contains(msg, "jot notes search \"tag:workflow status:active\"") {
		t.Fatalf("warning must include unified DSL migration example: %q", msg)
	}
	if !strings.Contains(msg, legacyQueryDeprecationTrackingEpic) {
		t.Fatalf("warning must include tracking epic path: %q", msg)
	}
	if !strings.Contains(msg, legacyQueryDeprecationSince) || !strings.Contains(msg, legacyQueryDeprecationRemovalTarget) {
		t.Fatalf("warning must include deprecation timeline placeholders: %q", msg)
	}
}
