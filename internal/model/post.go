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
