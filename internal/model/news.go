package model

import "time"

type NewsParams struct {
	Query    string   `json:"q,omitempty" example:"openai"`
	Sources  []string `json:"sources,omitempty" example:"[\"techcrunch\"]"`
	Category string   `json:"category,omitempty" example:"technology"`
	Country  string   `json:"country,omitempty" example:"us"`
	Language string   `json:"language,omitempty" example:"en"`
	PageSize int      `json:"pageSize,omitempty" example:"20"`
	Page     int      `json:"page,omitempty" example:"1"`
}

// NewsAPIArticleParams represents an article from News API
type NewsAPIArticleParams struct {
	Source struct {
		ID   *string `json:"id" example:"techcrunch"`
		Name string  `json:"name" example:"TechCrunch"`
	} `json:"source"`
	Author      *string `json:"author,omitempty" example:"John Doe"`
	Title       string  `json:"title" example:"New breakthrough in AI"`
	Description *string `json:"description,omitempty" example:"Short article description"`
	URL         string  `json:"url" example:"https://example.com/article"`
	URLToImage  *string `json:"urlToImage,omitempty" example:"https://example.com/image.jpg"`
	PublishedAt string  `json:"publishedAt" swaggertype:"string" example:"2024-01-20T10:00:00Z"`
	Content     *string `json:"content,omitempty" example:"Full article content..."`
}

// NewsAPIResponse represents the response from News API
type NewsAPIResponse struct {
	Status       string                 `json:"status" example:"ok"`
	TotalResults int                    `json:"totalResults" example:"100"`
	Articles     []NewsAPIArticleParams `json:"articles"`
}

// ToPost converts NewsAPIArticle to Post model
func (article *NewsAPIArticleParams) ToPost() (*CreatePostParams, error) {
	publishedAt, err := time.Parse(time.RFC3339, article.PublishedAt)
	if err != nil {
		publishedAt = time.Time{}
	}

	post := &CreatePostParams{
		Title:       article.Title,
		Description: article.Description,
		Content:     article.Content,
		URL:         article.URL,
		Source:      article.Source.Name,
		ImageURL:    article.URLToImage,
	}

	if !publishedAt.IsZero() {
		post.PublishedAt = &publishedAt
	}

	return post, nil
}
