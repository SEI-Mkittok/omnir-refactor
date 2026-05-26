package worker

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

type SchedulerRegistry struct {
	mu   sync.RWMutex
	jobs map[string]*domain.SchedulerJobStatus
	runs map[string][]domain.SchedulerJobRun
}

func NewSchedulerRegistry() *SchedulerRegistry {
	return &SchedulerRegistry{
		jobs: map[string]*domain.SchedulerJobStatus{},
		runs: map[string][]domain.SchedulerJobRun{},
	}
}

func (r *SchedulerRegistry) Register(key, name string, interval time.Duration, enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	next := time.Now().UTC().Add(interval)
	r.jobs[key] = &domain.SchedulerJobStatus{
		Key:      key,
		Name:     name,
		Interval: interval.String(),
		Enabled:  enabled,
		NextRun:  &next,
	}
}

func (r *SchedulerRegistry) Start(key string) uuid.UUID {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	runID := uuid.New()
	if job := r.jobs[key]; job != nil {
		job.LastStarted = &now
		job.ErrorText = nil
	}
	r.runs[key] = append([]domain.SchedulerJobRun{{
		ID:        runID,
		JobKey:    key,
		Status:    "running",
		StartedAt: now,
	}}, r.runs[key]...)
	if len(r.runs[key]) > 50 {
		r.runs[key] = r.runs[key][:50]
	}
	return runID
}

func (r *SchedulerRegistry) Finish(key string, runID uuid.UUID, err error, interval time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	status := "succeeded"
	var errText *string
	if err != nil {
		status = "failed"
		text := err.Error()
		errText = &text
	}
	if job := r.jobs[key]; job != nil {
		result := status
		job.LastFinished = &now
		job.LastResult = &result
		job.ErrorText = errText
		next := now.Add(interval)
		job.NextRun = &next
	}
	for i := range r.runs[key] {
		if r.runs[key][i].ID == runID {
			r.runs[key][i].Status = status
			r.runs[key][i].FinishedAt = &now
			r.runs[key][i].ErrorText = errText
			return
		}
	}
}

func (r *SchedulerRegistry) TrackRun(key string, interval time.Duration, fn func() error) error {
	if r == nil || key == "" {
		return fn()
	}
	runID := r.Start(key)
	err := fn()
	r.Finish(key, runID, err, interval)
	return err
}

func (r *SchedulerRegistry) Jobs() []domain.SchedulerJobStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.SchedulerJobStatus, 0, len(r.jobs))
	for _, job := range r.jobs {
		out = append(out, *job)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func (r *SchedulerRegistry) Runs(key string) []domain.SchedulerJobRun {
	r.mu.RLock()
	defer r.mu.RUnlock()
	runs := r.runs[key]
	out := make([]domain.SchedulerJobRun, len(runs))
	copy(out, runs)
	return out
}
