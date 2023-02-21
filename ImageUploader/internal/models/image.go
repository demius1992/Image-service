package models

import (
	"github.com/google/uuid"
)

type Image struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   string    `json:"created_at"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	Content     []byte    `json:"-"`
}
