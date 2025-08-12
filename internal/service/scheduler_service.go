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
