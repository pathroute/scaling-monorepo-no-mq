// Package api defines the files service's wire types.
package api

import "time"

// File is an attachment's metadata.
type File struct {
	Name       string    `json:"name"`
	Org        string    `json:"org"`
	Size       int64     `json:"size"`
	SHA256     string    `json:"sha256"`
	UploadedAt time.Time `json:"uploaded_at"`
}
