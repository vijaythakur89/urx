package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List URX runs with status",
	Run: func(cmd *cobra.Command, args []string) {

		home, _ := os.UserHomeDir()
		runsDir := filepath.Join(home, ".urx", "runs")

		files, err := os.ReadDir(runsDir)
		if err != nil {
			fmt.Println("No runs found")
			return
		}

		fmt.Printf("%-25s %-10s\n", "ID", "STATUS")
		fmt.Println("-----------------------------------------")

		for _, f := range files {
			id := f.Name()

			status := getContainerStatus(id)

			fmt.Printf("%-25s %-10s\n", id, status)
		}
	},
}

func getContainerStatus(id string) string {

	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", id)

	output, err := cmd.Output()
	if err != nil {
		return "not_found"
	}

	return strings.TrimSpace(string(output))
}

func init() {
	rootCmd.AddCommand(psCmd)
}
