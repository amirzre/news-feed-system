package model

// NewsAPIArticleParams represents an article from News API
type NewsAPIArticleParams struct {
	Source struct {
		ID   *string `json:"id"`
		Name string  `json:"name"`
	} `json:"source"`
	Author      *string `json:"author"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	URL         string  `json:"url"`
	URLToImage  *string `json:"urlToImage"`
	PublishedAt string  `json:"publishedAt"`
	Content     *string `json:"content"`
}

// NewsAPIResponse represents the response from News API
type NewsAPIResponse struct {
	Status       string                 `json:"status"`
	TotalResults int                    `json:"totalResults"`
	Articles     []NewsAPIArticleParams `json:"articles"`
}
