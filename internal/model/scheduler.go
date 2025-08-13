package model

import "time"

type JobStatus struct {
	Name           string        `json:"name" example:"aggregate_all"`
	Interval       time.Duration `json:"interval" swaggertype:"string" example:"1h"`
	LastRun        *time.Time    `json:"last_run,omitempty" swaggertype:"string" example:"2025-08-11T07:11:03Z"`
	NextRun        *time.Time    `json:"next_run,omitempty" swaggertype:"string" example:"2025-08-11T08:11:03Z"`
	RunCount       int64         `json:"run_count" example:"42"`
	ErrorCount     int64         `json:"error_count" example:"1"`
	LastError      string        `json:"last_error,omitempty" example:"timeout error"`
	IsRunning      bool          `json:"is_running" example:"false"`
	AverageRunTime time.Duration `json:"average_run_time" swaggertype:"string" example:"30s"`
}

// SchedulerStatusResponse represents the scheduler status response
type SchedulerStatusResponse struct {
	SchedulerRunning bool                 `json:"scheduler_running" example:"true"`
	JobsCount        int                  `json:"jobs_count" example:"3"`
	Timestamp        time.Time            `json:"timestamp" swaggertype:"string" example:"2025-08-11T07:11:03Z"`
	Jobs             map[string]JobStatus `json:"jobs"`
}

// JobsResponse represents the jobs listing response
type JobsResponse struct {
	Jobs      map[string]JobStatus `json:"jobs"`
	Count     int                  `json:"count" example:"3"`
	Timestamp time.Time            `json:"timestamp" swaggertype:"string" example:"2025-08-11T07:11:03Z"`
}

// JobTriggerResponse represents the job trigger response
type JobTriggerResponse struct {
	JobName   string     `json:"job_name" example:"aggregate_all"`
	Note      string     `json:"note" example:"Job will run according to its schedule."`
	NextRun   *time.Time `json:"next_run,omitempty" swaggertype:"string" example:"2025-08-11T08:11:03Z"`
	Timestamp time.Time  `json:"timestamp" swaggertype:"string" example:"2025-08-11T07:11:03Z"`
}
