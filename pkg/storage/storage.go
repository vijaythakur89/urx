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

// base dir: ~/.urx/runs/<id>
func GetRunDir(id string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".urx", "runs", id)
}

// save metadata safely
func SaveMeta(id string, meta RunMeta) error {
	dir := GetRunDir(id)
	if dir == "" {
		return nil
	}

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	file := filepath.Join(dir, "meta.json")

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, data, 0644)
}

// log file path helper
func LogFilePath(id string) string {
	return filepath.Join(GetRunDir(id), "logs.txt")
}

// load all metadata
func LoadAllMeta() ([]RunMeta, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	runsDir := filepath.Join(home, ".urx", "runs")

	files, err := os.ReadDir(runsDir)
	if err != nil {
		// ✅ important: handle first run (directory doesn't exist)
		if os.IsNotExist(err) {
			return []RunMeta{}, nil
		}
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
