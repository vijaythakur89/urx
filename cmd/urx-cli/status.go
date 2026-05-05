package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vijaythakur89/urx/pkg/storage"
)

var statusJSON bool

type StatusOutput struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Health string `json:"health"`
	Port   int    `json:"port"`
	Age    string `json:"age"`
}

// reuse health check
func getContainerHealth(id string) string {
	out, err := exec.Command(
		"docker", "exec", id,
		"sh", "-c", "test -f /tmp/urx_health && echo healthy || echo unhealthy",
	).Output()

	if err != nil {
		return "-"
	}

	return strings.TrimSpace(string(out))
}

var statusCmd = &cobra.Command{
	Use:   "status [id]",
	Short: "Get status of a URX container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		id := args[0]

		// docker inspect for status
		out, err := exec.Command(
			"docker", "inspect",
			"--format", "{{.State.Status}}",
			id,
		).Output()

		if err != nil {
			fmt.Println("Error: container not found")
			return
		}

		status := strings.TrimSpace(string(out))
		health := getContainerHealth(id)

		// load metadata
		metaList, _ := storage.LoadAllMeta()

		var meta storage.RunMeta
		for _, m := range metaList {
			if m.ID == id {
				meta = m
				break
			}
		}

		// calculate age
		age := "-"
		if meta.Timestamp != "" {
			t, err := time.Parse(time.RFC3339, meta.Timestamp)
			if err == nil {
				age = time.Since(t).Round(time.Second).String()
			}
		}

		output := StatusOutput{
			ID:     id,
			Status: status,
			Health: health,
			Port:   meta.Port,
			Age:    age,
		}

		// JSON output
		if statusJSON {
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Println(string(data))
			return
		}

		// default output
		fmt.Println("ID:     ", output.ID)
		fmt.Println("Status: ", output.Status)
		fmt.Println("Health: ", output.Health)
		fmt.Println("Port:   ", output.Port)
		fmt.Println("Age:    ", output.Age)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output in JSON format")
}
