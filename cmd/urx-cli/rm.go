package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm [id]",
	Short: "Remove a URX container and its metadata",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		id := args[0]

		// remove container
		exec.Command("docker", "rm", "-f", id).Run()

		// remove metadata
		home, _ := os.UserHomeDir()
		runDir := filepath.Join(home, ".urx", "runs", id)
		os.RemoveAll(runDir)

		fmt.Println("[URX] Removed:", id)
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
