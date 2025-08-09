package model

import "time"

type Post struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Content     *string    `json:"content,omitempty"`
	URL         string     `json:"url"`
	Source      string     `json:"source"`
	Category    *string    `json:"category,omitempty"`
	ImageURL    *string    `json:"image_url,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreatePostRequest represents the request to create a new post
type CreatePostParams struct {
	Title       string     `json:"title" validate:"required,min=1,max=500"`
	Description *string    `json:"description,omitempty"`
	Content     *string    `json:"content,omitempty"`
	URL         string     `json:"url" validate:"required,min=10,max=500"`
	Source      string     `json:"source" validate:"required,min=1,max=100"`
	Category    *string    `json:"category,omitempty" validate:"omitempty,max=50"`
	ImageURL    *string    `json:"image_url,omitempty" validate:"omitempty,url,max=1000"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

// UpdatePostRequest represents the request to update a post
type UpdatePostParams struct {
	Title       string  `json:"title" validate:"min=1,max=500"`
	Description *string `json:"description,omitempty"`
	Content     *string `json:"content,omitempty"`
	Category    *string `json:"category,omitempty" validate:"omitempty,max=50"`
	ImageURL    *string `json:"image_url,omitempty" validate:"omitempty,url,max=1000"`
}

// BasePostListParams holds common pagination parameters used by post-listing operations.
type BasePostListParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// PostListRequest represents the request parameters for listing posts
type PostListParams struct {
	Page     int     `json:"page" validate:"min=1"`
	Limit    int     `json:"limit" validate:"min=1,max=100"`
	Category *string `json:"category,omitempty"`
	Source   *string `json:"source,omitempty"`
	Search   *string `json:"search,omitempty"`
}

// PostListResponse represents the response for listing posts
type PostListResponse struct {
	Posts      []Post         `json:"posts"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// ListPostsByCategoryParams contains parameters for querying posts filtered by a specific category.
type ListPostsByCategoryParams struct {
	BasePostListParams
	Category string `json:"category"`
}

// ListPostsBySourceParams contains parameters for querying posts filtered by a specific source.
type ListPostsBySourceParams struct {
	BasePostListParams
	Source string `json:"source"`
}

// SearchPostsParams contains parameters for text-based search across posts.
type SearchPostsParams struct {
	BasePostListParams
	Query string `json:"query"`
}

// DefaultPostListParams returns default values for post list request
func DefaultPostListParams() PostListParams {
	return PostListParams{
		Page:  1,
		Limit: 20,
	}
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, limit int, total int64) PaginationMeta {
	totalPages := int(((total + int64(limit)) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	return PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
