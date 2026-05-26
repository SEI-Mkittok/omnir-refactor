package worker

import (
	"errors"
	"testing"
	"time"
)

func TestSchedulerRegistryTracksRuns(t *testing.T) {
	registry := NewSchedulerRegistry()
	registry.Register("mail", "Mail sync", time.Minute, true)

	runID := registry.Start("mail")
	registry.Finish("mail", runID, nil, time.Minute)

	jobs := registry.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("expected one job, got %d", len(jobs))
	}
	if jobs[0].LastResult == nil || *jobs[0].LastResult != "succeeded" {
		t.Fatalf("expected succeeded result, got %#v", jobs[0].LastResult)
	}
	runs := registry.Runs("mail")
	if len(runs) != 1 || runs[0].Status != "succeeded" {
		t.Fatalf("expected succeeded run, got %#v", runs)
	}
}

func TestSchedulerRegistryTracksErrors(t *testing.T) {
	registry := NewSchedulerRegistry()
	registry.Register("mail", "Mail sync", time.Minute, true)

	runID := registry.Start("mail")
	registry.Finish("mail", runID, errors.New("boom"), time.Minute)

	jobs := registry.Jobs()
	if jobs[0].ErrorText == nil || *jobs[0].ErrorText != "boom" {
		t.Fatalf("expected job error text, got %#v", jobs[0].ErrorText)
	}
}
