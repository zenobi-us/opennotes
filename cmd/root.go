package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/services"
)

var (
	// Services initialized in PersistentPreRunE
	cfgService      *services.ConfigService
	dbService       *services.DbService
	notebookService *services.NotebookService
)

var rootCmd = &cobra.Command{
	Use:   "opennotes",
	Short: "A CLI for managing markdown-based notes",
	Long: `OpenNotes is a CLI tool for managing your markdown-based notes
organized in notebooks. Notes are stored as markdown files and can be
queried using DuckDB's powerful SQL capabilities.

Environment Variables:
  OPENNOTES_CONFIG    Path to config file (default: ~/.config/opennotes/config.json)
  DEBUG               Enable debug logging (set to any value)
  LOG_LEVEL           Set log level (debug, info, warn, error)

Examples:
  # Initialize configuration
  opennotes init

  # Create a new notebook
  opennotes notebook create --name "My Notes"

  # List all notes in the current notebook
  opennotes notes list

  # Search for notes containing "todo"
  opennotes notes search "todo"`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger first
		services.InitLogger()

		// Initialize config service
		var err error
		cfgService, err = services.NewConfigService()
		if err != nil {
			return err
		}

		// Initialize database service
		dbService = services.NewDbService()

		// Initialize notebook service
		notebookService = services.NewNotebookService(cfgService, dbService)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Cleanup database connection
		if dbService != nil {
			dbService.Close()
		}
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().String("notebook", "", "Path to notebook")
}
