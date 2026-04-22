package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vijaythakur89/urx/artifacts/inspector"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [file]",
	Short: "Inspect a URX artifact",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		file := args[0]

		err := inspector.Inspect(file)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
