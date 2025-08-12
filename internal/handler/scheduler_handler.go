package handler

import (
	"github.com/amirzre/news-feed-system/internal/service"
	"github.com/amirzre/news-feed-system/pkg/logger"
)

// schedulerHandler implements SchedulerHandler interface
type schedulerHandler struct {
	schedulerService service.SchedulerService
	logger           *logger.Logger
}

// NewSchedulerHandler creates a new scheduler handler
func NewSchedulerHandler(schedulerService service.SchedulerService, logger *logger.Logger) SchedulerHandler {
	return &schedulerHandler{
		schedulerService: schedulerService,
		logger:           logger,
	}
}
