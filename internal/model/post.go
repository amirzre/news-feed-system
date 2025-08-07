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
