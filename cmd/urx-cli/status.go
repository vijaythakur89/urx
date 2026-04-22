package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [id]",
	Short: "Show detailed status of a URX run",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		id := args[0]

		// -------------------------
		// Inspect container
		// -------------------------
		out, err := exec.Command(
			"docker", "inspect", id,
			"--format",
			"{{.State.Status}}|{{.Config.Image}}|{{.State.StartedAt}}",
		).Output()

		if err != nil {
			fmt.Println("Error: container not found")
			os.Exit(1)
		}

		parts := strings.Split(strings.TrimSpace(string(out)), "|")

		status := parts[0]
		image := parts[1]
		started := parts[2]

		// -------------------------
		// Health
		// -------------------------
		health := getHealth(id)

		// -------------------------
		// Logs preview
		// -------------------------
		logsOut, _ := exec.Command("docker", "logs", "--tail", "5", id).Output()

		// -------------------------
		// Output
		// -------------------------
		fmt.Println("ID:       ", id)
		fmt.Println("Status:   ", status)
		fmt.Println("Health:   ", health)
		fmt.Println("Image:    ", image)
		fmt.Println("Started:  ", started)

		fmt.Println("\n--- Logs (last 5 lines) ---")
		fmt.Println(string(logsOut))
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
