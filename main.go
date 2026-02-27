package main

import (
	"fmt"
	"os"

	"github.com/zenobi-us/jot/cmd"
)

func main() {
	// Set version information from this package's variables
	cmd.Version = Version
	cmd.BuildDate = BuildDate
	cmd.GitCommit = GitCommit
	cmd.GitBranch = GitBranch

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cmd.ExitCode(err))
	}
}

// Version information for Jot
// These variables are updated automatically by release-please
var (
	// Version is the current version of Jot
	// x-release-please-start-version
	Version = "0.0.2"
	// x-release-please-end

	// BuildDate is set during build time
	BuildDate = "unknown"

	// GitCommit is set during build time
	GitCommit = "unknown"

	// GitBranch is set during build time
	GitBranch = "unknown"
)
