/**
 * Jot - A CLI for managing markdown-based notes
 */
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/services"
)

// Version information - these are set from version.go at build time
var (
	Version   string
	BuildDate string
	GitCommit string
	GitBranch string
)

var (
	// Services initialized in PersistentPreRunE
	cfgService      *services.ConfigService
	notebookService *services.NotebookService
)

var rootCmd = &cobra.Command{
	Use:     "jot",
	Version: "0.0.2", // This will be updated by the version setting in main()
	Short:   "A CLI for managing markdown-based notes with fast search and automation",
	Long: `Jot is a CLI tool for managing your markdown-based notes
organized in notebooks. Notes are stored as markdown files and can be
searched using fast full-text queries.

QUICK START:
  1. Import existing markdown: jot notebook create "My Notes" --path ~/my-notes
  2. List notes: jot notes list
  3. Filter with DSL: jot notes search "path:projects/*.md"
  4. JSON output ready for jq and automation

DOCUMENTATION:
  📚 Search Guide: https://github.com/zenobi-us/jot/blob/main/docs/commands/notes-search.md
  📋 Notebook Management: https://github.com/zenobi-us/jot/blob/main/docs/notebook-discovery.md

Environment Variables:
  JOT_CONFIG    Path to config file (default: ~/.config/jot/config.json)
  DEBUG               Enable debug logging (set to any value)
  LOG_LEVEL           Set log level (debug, info, warn, error)
  LOG_FORMAT          Set log format (compact, console, json, ci) [default: compact]

Examples:
  # Initialize configuration
  jot init

  # Create a notebook with existing markdown
  jot notebook create "My Notes" --path ~/my-notes

  # List all notes
  jot notes list

  # Search with DSL filters and automation
  jot notes search "path:projects/*.md" | jq`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger first
		services.InitLogger()

		// Initialize config service
		var err error
		cfgService, err = services.NewConfigService()
		if err != nil {
			return err
		}

		// Initialize notebook service
		notebookService = services.NewNotebookService(cfgService)

		return nil
	},
}

// Execute runs the root command.
func Execute() error {
	// Update the version on the root command
	rootCmd.Version = Version
	return rootCmd.Execute()
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().String("notebook", "", "Path to notebook")
}
