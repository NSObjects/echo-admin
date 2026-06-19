// Package domain contains uploaded file metadata rules.
package domain

import (
	"errors"
	"strings"
	"time"
)

// File validation errors.
var (
	ErrInvalidFileID      = errors.New("invalid file id")
	ErrInvalidFileName    = errors.New("invalid file name")
	ErrInvalidFileURL     = errors.New("invalid file url")
	ErrInvalidFileSize    = errors.New("invalid file size")
	ErrInvalidContentType = errors.New("invalid content type")
)

// FileObject is validated metadata for a file uploaded through the back office.
type FileObject struct {
	ID          int64
	Name        string
	URL         string
	Size        int64
	ContentType string
	CreatedAt   time.Time
}

// RestoreFileObject rebuilds an uploaded file record from a trusted store representation.
func RestoreFileObject(id int64, name, url string, size int64, contentType string, createdAt time.Time) (FileObject, error) {
	name = strings.TrimSpace(name)
	url = strings.TrimSpace(url)
	contentType = strings.TrimSpace(contentType)
	if id < 0 {
		return FileObject{}, ErrInvalidFileID
	}
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "\\") || len(name) > 180 {
		return FileObject{}, ErrInvalidFileName
	}
	if url == "" || !strings.HasPrefix(url, "/") || strings.HasPrefix(url, "//") {
		return FileObject{}, ErrInvalidFileURL
	}
	if size <= 0 {
		return FileObject{}, ErrInvalidFileSize
	}
	if contentType == "" || len(contentType) > 120 {
		return FileObject{}, ErrInvalidContentType
	}
	return FileObject{ID: id, Name: name, URL: url, Size: size, ContentType: contentType, CreatedAt: createdAt}, nil
}
