package cmd

import (
	"strings"
	"testing"
)

func TestNotesSearchQuerySubcommandRemoved(t *testing.T) {
	for _, sub := range notesSearchCmd.Commands() {
		if sub.Name() == "query" {
			t.Fatalf("expected legacy 'query' subcommand to be removed")
		}
	}
}

func TestSearchHelpDoesNotMentionLegacyQuerySubcommand(t *testing.T) {
	if strings.Contains(notesSearchCmd.Long, "jot notes search query") {
		t.Fatalf("notes search help still references removed legacy query command")
	}
	if strings.Contains(notesSearchCmd.Long, "BOOLEAN QUERY SUBCOMMAND") {
		t.Fatalf("notes search help still documents removed boolean query subcommand")
	}
}

func TestRootAndNotesHelpDoNotMentionLegacyQueryCommand(t *testing.T) {
	if strings.Contains(rootCmd.Long, "jot notes search query") {
		t.Fatalf("root help still references removed legacy query command")
	}
	if strings.Contains(notesCmd.Long, "jot notes search query") {
		t.Fatalf("notes help still references removed legacy query command")
	}
}
