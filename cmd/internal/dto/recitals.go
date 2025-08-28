package dto

import "time"

type Recital struct {
	ID          int32     `json:"id"`
	Url         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
}
