package models

import (
	"github.com/google/uuid"
)

type Image struct {
	ID           uuid.UUID    `json:"id"`
	Original     ImageVariant `json:"original"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	ImageURL     string       `json:"image_url"`
	ThumbnailURL string       `json:"thumbnail_url"`
	CreatedAt    string       `json:"created_at"`
}

type ImageVariant struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Content     []byte `json:"-"`
}
