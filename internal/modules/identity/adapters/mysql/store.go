// Package mysql persists administrators in MySQL.
package mysql

import (
	"context"
	"errors"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	"github.com/NSObjects/echo-admin/internal/modules/identity/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/infrastructure/mysqljson"
)

const defaultAdminPassword = "admin123"

// Store persists administrators in MySQL.
type Store struct {
	db *gorm.DB
}

// NewStore migrates the MySQL administrator table.
func NewStore(ctx context.Context, db *gorm.DB) (*Store, error) {
	if ctx == nil {
		return nil, errors.New("create identity store: nil context")
	}
	if db == nil {
		return nil, errors.New("create identity store: nil db")
	}
	if err := db.WithContext(ctx).AutoMigrate(&adminModel{}); err != nil {
		return nil, apperr.WrapDatabase(err, "migrate identity tables")
	}
	return &Store{db: db}, nil
}

// SeedDefaultAdmin creates the first administrator when none exists.
func (s *Store) SeedDefaultAdmin(ctx context.Context, roleID int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var existing adminModel
	err := s.db.WithContext(ctx).Where("username = ?", "admin").First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.WrapDatabase(err, "find seed admin")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	admin, err := domain.RestoreAdmin(0, "admin", "系统管理员", "admin@example.com", hash, []int64{roleID}, true, now, now)
	if err != nil {
		return err
	}
	model := adminModelFromDomain(admin)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return apperr.WrapDatabase(err, "create seed admin")
	}
	return nil
}

// FindByUsername returns an admin by normalized username.
func (s *Store) FindByUsername(ctx context.Context, username string) (domain.Admin, error) {
	if err := ctx.Err(); err != nil {
		return domain.Admin{}, err
	}
	var model adminModel
	err := s.db.WithContext(ctx).First(&model, "username = ?", username).Error
	if err != nil {
		return domain.Admin{}, mapReadError(err, "admin", "find admin by username")
	}
	return model.toDomain()
}

// FindByID returns an admin by id.
func (s *Store) FindByID(ctx context.Context, id int64) (domain.Admin, error) {
	if err := ctx.Err(); err != nil {
		return domain.Admin{}, err
	}
	var model adminModel
	err := s.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		return domain.Admin{}, mapReadError(err, "admin", "find admin")
	}
	return model.toDomain()
}

// List returns admins ordered by id descending.
func (s *Store) List(ctx context.Context, filter usecase.ListFilter) ([]domain.Admin, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	var total int64
	query := s.db.WithContext(ctx).Model(&adminModel{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperr.WrapDatabase(err, "count admins")
	}
	var models []adminModel
	err := query.Order("id DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&models).Error
	if err != nil {
		return nil, 0, apperr.WrapDatabase(err, "list admins")
	}
	admins := make([]domain.Admin, 0, len(models))
	for _, model := range models {
		admin, err := model.toDomain()
		if err != nil {
			return nil, 0, err
		}
		admins = append(admins, admin)
	}
	return admins, int(total), nil
}

// Create inserts an admin.
func (s *Store) Create(ctx context.Context, admin domain.Admin) (domain.Admin, error) {
	if err := ctx.Err(); err != nil {
		return domain.Admin{}, err
	}
	model := adminModelFromDomain(admin)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Admin{}, mapWriteError(err, "admin username already exists", "create admin")
	}
	return model.toDomain()
}

// Update replaces mutable admin fields.
func (s *Store) Update(ctx context.Context, admin domain.Admin) (domain.Admin, error) {
	if err := ctx.Err(); err != nil {
		return domain.Admin{}, err
	}
	model := adminModelFromDomain(admin)
	result := s.db.WithContext(ctx).Save(&model)
	if result.Error != nil {
		return domain.Admin{}, mapWriteError(result.Error, "admin username already exists", "update admin")
	}
	if result.RowsAffected == 0 {
		return domain.Admin{}, apperr.NewNotFound("admin")
	}
	return model.toDomain()
}

type adminModel struct {
	ID           int64            `gorm:"primaryKey"`
	Username     string           `gorm:"type:varchar(64);not null;uniqueIndex"`
	DisplayName  string           `gorm:"type:varchar(80);not null"`
	Email        string           `gorm:"type:varchar(160);not null"`
	PasswordHash string           `gorm:"type:varchar(100);not null"`
	RoleIDs      mysqljson.Int64s `gorm:"type:json;not null"`
	Active       bool             `gorm:"not null"`
	CreatedAt    time.Time        `gorm:"not null"`
	UpdatedAt    time.Time        `gorm:"not null"`
}

func (adminModel) TableName() string {
	return "identity_admins"
}

func adminModelFromDomain(admin domain.Admin) adminModel {
	return adminModel{
		ID:           admin.ID(),
		Username:     admin.Username(),
		DisplayName:  admin.DisplayName(),
		Email:        admin.Email(),
		PasswordHash: string(admin.PasswordHash()),
		RoleIDs:      mysqljson.Int64s(admin.RoleIDs()),
		Active:       admin.Active(),
		CreatedAt:    admin.CreatedAt(),
		UpdatedAt:    admin.UpdatedAt(),
	}
}

func (m adminModel) toDomain() (domain.Admin, error) {
	return domain.RestoreAdmin(m.ID, m.Username, m.DisplayName, m.Email, []byte(m.PasswordHash), []int64(m.RoleIDs), m.Active, m.CreatedAt, m.UpdatedAt)
}

func mapReadError(err error, resource, operation string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.NewNotFound(resource)
	}
	return apperr.WrapDatabase(err, operation)
}

func mapWriteError(err error, conflictMessage, operation string) error {
	var mysqlErr *drivermysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return apperr.NewConflict(conflictMessage)
	}
	return apperr.WrapDatabase(err, operation)
}
