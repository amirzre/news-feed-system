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
// @Summary      Get scheduler status
// @Description  Retrieve current scheduler status including running flag and job list
// @Tags         scheduler
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.APIResponse{data=model.SchedulerStatusResponse}  "Scheduler status"
// @Failure      500  {object}  response.APIResponse{error=response.ErrorInfo}		       "Internal server error"
// @Router       /scheduler/status [get]
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

// GetJobs handles GET /api/v1/scheduler/jobs
// @Summary      List scheduler jobs
// @Description  Retrieve list of scheduled jobs and their statuses
// @Tags         scheduler
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.APIResponse{data=model.JobsResponse}	"Jobs list"
// @Failure      500  {object}  response.APIResponse{error=response.ErrorInfo}	"Internal server error"
// @Router       /scheduler/jobs [get]
func (h *schedulerHandler) GetJobs(c echo.Context) error {
	jobStatus := h.schedulerService.GetJobStatus()

	jobsData := model.JobsResponse{
		Jobs:      jobStatus,
		Count:     len(jobStatus),
		Timestamp: time.Now(),
	}

	return response.Success(c, http.StatusOK, jobsData, "Jobs retrieved successfully")
}

// TriggerJob handles POST /api/v1/scheduler/jobs/:name/trigger
// @Summary      Trigger a scheduler job
// @Description  Trigger a specific job by name (acknowledges trigger; job runs according to schedule)
// @Tags         scheduler
// @Accept       json
// @Produce      json
// @Param        name  path      string                   true  "Job name"
// @Success      200   {object}  response.APIResponse{data=model.JobTriggerResponse}  "Trigger acknowledged"
// @Failure      400   {object}  response.APIResponse{error=response.ErrorInfo}		  "Job name required"
// @Failure      404   {object}  response.APIResponse{error=response.ErrorInfo}		  "Job not found"
// @Failure      409   {object}  response.APIResponse{error=response.ErrorInfo}		  "Job already running"
// @Failure      500   {object}  response.APIResponse{error=response.ErrorInfo}		  "Internal server error"
// @Router       /scheduler/jobs/{name}/trigger [post]
func (h *schedulerHandler) TriggerJob(c echo.Context) error {
	start := time.Now()

	jobName := c.Param("name")
	if jobName == "" {
		h.logger.LogServiceOperation("scheduler_handler", "trigger_job", false, time.Since(start).Milliseconds())
		return response.BadRequest(c, "Job name is required")
	}

	jobStatus := h.schedulerService.GetJobStatus()
	job, exists := jobStatus[jobName]
	if !exists {
		h.logger.LogServiceOperation("scheduler_handler", "trigger_job", false, time.Since(start).Milliseconds())
		availableJobs := getJobNames(jobStatus)
		return response.NotFound(c, "Job not found", "Available jobs: "+joinStrings(availableJobs, ", "))
	}

	if job.IsRunning {
		h.logger.LogServiceOperation("scheduler_handler", "trigger_job", false, time.Since(start).Milliseconds())
		return response.Conflict(c, "Job is already running")
	}

	h.logger.Info("Manual job trigger requested via API", "job_name", jobName)

	h.logger.LogServiceOperation("scheduler_handler", "trigger_job", true, time.Since(start).Milliseconds())

	triggerData := model.JobTriggerResponse{
		JobName:   jobName,
		Note:      "Job will run according to its schedule. For immediate execution, use specific aggregation endpoints.",
		NextRun:   job.NextRun,
		Timestamp: time.Now(),
	}

	return response.Success(c, http.StatusOK, triggerData, "Job trigger acknowledged")
}

// getJobNames extracts job names from job status map
func getJobNames(jobStatus map[string]model.JobStatus) []string {
	names := make([]string, 0, len(jobStatus))
	for name := range jobStatus {
		names = append(names, name)
	}
	return names
}

// joinStrings is a helper function to join strings with a separator
func joinStrings(strings []string, separator string) string {
	if len(strings) == 0 {
		return ""
	}

	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += separator + strings[i]
	}
	return result
}
