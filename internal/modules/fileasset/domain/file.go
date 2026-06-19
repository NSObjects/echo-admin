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

// FileObject describes a file uploaded through the back office.
type FileObject struct {
	id          int64
	name        string
	url         string
	size        int64
	contentType string
	createdAt   time.Time
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
	return FileObject{id: id, name: name, url: url, size: size, contentType: contentType, createdAt: createdAt}, nil
}

// ID returns the persisted file id.
func (f FileObject) ID() int64 { return f.id }

// Name returns the original sanitized file name.
func (f FileObject) Name() string { return f.name }

// URL returns the local URL used to retrieve the file.
func (f FileObject) URL() string { return f.url }

// Size returns the uploaded byte size.
func (f FileObject) Size() int64 { return f.size }

// ContentType returns the detected or supplied MIME type.
func (f FileObject) ContentType() string { return f.contentType }

// CreatedAt returns the upload timestamp.
func (f FileObject) CreatedAt() time.Time { return f.createdAt }
