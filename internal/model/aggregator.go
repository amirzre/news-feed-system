package model

import (
	"time"
)

type AggregationResponse struct {
	TotalFetched    int                      `json:"total_fetched"`
	TotalCreated    int                      `json:"total_created"`
	TotalDuplicates int                      `json:"total_duplicates"`
	TotalErrors     int                      `json:"total_errors"`
	Duration        time.Duration            `json:"duration"`
	Categories      map[string]CategoryStats `json:"categories,omitempty"`
	Sources         map[string]SourceStats   `json:"sources,omitempty"`
	Errors          []string                 `json:"errors,omitempty"`
}

type BaseStats struct {
	Fetched    int `json:"fetched"`
	Created    int `json:"created"`
	Duplicates int `json:"duplicates"`
	Errors     int `json:"errors"`
}

type CategoryStats struct {
	BaseStats
}

type SourceStats struct {
	BaseStats
}
