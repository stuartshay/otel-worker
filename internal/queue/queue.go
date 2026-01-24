// Package queue provides an in-memory job queue system with worker pool
// for concurrent processing of distance calculation tasks.
package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the state of a distance calculation job
type JobStatus string

// Job status constants define the lifecycle states
const (
	StatusQueued     JobStatus = "queued"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

// Job represents a distance calculation job
type Job struct {
	ID           string
	Date         string
	DeviceID     string
	Status       JobStatus
	QueuedAt     time.Time
	StartedAt    *time.Time
	CompletedAt  *time.Time
	ErrorMessage string
	Result       *JobResult
}

// JobResult contains the output of a completed distance calculation
type JobResult struct {
	CSVPath          string
	TotalDistanceKM  float64
	MaxDistanceKM    float64
	MinDistanceKM    float64
	TotalLocations   int
	ProcessingTimeMS int64
}

// ProcessFunc is a function that processes a job
type ProcessFunc func(ctx context.Context, job *Job) (*JobResult, error)

// Queue manages distance calculation jobs with a worker pool
type Queue struct {
	mu           sync.RWMutex
	jobs         map[string]*Job
	pendingQueue chan *Job
	workers      int
	processor    ProcessFunc
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewQueue creates a new job queue with the specified number of workers
func NewQueue(workers int, processor ProcessFunc) *Queue {
	ctx, cancel := context.WithCancel(context.Background())
	q := &Queue{
		jobs:         make(map[string]*Job),
		pendingQueue: make(chan *Job, 100),
		workers:      workers,
		processor:    processor,
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start worker pool
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	return q
}

// Enqueue adds a new job to the queue
func (q *Queue) Enqueue(date, deviceID string) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Generate unique job ID
	jobID := uuid.New().String()

	// Create job
	job := &Job{
		ID:       jobID,
		Date:     date,
		DeviceID: deviceID,
		Status:   StatusQueued,
		QueuedAt: time.Now().UTC(),
	}

	// Store job
	q.jobs[jobID] = job

	// Add to pending queue (non-blocking)
	select {
	case q.pendingQueue <- job:
		return jobID, nil
	default:
		job.Status = StatusFailed
		job.ErrorMessage = "queue is full"
		return "", fmt.Errorf("queue is full")
	}
}

// GetJob retrieves a job by ID
func (q *Queue) GetJob(jobID string) (*Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	job, exists := q.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	// Return a copy to prevent external mutation
	jobCopy := *job
	if job.StartedAt != nil {
		startedCopy := *job.StartedAt
		jobCopy.StartedAt = &startedCopy
	}
	if job.CompletedAt != nil {
		completedCopy := *job.CompletedAt
		jobCopy.CompletedAt = &completedCopy
	}
	if job.Result != nil {
		resultCopy := *job.Result
		jobCopy.Result = &resultCopy
	}

	return &jobCopy, nil
}

// ListJobs returns jobs filtered by status
func (q *Queue) ListJobs(status JobStatus, limit, offset int) []*Job {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var filtered []*Job
	for _, job := range q.jobs {
		if status == "" || job.Status == status {
			jobCopy := *job
			filtered = append(filtered, &jobCopy)
		}
	}

	// Sort by queue time (newest first)
	// For production, consider using a more efficient data structure

	// Apply pagination
	start := offset
	if start > len(filtered) {
		return []*Job{}
	}

	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end]
}

// GetStats returns queue statistics
func (q *Queue) GetStats() map[string]int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	stats := map[string]int{
		"total":      len(q.jobs),
		"queued":     0,
		"processing": 0,
		"completed":  0,
		"failed":     0,
	}

	for _, job := range q.jobs {
		switch job.Status {
		case StatusQueued:
			stats["queued"]++
		case StatusProcessing:
			stats["processing"]++
		case StatusCompleted:
			stats["completed"]++
		case StatusFailed:
			stats["failed"]++
		}
	}

	return stats
}

// worker processes jobs from the queue
// nolint:unparam // id parameter reserved for future debugging/logging
func (q *Queue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			return
		case job := <-q.pendingQueue:
			q.processJob(job)
		}
	}
}

// processJob executes a single job
func (q *Queue) processJob(job *Job) {
	startTime := time.Now()

	// Update status to processing
	q.mu.Lock()
	job.Status = StatusProcessing
	now := time.Now().UTC()
	job.StartedAt = &now
	q.mu.Unlock()

	// Process the job
	result, err := q.processor(q.ctx, job)

	// Update job with result
	q.mu.Lock()
	defer q.mu.Unlock()

	completedAt := time.Now().UTC()
	job.CompletedAt = &completedAt

	if err != nil {
		job.Status = StatusFailed
		job.ErrorMessage = err.Error()
	} else {
		job.Status = StatusCompleted
		job.Result = result
		if result != nil {
			result.ProcessingTimeMS = time.Since(startTime).Milliseconds()
		}
	}
}

// Shutdown gracefully shuts down the queue
func (q *Queue) Shutdown(timeout time.Duration) error {
	// Stop accepting new jobs
	q.cancel()

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("shutdown timeout exceeded")
	}
}
