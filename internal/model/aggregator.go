package model

import "time"

type AggregationResponse struct {
	TotalFetched    int                      `json:"total_fetched" example:"150"`
	TotalCreated    int                      `json:"total_created" example:"120"`
	TotalDuplicates int                      `json:"total_duplicates" example:"25"`
	TotalErrors     int                      `json:"total_errors" example:"5"`
	Duration        time.Duration            `json:"duration" swaggertype:"string" example:"1s"`
	Categories      map[string]CategoryStats `json:"categories,omitempty"`
	Sources         map[string]SourceStats   `json:"sources,omitempty"`
	Errors          []string                 `json:"errors,omitempty" example:"[]"`
}

type BaseStats struct {
	Fetched    int `json:"fetched" example:"100"`
	Created    int `json:"created" example:"80"`
	Duplicates int `json:"duplicates" example:"15"`
	Errors     int `json:"errors" example:"2"`
}

type CategoryStats struct {
	BaseStats
}

type SourceStats struct {
	BaseStats
}

// CategoryAggregationRequest represents the request body for category aggregation
type CategoryAggregationRequest struct {
	Categories []string `json:"categories,omitempty" example:"[\"technology\",\"business\"]"`
}

// CategoryAggregationResponse represents the response for category aggregation
type CategoryAggregationResponse struct {
	Categories []string            `json:"categories" example:"[\"technology\"]"`
	Result     AggregationResponse `json:"result"`
}

// SourceAggregationRequest represents the request body for source aggregation
type SourceAggregationRequest struct {
	Sources []string `json:"sources,omitempty" example:"[\"techcrunch\",\"wired\"]"`
}

// SourceAggregationResponse represents the response for source aggregation
type SourceAggregationResponse struct {
	Sources []string            `json:"sources" example:"[\"techcrunch\"]"`
	Result  AggregationResponse `json:"result"`
}
