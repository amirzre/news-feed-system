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

// CategoryAggregationRequest represents the request body for category aggregation
type CategoryAggregationRequest struct {
	Categories []string `json:"categories,omitempty"`
}

// CategoryAggregationResponse represents the response for category aggregation
type CategoryAggregationResponse struct {
	Categories []string `json:"categories"`
	Result     any      `json:"result"`
}

// SourceAggregationRequest represents the request body for source aggregation
type SourceAggregationRequest struct {
	Sources []string `json:"sources,omitempty"`
}

// SourceAggregationResponse represents the response for source aggregation
type SourceAggregationResponse struct {
	Sources []string `json:"sources"`
	Result  any      `json:"result"`
}
