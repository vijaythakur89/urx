package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List URX runs",
	Run: func(cmd *cobra.Command, args []string) {

		out, err := exec.Command("docker", "ps", "-a", "--format", "{{.Names}} {{.Status}}").Output()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Printf("%-25s %-12s %-10s\n", "ID", "STATUS", "HEALTH")
		fmt.Println("-------------------------------------------------------")

		lines := strings.Split(string(out), "\n")

		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}

			parts := strings.SplitN(line, " ", 2)
			name := parts[0]
			status := parts[1]

			if !strings.HasPrefix(name, "urx-") {
				continue
			}

			health := getHealth(name)

			fmt.Printf("%-25s %-12s %-10s\n", name, simplifyStatus(status), health)
		}
	},
}

func simplifyStatus(s string) string {
	if strings.HasPrefix(s, "Up") {
		return "running"
	}
	if strings.HasPrefix(s, "Exited") {
		return "exited"
	}
	return "unknown"
}

func getHealth(container string) string {

	out, err := exec.Command(
		"docker", "exec", container,
		"sh", "-c", "test -f /tmp/urx_health && echo healthy || echo unhealthy",
	).Output()

	if err != nil {
		return "-"
	}

	return strings.TrimSpace(string(out))
}

func init() {
	rootCmd.AddCommand(psCmd)
}
