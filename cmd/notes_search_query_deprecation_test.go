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
