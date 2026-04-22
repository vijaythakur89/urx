package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vijaythakur89/urx/artifacts/builder"
)

var buildCmd = &cobra.Command{
	Use:   "build [path]",
	Short: "Build a URX artifact",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		source := "."
		if len(args) == 1 {
			source = args[0]
		}

		output := "app.urx"

		err := builder.Build(source, output)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
