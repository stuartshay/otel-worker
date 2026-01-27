package queue

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewQueue(t *testing.T) {
	processor := func(_ context.Context, _ *Job) (*JobResult, error) {
		return &JobResult{CSVPath: "/data/test.csv"}, nil
	}

	q := NewQueue(3, processor)
	defer func() { _ = q.Shutdown(time.Second) }()

	if q.workers != 3 {
		t.Errorf("expected 3 workers, got %d", q.workers)
	}
	if len(q.jobs) != 0 {
		t.Errorf("expected empty jobs map, got %d jobs", len(q.jobs))
	}
}

func TestEnqueue(t *testing.T) {
	processor := func(_ context.Context, _ *Job) (*JobResult, error) {
		return &JobResult{CSVPath: "/data/test.csv"}, nil
	}

	q := NewQueue(1, processor)
	defer func() { _ = q.Shutdown(time.Second) }()

	jobID, err := q.Enqueue("2026-01-24", "test-device")
	if err != nil {
		t.Fatalf("Enqueue() failed: %v", err)
	}

	if jobID == "" {
		t.Error("expected non-empty job ID")
	}

	// Verify job was stored
	job, err := q.GetJob(jobID)
	if err != nil {
		t.Fatalf("GetJob() failed: %v", err)
	}

	if job.Date != "2026-01-24" {
		t.Errorf("expected date '2026-01-24', got '%s'", job.Date)
	}
	if job.DeviceID != "test-device" {
		t.Errorf("expected device_id 'test-device', got '%s'", job.DeviceID)
	}
	if job.Status != StatusQueued && job.Status != StatusProcessing {
		t.Errorf("expected status 'queued' or 'processing', got '%s'", job.Status)
	}
}

func TestGetJob_NotFound(t *testing.T) {
	processor := func(_ context.Context, _ *Job) (*JobResult, error) {
		return nil, nil
	}

	q := NewQueue(1, processor)
	defer func() { _ = q.Shutdown(time.Second) }()

	_, err := q.GetJob("non-existent-id")
	if err == nil {
		t.Error("expected error for non-existent job, got nil")
	}
}

func TestListJobs(t *testing.T) {
	processor := func(_ context.Context, _ *Job) (*JobResult, error) {
		time.Sleep(50 * time.Millisecond)
		return &JobResult{CSVPath: "/data/test.csv"}, nil
	}

	q := NewQueue(1, processor)
	defer func() { _ = q.Shutdown(time.Second) }()

	// Enqueue multiple jobs
	_, _ = q.Enqueue("2026-01-24", "device1")
	_, _ = q.Enqueue("2026-01-25", "device2")
	_, _ = q.Enqueue("2026-01-26", "device3")

	// List all jobs
	jobs := q.ListJobs("", 10, 0)
	if len(jobs) == 0 {
		t.Error("expected at least 1 job, got 0")
	}

	// List with pagination
	jobs = q.ListJobs("", 1, 0)
	if len(jobs) != 1 {
		t.Errorf("expected 1 job with limit=1, got %d", len(jobs))
	}

	// List with offset beyond available
	jobs = q.ListJobs("", 10, 100)
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs with offset=100, got %d", len(jobs))
	}
}

func TestProcessJob_Success(t *testing.T) {
	var processorCalled atomic.Bool
	processor := func(_ context.Context, _ *Job) (*JobResult, error) {
		processorCalled.Store(true)
		return &JobResult{
			CSVPath:         "/data/distance_20260124.csv",
			TotalDistanceKM: 25.5,
			MaxDistanceKM:   15.2,
			MinDistanceKM:   0.1,
			TotalLocations:  100,
		}, nil
	}

	q := NewQueue(1, processor)
	defer func() { _ = q.Shutdown(time.Second) }()

	jobID, _ := q.Enqueue("2026-01-24", "test-device")

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if !processorCalled.Load() {
		t.Error("expected processor to be called")
	}

	job, _ := q.GetJob(jobID)
	if job.Status != StatusCompleted {
		t.Errorf("expected status 'completed', got '%s'", job.Status)
	}
	if job.Result == nil {
		t.Fatal("expected non-nil result")
	}
	if job.Result.TotalDistanceKM != 25.5 {
		t.Errorf("expected TotalDistanceKM 25.5, got %.2f", job.Result.TotalDistanceKM)
	}
}

func TestProcessJob_Failure(t *testing.T) {
	processor := func(_ context.Context, _ *Job) (*JobResult, error) {
		return nil, errors.New("processing failed")
	}

	q := NewQueue(1, processor)
	defer func() { _ = q.Shutdown(time.Second) }()

	jobID, _ := q.Enqueue("2026-01-24", "test-device")

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	job, _ := q.GetJob(jobID)
	if job.Status != StatusFailed {
		t.Errorf("expected status 'failed', got '%s'", job.Status)
	}
	if job.ErrorMessage == "" {
		t.Error("expected non-empty error message")
	}
}

func TestGetStats(t *testing.T) {
	processor := func(_ context.Context, _ *Job) (*JobResult, error) {
		time.Sleep(50 * time.Millisecond)
		return &JobResult{}, nil
	}

	q := NewQueue(1, processor)
	defer func() { _ = q.Shutdown(time.Second) }()

	// Enqueue jobs
	_, _ = q.Enqueue("2026-01-24", "device1")
	_, _ = q.Enqueue("2026-01-25", "device2")

	stats := q.GetStats()
	if stats["total"] < 2 {
		t.Errorf("expected total >= 2, got %d", stats["total"])
	}
}

func TestShutdown(t *testing.T) {
	processor := func(_ context.Context, _ *Job) (*JobResult, error) {
		return &JobResult{}, nil
	}

	q := NewQueue(3, processor)

	err := q.Shutdown(time.Second)
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}

	// Verify queue is stopped
	select {
	case <-q.ctx.Done():
		// Expected
	default:
		t.Error("expected context to be canceled")
	}
}
