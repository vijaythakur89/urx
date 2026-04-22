package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vijaythakur89/urx/runtime/local"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [file]",
	Short: "Deploy a URX app (service mode)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		file := args[0]

		err := local.Deploy(file)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
