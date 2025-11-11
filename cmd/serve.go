package cmd

import (
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/web"

	"github.com/spf13/cobra"
)

var port int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web interface",
	Long:  "Starts the web server for the OpforJellyfin UI",
	Run: func(cmd *cobra.Command, args []string) {
		if err := web.StartServer(port); err != nil {
			logger.Log(true, "❌ Failed to start server: %v", err)
		}
	},
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8090, "Port to run the web server on")
	rootCmd.AddCommand(serveCmd)
}
