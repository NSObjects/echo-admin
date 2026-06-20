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

// Store persists uploaded file metadata and its operator-managed categories.
type Store interface {
	CreateFile(context.Context, domain.FileObject) (domain.FileObject, error)
	FindFileByID(context.Context, int64) (domain.FileObject, error)
	ListFiles(context.Context, ListFilter) ([]domain.FileObject, int, error)
	UpdateFile(context.Context, domain.FileObject) (domain.FileObject, error)
	DeleteFile(context.Context, int64) error
	CreateCategory(context.Context, domain.FileCategory) (domain.FileCategory, error)
	UpdateCategory(context.Context, domain.FileCategory) (domain.FileCategory, error)
	FindCategoryByID(context.Context, int64) (domain.FileCategory, error)
	ListCategories(context.Context) ([]domain.FileCategory, error)
	DeleteCategory(context.Context, int64) error
	CategoryNameExists(context.Context, string, int64, int64) (bool, error)
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
	CategoryID  int64
}

// URLImportInput carries an external HTTP(S) URL to register as a file asset.
type URLImportInput struct {
	Name       string
	URL        string
	CategoryID int64
}

// RenameInput carries a display-name update for one file.
type RenameInput struct {
	ID   int64
	Name string
}

// ListInput carries pagination for file lists.
type ListInput struct {
	Page       int
	PageSize   int
	CategoryID int64
}

// ListFilter is the validated store-facing pagination window.
type ListFilter struct {
	Offset     int
	Limit      int
	Page       int
	PageSize   int
	CategoryID int64
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
	CategoryID  int64     `json:"category_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// CategoryInput carries category fields for creation.
type CategoryInput struct {
	Name     string
	ParentID int64
}

// UpdateCategoryInput carries category fields for updates.
type UpdateCategoryInput struct {
	ID       int64
	Name     string
	ParentID int64
}

// Category is a file category tree node returned to HTTP adapters.
type Category struct {
	ID        int64      `json:"id"`
	ParentID  int64      `json:"parent_id"`
	Name      string     `json:"name"`
	Children  []Category `json:"children"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
