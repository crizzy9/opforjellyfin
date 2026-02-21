// cmd/root.go
package cmd

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"os"

	"github.com/spf13/cobra"
)

var debugMode bool
var verboseMode bool

var rootCmd = &cobra.Command{
	Use:   "opfor",
	Short: "Automates download and metadata for One Pace to Jellyfin",
	Long:  "A CLI tool to download One Pace releases and organize them for use with Jellyfin.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debugMode {
			logger.EnableDebugLogging()
		}

		// Enable verbose logging if flag is set OR if VERBOSE env var is "true"
		// This makes it easy to get full logs in Docker without redeploying:
		//   docker run ... -e VERBOSE=true ...
		if verboseMode || os.Getenv("VERBOSE") == "true" {
			logger.EnableVerboseLogging()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📦 Use a subcommand, e.g. 'download', 'progress' or 'list'")
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug logging to debug.log file")
	rootCmd.PersistentFlags().BoolVar(&verboseMode, "verbose", false, "Enable verbose stdout logging (also set VERBOSE=true env var)")
}

func RootCommand() *cobra.Command {
	return rootCmd
}
