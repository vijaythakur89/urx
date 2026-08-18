package events

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEventCreation(t *testing.T) {
	now := time.Now()

	event := Event{
		ID:        "evt-123",
		Timestamp: now,
		Type:      RunStarted,
		Component: "runtime",
		RunID:     "urx-test",
		Message:   "URX application started",
	}

	if event.ID != "evt-123" {
		t.Fatalf("expected event ID evt-123, got %s", event.ID)
	}

	if event.Type != RunStarted {
		t.Fatalf("expected event type %s, got %s", RunStarted, event.Type)
	}

	if event.Component != "runtime" {
		t.Fatalf("expected component runtime, got %s", event.Component)
	}

	if event.RunID != "urx-test" {
		t.Fatalf("expected run ID urx-test, got %s", event.RunID)
	}

	if event.Timestamp.IsZero() {
		t.Fatal("expected timestamp to be set")
	}
}

func TestEventTypes(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"run started", RunStarted, "run.started"},
		{"run completed", RunCompleted, "run.completed"},
		{"run failed", RunFailed, "run.failed"},
		{"deploy started", DeployStarted, "deploy.started"},
		{"deploy completed", DeployCompleted, "deploy.completed"},
		{"deploy failed", DeployFailed, "deploy.failed"},
		{"container started", ContainerStarted, "container.started"},
		{"container failed", ContainerFailed, "container.failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, tt.got)
			}
		})
	}
}
func TestEmitter(t *testing.T) {
	runDir := t.TempDir()

	eventFile := filepath.Join(runDir, "events.jsonl")

	emitter, err := NewEmitter(eventFile)
	if err != nil {
		t.Fatalf("failed to create emitter: %v", err)
	}

	event := Event{
		Type:      RunStarted,
		Component: "runtime",
		RunID:     "urx-test",
		Message:   "URX application started",
	}

	if err := emitter.Emit(event); err != nil {
		t.Fatalf("failed to emit event: %v", err)
	}

	data, err := os.ReadFile(eventFile)
	if err != nil {
		t.Fatalf("failed to read event file: %v", err)
	}

	var stored Event

	if err := json.Unmarshal(bytes.TrimSpace(data), &stored); err != nil {
		t.Fatalf("failed to decode stored event: %v", err)
	}

	if stored.Type != RunStarted {
		t.Fatalf("expected type %s, got %s", RunStarted, stored.Type)
	}

	if stored.RunID != "urx-test" {
		t.Fatalf("expected run ID urx-test, got %s", stored.RunID)
	}

	if stored.ID == "" {
		t.Fatal("expected generated event ID")
	}

	if stored.Timestamp.IsZero() {
		t.Fatal("expected generated timestamp")
	}
}
