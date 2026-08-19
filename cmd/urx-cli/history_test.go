package main

import (
	"testing"
	"time"

	"github.com/vijaythakur89/urx/pkg/storage"
)

func TestGetRunAge(t *testing.T) {
	timestamp := time.Now().Add(-2 * time.Minute).Format(time.RFC3339)

	age := getRunAge(timestamp)

	if age == "-" {
		t.Fatal("expected valid age")
	}
}

func TestGetRunAgeInvalid(t *testing.T) {
	age := getRunAge("invalid-timestamp")

	if age != "-" {
		t.Fatalf("expected -, got %s", age)
	}
}
func TestHistoryOrdering(t *testing.T) {
	metas := []storage.RunMeta{
		{
			ID:        "older-run",
			Timestamp: "2026-08-18T10:00:00Z",
		},
		{
			ID:        "newer-run",
			Timestamp: "2026-08-19T10:00:00Z",
		},
	}

	results := buildHistoryResults(metas)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].ID != "newer-run" {
		t.Fatalf("expected newest run first, got %s", results[0].ID)
	}

	if results[1].ID != "older-run" {
		t.Fatalf("expected older run second, got %s", results[1].ID)
	}
}
