package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type RunMeta struct {
	ID        string `json:"id"`
	Artifact  string `json:"artifact"`
	Timestamp string `json:"timestamp"`
}

func GetRunDir(id string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".urx", "runs", id)
}

func SaveMeta(id string, meta RunMeta) error {
	dir := GetRunDir(id)
	os.MkdirAll(dir, 0755)

	file := filepath.Join(dir, "meta.json")

	data, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(file, data, 0644)
}

func LogFilePath(id string) string {
	return filepath.Join(GetRunDir(id), "logs.txt")
}
