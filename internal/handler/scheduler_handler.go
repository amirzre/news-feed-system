package handler

import (
	"net/http"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/internal/service"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/amirzre/news-feed-system/pkg/response"
	"github.com/labstack/echo/v4"
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

// GetStatus handles GET /api/v1/scheduler/status
func (h *schedulerHandler) GetStatus(c echo.Context) error {
	isRunning := h.schedulerService.IsRunning()
	jobStatus := h.schedulerService.GetJobStatus()

	statusData := model.SchedulerStatusResponse{
		SchedulerRunning: isRunning,
		JobsCount:        len(jobStatus),
		Timestamp:        time.Now(),
		Jobs:             jobStatus,
	}

	return response.Success(c, http.StatusOK, statusData, "Scheduler status retrieved successfully")
}
