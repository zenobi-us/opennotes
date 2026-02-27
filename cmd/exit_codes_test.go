package cmd

import (
	"errors"
	"testing"
)

func TestExitCode_DefaultsToGeneralForRegularErrors(t *testing.T) {
	err := errors.New("boom")
	if ExitCode(err) != ExitCodeGeneral {
		t.Fatalf("expected general exit code %d, got %d", ExitCodeGeneral, ExitCode(err))
	}
}

func TestExitCode_UsesWrappedCode(t *testing.T) {
	err := withExitCode(ExitCodeConflict, errors.New("conflict"))
	if ExitCode(err) != ExitCodeConflict {
		t.Fatalf("expected conflict exit code %d, got %d", ExitCodeConflict, ExitCode(err))
	}
}
