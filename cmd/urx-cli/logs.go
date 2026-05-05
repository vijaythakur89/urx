package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var logsJSON bool

type LogsOutput struct {
	ID   string   `json:"id"`
	Logs []string `json:"logs"`
}

var logsCmd = &cobra.Command{
	Use:   "logs [id]",
	Short: "View logs of a URX container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		id := args[0]

		out, err := exec.Command("docker", "logs", id).Output()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		logText := strings.TrimSpace(string(out))

		// JSON output
		if logsJSON {
			lines := []string{}
			if logText != "" {
				lines = strings.Split(logText, "\n")
			}

			result := LogsOutput{
				ID:   id,
				Logs: lines,
			}

			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			fmt.Println(string(data))
			return
		}

		// default output
		fmt.Println(logText)
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVar(&logsJSON, "json", false, "Output in JSON format")
}
