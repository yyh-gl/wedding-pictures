package db

import (
	"time"
)

// Image represents an image stored in S3
type Image struct {
	ID          uint      `json:"id"`
	S3Key       string    `json:"s3_key"`
	FileURL     string    `json:"file_url"`
	UploadedAt  time.Time `json:"uploaded_at"`
	IsNew       bool      `json:"is_new"`
	IsDisplayed bool      `json:"is_displayed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
