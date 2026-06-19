// Package usecase coordinates uploaded file metadata workflows.
package usecase

import (
	"context"
	"time"

	"github.com/NSObjects/echo-admin/internal/modules/fileasset/domain"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// Store persists uploaded file metadata.
type Store interface {
	CreateFile(context.Context, domain.FileObject) (domain.FileObject, error)
	ListFiles(context.Context, ListFilter) ([]domain.FileObject, int, error)
}

// Usecase coordinates uploaded file metadata rules.
type Usecase struct {
	store Store
}

// New creates a file asset usecase.
func New(store Store) *Usecase {
	return &Usecase{store: store}
}

// FileInput carries an uploaded file record after the HTTP adapter stores bytes.
type FileInput struct {
	Name        string
	URL         string
	Size        int64
	ContentType string
}

// ListInput carries pagination for file lists.
type ListInput struct {
	Page     int
	PageSize int
}

// ListFilter is the validated store-facing pagination window.
type ListFilter struct {
	Offset   int
	Limit    int
	Page     int
	PageSize int
}

// ListOutput is a paginated uploaded file result.
type ListOutput struct {
	Items    []FileObject
	Page     int
	PageSize int
	Total    int
}

// FileObject is the adapter-facing uploaded file DTO.
type FileObject struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}
