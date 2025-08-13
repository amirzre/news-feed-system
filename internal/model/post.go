package model

import "time"

type Post struct {
	ID          int64      `json:"id" example:"1"`
	Title       string     `json:"title" example:"Breaking: new Go release"`
	Description *string    `json:"description,omitempty" example:"A brief description of the news article"`
	Content     *string    `json:"content,omitempty" example:"Full content of the article..."`
	URL         string     `json:"url" example:"https://example.com/article"`
	Source      string     `json:"source" example:"TechCrunch"`
	Category    *string    `json:"category,omitempty" example:"technology"`
	ImageURL    *string    `json:"image_url,omitempty" example:"https://example.com/image.jpg"`
	PublishedAt *time.Time `json:"published_at,omitempty" swaggertype:"string" example:"2024-01-20T10:00:00Z"`
	CreatedAt   time.Time  `json:"created_at" swaggertype:"string" example:"2025-08-11T07:11:03Z"`
	UpdatedAt   time.Time  `json:"updated_at" swaggertype:"string" example:"2025-08-11T07:16:04Z"`
}

// CreatePostRequest represents the request to create a new post
type CreatePostParams struct {
	Title       string     `json:"title" validate:"required,min=1,max=500" example:"Breaking: new Go release"`
	Description *string    `json:"description,omitempty" example:"A brief description"`
	Content     *string    `json:"content,omitempty" example:"Full content..."`
	URL         string     `json:"url" validate:"required,min=10,max=500" example:"https://example.com/article"`
	Source      string     `json:"source" validate:"required,min=1,max=100" example:"TechCrunch"`
	Category    *string    `json:"category,omitempty" validate:"omitempty,max=50" example:"technology"`
	ImageURL    *string    `json:"image_url,omitempty" validate:"omitempty,url,max=1000" example:"https://example.com/image.jpg"`
	PublishedAt *time.Time `json:"published_at,omitempty" swaggertype:"string" example:"2024-01-20T10:00:00Z"`
}

// UpdatePostRequest represents the request to update a post
type UpdatePostParams struct {
	Title       string  `json:"title" validate:"min=1,max=500" example:"Updated title"`
	Description *string `json:"description,omitempty" example:"Updated description"`
	Content     *string `json:"content,omitempty" example:"Updated content"`
	Category    *string `json:"category,omitempty" validate:"omitempty,max=50" example:"business"`
	ImageURL    *string `json:"image_url,omitempty" validate:"omitempty,url,max=1000" example:"https://example.com/updated.jpg"`
}

// BasePostListParams holds common pagination parameters used by post-listing operations.
type BasePostListParams struct {
	Limit  int `json:"limit" example:"10"`
	Offset int `json:"offset" example:"0"`
}

// PostListRequest represents the request parameters for listing posts
type PostListParams struct {
	Page     int     `json:"page" validate:"min=1" example:"1"`
	Limit    int     `json:"limit" validate:"min=1,max=100" example:"10"`
	Category *string `json:"category,omitempty" example:"technology"`
	Source   *string `json:"source,omitempty" example:"TechCrunch"`
	Search   *string `json:"search,omitempty" example:"openai"`
}

// PostListResponse represents the response for listing posts
type PostListResponse struct {
	Posts      []Post         `json:"posts"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page" example:"1"`
	Limit      int   `json:"limit" example:"10"`
	Total      int64 `json:"total" example:"123"`
	TotalPages int   `json:"total_pages" example:"13"`
	HasNext    bool  `json:"has_next" example:"true"`
	HasPrev    bool  `json:"has_prev" example:"false"`
}

// ListPostsByCategoryParams contains parameters for querying posts filtered by a specific category.
type ListPostsByCategoryParams struct {
	BasePostListParams
	Category string `json:"category" example:"technology"`
}

// ListPostsBySourceParams contains parameters for querying posts filtered by a specific source.
type ListPostsBySourceParams struct {
	BasePostListParams
	Source string `json:"source" example:"TechCrunch"`
}

// SearchPostsParams contains parameters for text-based search across posts.
type SearchPostsParams struct {
	BasePostListParams
	Query string `json:"query" example:"openai"`
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
