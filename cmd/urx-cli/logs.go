package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var follow bool

var logsCmd = &cobra.Command{
	Use:   "logs [id]",
	Short: "Show logs for a URX run",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		id := args[0]

		var logCmd *exec.Cmd

		if follow {
			logCmd = exec.Command("docker", "logs", "-f", id)
		} else {
			logCmd = exec.Command("docker", "logs", id)
		}

		logCmd.Stdout = os.Stdout
		logCmd.Stderr = os.Stderr

		err := logCmd.Run()
		if err != nil {
			fmt.Println("[URX] Error fetching logs:", err)
		}
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow logs output")
	rootCmd.AddCommand(logsCmd)
}
