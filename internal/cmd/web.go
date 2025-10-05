package cmd

import (
	"github.com/gowsp/cloud189/pkg/webui"
	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "start web server with modern UI, arg: port (default: 8080)",
	Long:  "Start a web server with a modern web interface for managing cloud files. The web interface provides file listing, uploading, downloading, and other file operations through a user-friendly web UI.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		port := "8080"
		if len(args) > 0 {
			port = args[0]
		}
		webui.Serve(port, App())
	},
}