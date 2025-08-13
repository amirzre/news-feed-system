package service

import (
	"context"
	"sync"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/pkg/logger"
)

type scheduledJob struct {
	name     string
	interval time.Duration
	job      func(context.Context) error
	ticker   *time.Ticker
	status   model.JobStatus
	mu       sync.RWMutex
}

// schedulerService implements SchedulerService interface
type schedulerService struct {
	jobs    map[string]*scheduledJob
	mu      sync.RWMutex
	logger  *logger.Logger
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(logger *logger.Logger) SchedulerService {
	return &schedulerService{
		jobs:   make(map[string]*scheduledJob),
		logger: logger,
	}
}

// Start starts the scheduler service
func (s *schedulerService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	s.logger.Info("Starting scheduler service")

	for _, job := range s.jobs {
		s.startJob(job)
	}

	s.logger.Info("Scheduler service started", "jobs_count", len(s.jobs))

	return nil
}

// Stop stops the scheduler service
func (s *schedulerService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.logger.Info("Stopping scheduler service")

	for _, job := range s.jobs {
		if job.ticker != nil {
			job.ticker.Stop()
		}
	}

	s.cancel()
	s.wg.Wait()

	s.running = false
	s.logger.Info("Scheduler service stopped")
	return nil
}

// IsRunning returns whether the scheduler is running
func (s *schedulerService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// AddJob adds a new scheduled job
func (s *schedulerService) AddJob(name string, interval time.Duration, job func(context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop existing job if it exists
	if existingJob, exists := s.jobs[name]; exists {
		if existingJob.ticker != nil {
			existingJob.ticker.Stop()
		}
	}

	// Create new job
	scheduledJob := &scheduledJob{
		name:     name,
		interval: interval,
		job:      job,
		status: model.JobStatus{
			Name:     name,
			Interval: interval,
		},
	}

	s.jobs[name] = scheduledJob

	// Start the job if scheduler is running
	if s.running {
		s.startJob(scheduledJob)
	}

	s.logger.Info("Added scheduled job",
		"name", name,
		"interval", interval.String(),
	)
}

// RemoveJob removes a scheduled job
func (s *schedulerService) RemoveJob(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, exists := s.jobs[name]; exists {
		if job.ticker != nil {
			job.ticker.Stop()
		}
		delete(s.jobs, name)
		s.logger.Info("Removed scheduled job", "name", name)
	}
}

// GetJobStatus returns the status of all jobs
func (s *schedulerService) GetJobStatus() map[string]model.JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]model.JobStatus)
	for name, job := range s.jobs {
		job.mu.RLock()
		status[name] = job.status
		job.mu.RUnlock()
	}

	return status
}

// startJob starts a single job
func (s *schedulerService) startJob(job *scheduledJob) {
	job.ticker = time.NewTicker(job.interval)

	job.mu.Lock()
	nextRun := time.Now().Add(job.interval)
	job.status.NextRun = &nextRun
	job.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer job.ticker.Stop()

		s.logger.Info("Started scheduled job",
			"name", job.name,
			"interval", job.interval.String(),
			"next_run", nextRun.Format(time.RFC3339),
		)

		for {
			select {
			case <-s.ctx.Done():
				s.logger.Info("Stopping job due to context cancellation", "name", job.name)
				return
			case <-job.ticker.C:
				s.executeJob(job)
			}
		}
	}()
}

// executeJob executes a single job run with error handling and metrics
func (s *schedulerService) executeJob(job *scheduledJob) {
	start := time.Now()

	job.mu.Lock()
	job.status.IsRunning = true
	job.status.RunCount++
	runCount := job.status.RunCount
	job.mu.Unlock()

	s.logger.Info("Executing scheduled job",
		"name", job.name,
		"run_count", runCount,
	)

	// Create a timeout context for the job execution
	jobCtx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	// Execute the job
	err := job.job(jobCtx)
	duration := time.Since(start)

	// Update job status
	job.mu.Lock()
	job.status.IsRunning = false
	job.status.LastRun = &start

	// Calculate next run time
	nextRun := time.Now().Add(job.interval)
	job.status.NextRun = &nextRun

	// Update average run time
	if job.status.AverageRunTime == 0 {
		job.status.AverageRunTime = duration
	} else {
		// Simple moving average
		job.status.AverageRunTime = (job.status.AverageRunTime + duration) / 2
	}

	if err != nil {
		job.status.ErrorCount++
		job.status.LastError = err.Error()
		job.mu.Unlock()

		s.logger.Error("Scheduled job failed",
			"name", job.name,
			"error", err.Error(),
			"durationMS", duration.Milliseconds(),
			"run_count", runCount,
			"error_count", job.status.ErrorCount,
		)
	} else {
		job.status.LastError = ""
		job.mu.Unlock()

		s.logger.Info("Scheduled job completed successfully",
			"name", job.name,
			"durationMS", duration.Milliseconds(),
			"run_count", runCount,
		)
	}
}
