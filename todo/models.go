package main

import "time"

// Todo types (existing)
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Todo struct {
	ID        int64     `json:"id"`
	Task      string    `json:"task"`
	Done      bool      `json:"done"`
	Priority  Priority  `json:"priority"`
	Category  string    `json:"category"`
	DueDate   string    `json:"due_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TodoFilter struct {
	Status   string // all, done, pending
	Priority string // all, low, medium, high
	Category string
	Search   string
}

// Vault types (new)
type ContentType string

const (
	ContentTypeTweet   ContentType = "tweet"
	ContentTypeTikTok  ContentType = "tiktok"
	ContentTypeYouTube ContentType = "youtube"
	ContentTypeArticle ContentType = "article"
	ContentTypeNote    ContentType = "note"
)

type VaultItem struct {
	ID              int64       `json:"id"`
	ContentType     ContentType `json:"content_type"`
	Title           string      `json:"title"`
	Content         string      `json:"content"`
	URL             string      `json:"url"`
	MetaTitle       string      `json:"meta_title"`
	MetaDescription string      `json:"meta_description"`
	MetaThumbnail   string      `json:"meta_thumbnail"`
	MetaAuthor      string      `json:"meta_author"`
	MetaSiteName    string      `json:"meta_site_name"`
	Pinned          bool        `json:"pinned"`
	Archived        bool        `json:"archived"`
	Tags            []Tag       `json:"tags"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

type VaultFilter struct {
	ContentType string
	TagNames    []string
	Pinned      *bool
	Archived    *bool
	Search      string
	Limit       int
	Offset      int
}
