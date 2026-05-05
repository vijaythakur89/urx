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
	Port      int    `json:"port"`
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

func LoadAllMeta() ([]RunMeta, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	runsDir := filepath.Join(home, ".urx", "runs")

	files, err := os.ReadDir(runsDir)
	if err != nil {
		return nil, err
	}

	var metas []RunMeta

	for _, f := range files {
		metaPath := filepath.Join(runsDir, f.Name(), "meta.json")

		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}

		var m RunMeta
		err = json.Unmarshal(data, &m)
		if err != nil {
			continue
		}

		metas = append(metas, m)
	}

	return metas, nil
}
