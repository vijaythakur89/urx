package events

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Emitter writes URX lifecycle events to a JSONL file.
type Emitter struct {
	filePath string
}

// NewEmitter creates an event emitter for a specific event file.
func NewEmitter(filePath string) (*Emitter, error) {
	dir := filepath.Dir(filePath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &Emitter{
		filePath: filePath,
	}, nil
}

// Emit appends an event to the JSONL event log.
func (e *Emitter) Emit(event Event) error {
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(
		e.filePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}

	return nil
}
