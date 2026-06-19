// Package mysql persists uploaded file metadata in MySQL.
package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/NSObjects/echo-admin/internal/modules/fileasset/domain"
	"github.com/NSObjects/echo-admin/internal/modules/fileasset/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
)

// Store persists uploaded file metadata in MySQL.
type Store struct {
	db *gorm.DB
}

// NewStore migrates the MySQL file metadata table.
func NewStore(ctx context.Context, db *gorm.DB) (*Store, error) {
	if ctx == nil {
		return nil, errors.New("create file store: nil context")
	}
	if db == nil {
		return nil, errors.New("create file store: nil db")
	}
	if err := db.WithContext(ctx).AutoMigrate(&fileModel{}); err != nil {
		return nil, apperr.WrapDatabase(err, "migrate file tables")
	}
	return &Store{db: db}, nil
}

// CreateFile inserts uploaded file metadata.
func (s *Store) CreateFile(ctx context.Context, file domain.FileObject) (domain.FileObject, error) {
	if err := ctx.Err(); err != nil {
		return domain.FileObject{}, err
	}
	model := fileModelFromDomain(file, time.Now().UTC())
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.FileObject{}, apperr.WrapDatabase(err, "create file")
	}
	return model.toDomain()
}

// ListFiles returns uploaded file metadata ordered by id descending.
func (s *Store) ListFiles(ctx context.Context, filter usecase.ListFilter) ([]domain.FileObject, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	var total int64
	query := s.db.WithContext(ctx).Model(&fileModel{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperr.WrapDatabase(err, "count files")
	}
	var models []fileModel
	err := query.Order("id DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&models).Error
	if err != nil {
		return nil, 0, apperr.WrapDatabase(err, "list files")
	}
	files := make([]domain.FileObject, 0, len(models))
	for _, model := range models {
		file, err := model.toDomain()
		if err != nil {
			return nil, 0, err
		}
		files = append(files, file)
	}
	return files, int(total), nil
}

type fileModel struct {
	ID          int64     `gorm:"primaryKey"`
	Name        string    `gorm:"type:varchar(180);not null"`
	URL         string    `gorm:"type:varchar(255);not null"`
	Size        int64     `gorm:"not null"`
	ContentType string    `gorm:"type:varchar(120);not null"`
	CreatedAt   time.Time `gorm:"not null"`
}

func (fileModel) TableName() string {
	return "file_assets"
}

func fileModelFromDomain(file domain.FileObject, createdAt time.Time) fileModel {
	return fileModel{
		ID:          file.ID,
		Name:        file.Name,
		URL:         file.URL,
		Size:        file.Size,
		ContentType: file.ContentType,
		CreatedAt:   coalesceTime(file.CreatedAt, createdAt),
	}
}

func (m fileModel) toDomain() (domain.FileObject, error) {
	return domain.RestoreFileObject(m.ID, m.Name, m.URL, m.Size, m.ContentType, m.CreatedAt)
}

func coalesceTime(value, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}
