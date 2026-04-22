package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [id]",
	Short: "Stop a running URX container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		id := args[0]

		stopCmd := exec.Command("docker", "stop", id)
		stopCmd.Stdout = os.Stdout
		stopCmd.Stderr = os.Stderr

		err := stopCmd.Run()
		if err != nil {
			fmt.Println("[URX] Error stopping container:", err)
			return
		}

		fmt.Println("[URX] Stopped:", id)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
