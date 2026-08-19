package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vijaythakur89/urx/pkg/events"
	"github.com/vijaythakur89/urx/pkg/storage"
)

type HistoryOutput struct {
	ID        string
	Status    string
	Port      int
	Age       string
	Timestamp string
}

var historyLimit int

func getRunStatus(id string) string {
	eventFile := storage.EventFilePath(id)

	file, err := os.Open(eventFile)
	if err != nil {
		return "-"
	}
	defer file.Close()

	status := "-"

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var event events.Event

		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}

		switch event.Type {
		case events.RunStarted:
			status = "RUNNING"

		case events.RunCompleted:
			status = "COMPLETED"

		case events.RunFailed:
			status = "FAILED"
		}
	}

	return status
}

func getRunAge(timestamp string) string {
	if timestamp == "" {
		return "-"
	}

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "-"
	}

	return time.Since(t).Round(time.Second).String()
}
func buildHistoryResults(metas []storage.RunMeta) []HistoryOutput {
	var results []HistoryOutput

	for _, meta := range metas {
		results = append(results, HistoryOutput{
			ID:        meta.ID,
			Status:    getRunStatus(meta.ID),
			Port:      meta.Port,
			Age:       getRunAge(meta.Timestamp),
			Timestamp: meta.Timestamp,
		})
	}

	// Sort newest runs first.
	sort.Slice(results, func(i, j int) bool {
		ti, errI := time.Parse(time.RFC3339, results[i].Timestamp)
		tj, errJ := time.Parse(time.RFC3339, results[j].Timestamp)

		if errI != nil {
			return false
		}

		if errJ != nil {
			return true
		}

		return ti.After(tj)
	})

	return results
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show URX run history",

	Run: func(cmd *cobra.Command, args []string) {

		//load metadata
		metas, err := storage.LoadAllMeta()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if len(metas) == 0 {
			fmt.Println("No URX runs found.")
			return
		}

		results := buildHistoryResults(metas)

		if historyLimit > 0 && len(results) > historyLimit {
			results = results[:historyLimit]
		}

		// Apply history limit after sorting.
		if historyLimit > 0 && len(results) > historyLimit {
			results = results[:historyLimit]
		}

		fmt.Println("URX RUN HISTORY")
		fmt.Println()
		fmt.Println("ID\tSTATUS\tPORT\tAGE")
		fmt.Println(strings.Repeat("-", 70))

		for _, r := range results {
			port := "-"
			if r.Port != 0 {
				port = fmt.Sprintf("%d", r.Port)
			}

			fmt.Printf("%-25s %-12s %-8s %-10s\n",
				r.ID,
				r.Status,
				port,
				r.Age,
			)
		}
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.Flags().IntVar(
		&historyLimit,
		"limit",
		10,
		"Limit number of history entries",
	)
}
