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

// flag for json output
var outputJSON bool

// response struct
type PsOutput struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Health string `json:"health"`
	Port   int    `json:"port"`
	Age    string `json:"age"`
}

// health check
func getHealth(container string) string {
	out, err := exec.Command(
		"docker", "exec", container,
		"sh", "-c", "test -f /tmp/urx_health && echo healthy || echo unhealthy",
	).Output()

	if err != nil {
		return "-" // keep consistent with your current behavior
	}

	return strings.TrimSpace(string(out))
}

// ps command
var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List running URX containers",
	Run: func(cmd *cobra.Command, args []string) {

		// get docker ps output
		out, err := exec.Command(
			"docker", "ps", "-a",
			"--format", "{{.Names}} {{.Status}}",
		).Output()

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		lines := strings.Split(strings.TrimSpace(string(out)), "\n")

		// load metadata
		metas, _ := storage.LoadAllMeta()
		metaMap := make(map[string]storage.RunMeta)

		for _, m := range metas {
			metaMap[m.ID] = m
		}

		// collect results
		var results []PsOutput

		for _, line := range lines {
			if line == "" {
				continue
			}

			parts := strings.SplitN(line, " ", 2)
			if len(parts) < 2 {
				continue
			}

			id := strings.TrimSpace(parts[0])
			status := parts[1]

			// only urx containers
			if !strings.HasPrefix(id, "urx-") {
				continue
			}

			health := getHealth(id)

			meta, ok := metaMap[id]
			
			if !ok {
				meta = storage.RunMeta{} // fallback
			}

			// calculate age
			age := "-"
			if meta.Timestamp != "" {
				t, _ := time.Parse(time.RFC3339, meta.Timestamp)
				age = time.Since(t).Round(time.Second).String()
			}

			results = append(results, PsOutput{
				ID:     id,
				Status: status,
				Health: health,
				Port:   meta.Port,
				Age:    age,
			})
		}

		// JSON output
		if outputJSON {
			data, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(data))
			return
		}

		// default table output
		fmt.Println("ID\tSTATUS\tHEALTH\tPORT\tAGE")
		fmt.Println("-------------------------------------------------------")

		for _, r := range results {
			port := "-"
			if r.Port != 0 {
				port = fmt.Sprintf("%d", r.Port)
			}

			fmt.Printf("%-25s %-15s %-10s %-8s %-10s\n",
				r.ID, r.Status, r.Health, port, r.Age)
		}
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
	psCmd.Flags().BoolVar(&outputJSON, "json", false, "Output in JSON format")
}
