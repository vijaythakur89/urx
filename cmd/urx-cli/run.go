package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vijaythakur89/urx/runtime/local"
)

// 👇 1. Add this
var envVars []string

var runCmd = &cobra.Command{
	Use:   "run [file]",
	Short: "Run a URX artifact",
	Args:  cobra.ExactArgs(1), // good practice

	Run: func(cmd *cobra.Command, args []string) {

		file := args[0]

		// 👇 3. Pass envVars here
		err := local.Run(file, envVars)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {

	// 👇 2. Add flag here
	runCmd.Flags().StringArrayVarP(
		&envVars,
		"env",
		"e",
		[]string{},
		"Environment variables (KEY=VALUE)",
	)

	rootCmd.AddCommand(runCmd)
}
