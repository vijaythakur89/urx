package main

import (
	"fmt"
        "os"

	"github.com/spf13/cobra"
        "github.com/vijaythakur89/urx/runtime/local"
)

var runCmd = &cobra.Command{
	Use:   "run [file]",
	Short: "Run a URX artifact",
	Run: func(cmd *cobra.Command, args []string) {


 	    file := args[0]

	    err := local.Run(file)
	    if err != nil {
		    fmt.Println("Error:", err)
		    os.Exit(1)
           }
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
