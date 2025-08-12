package model

import "time"

type JobStatus struct {
	Name           string        `json:"name"`
	Interval       time.Duration `json:"interval"`
	LastRun        *time.Time    `json:"last_run,omitempty"`
	NextRun        *time.Time    `json:"next_run,omitempty"`
	RunCount       int64         `json:"run_count"`
	ErrorCount     int64         `json:"error_count"`
	LastError      string        `json:"last_error,omitempty"`
	IsRunning      bool          `json:"is_running"`
	AverageRunTime time.Duration `json:"average_run_time"`
}

// SchedulerStatusResponse represents the scheduler status response
type SchedulerStatusResponse struct {
	SchedulerRunning bool                 `json:"scheduler_running"`
	JobsCount        int                  `json:"jobs_count"`
	Timestamp        time.Time            `json:"timestamp"`
	Jobs             map[string]JobStatus `json:"jobs"`
}

// JobsResponse represents the jobs listing response
type JobsResponse struct {
	Jobs      map[string]JobStatus `json:"jobs"`
	Count     int                  `json:"count"`
	Timestamp time.Time            `json:"timestamp"`
}
