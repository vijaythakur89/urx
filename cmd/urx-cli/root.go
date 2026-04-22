package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "urx",
	Short: "Universal Runtime CLI",
	Long: `URX - Universal Runtime

Run applications anywhere with a single command.

URX simplifies execution by abstracting containers
and runtime complexity from developers.`,
}

// Execute runs the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("[URX] Error:", err)
		os.Exit(1)
	}
}

func init() {
	// Customize help command (optional but clean)
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:   "help",
		Short: "Help about any command",
	})
}
