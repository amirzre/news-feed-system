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
