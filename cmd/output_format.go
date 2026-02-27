package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/zenobi-us/jot/internal/services"
)

func validateOutputFormat(format string, allowed ...string) error {
	for _, candidate := range allowed {
		if format == candidate {
			return nil
		}
	}

	return fmt.Errorf("invalid format %q (supported: %s)", format, strings.Join(allowed, ", "))
}

func renderNotesByFormat(notes []services.Note, format string) error {
	switch format {
	case "json":
		return renderJSON(notes)
	case "list":
		return displayNoteList(notes)
	default:
		return validateOutputFormat(format, "list", "json")
	}
}

func renderSingleNoteByFormat(note services.Note, format string) error {
	switch format {
	case "json":
		return renderJSON(note)
	case "list":
		return displayNoteDetail(note)
	default:
		return validateOutputFormat(format, "list", "json")
	}
}

func renderNotebookInfoByFormat(nb *services.Notebook, format string) error {
	switch format {
	case "json":
		return renderJSON(notebookInfoPayload(nb))
	case "list":
		return displayNotebookInfo(nb)
	default:
		return validateOutputFormat(format, "list", "json")
	}
}

func notebookInfoPayload(nb *services.Notebook) map[string]any {
	return map[string]any{
		"name":           nb.Config.Name,
		"config_path":    nb.Config.Path,
		"root":           nb.Config.Root,
		"config_version": nb.Config.ConfigVersion,
		"contexts":       nb.Config.Contexts,
		"templates":      nb.Config.Templates,
		"groups":         nb.Config.Groups,
	}
}

func renderJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}
