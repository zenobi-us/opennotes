package services

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

// SelectGroupInteractively displays a list of groups and returns the selected one.
// The user can navigate with arrow keys and select with Enter.
func SelectGroupInteractively(groups []NotebookGroup) (*NotebookGroup, error) {
	if len(groups) == 0 {
		return nil, fmt.Errorf("no groups available for selection")
	}

	options := BuildGroupSelectOptions(groups)

	var selectedIdx int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select a note type:").
				Options(options...).
				Value(&selectedIdx),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	return &groups[selectedIdx], nil
}

// BuildGroupSelectOptions creates huh.Option items from a slice of NotebookGroup.
// This is exported separately to allow testing option building logic.
func BuildGroupSelectOptions(groups []NotebookGroup) []huh.Option[int] {
	options := make([]huh.Option[int], len(groups))
	for i, g := range groups {
		label := BuildGroupLabel(g)
		options[i] = huh.NewOption(label, i)
	}
	return options
}

// BuildGroupLabel creates a display label for a NotebookGroup.
// Uses Name as primary label, and appends Type if different from Name.
func BuildGroupLabel(g NotebookGroup) string {
	label := g.Name
	// Add type info if it's different from the name (provides context)
	if g.Type != "" && g.Type != g.Name {
		label = fmt.Sprintf("%s (%s)", g.Name, g.Type)
	}
	return label
}

// InteractiveContext holds the conditions for interactive selection.
type InteractiveContext struct {
	TypeFlag      string          // --type flag value
	ExplicitPath  string          // explicit path argument
	Groups        []NotebookGroup // available groups
	IsTTY         bool            // is stdin a terminal
	NoInteractive bool            // --no-interactive flag
}

// ShouldShowInteractiveSelector determines if interactive group selection is needed.
// It returns true only when all conditions are met:
// - NoInteractive flag is not set
// - No --type flag was provided
// - No explicit path was provided
// - Multiple groups exist (more than 1)
// - Running in a TTY (interactive terminal)
func ShouldShowInteractiveSelector(ctx InteractiveContext) bool {
	if ctx.NoInteractive {
		return false
	}
	if ctx.TypeFlag != "" {
		return false // Type explicitly specified
	}
	if ctx.ExplicitPath != "" {
		return false // Path explicitly specified
	}
	if len(ctx.Groups) <= 1 {
		return false // Only one or no groups - no choice needed
	}
	if !ctx.IsTTY {
		return false // Not a terminal - can't show interactive UI
	}
	return true
}

// IsTTY returns true if stdin is a terminal (interactive mode).
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
