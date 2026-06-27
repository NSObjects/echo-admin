// Package mysql persists administrators in MySQL.
package mysql

import (
	"context"
	"errors"
	"strings"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/infrastructure/mysqljson"
)

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

// SeedDefaultAdmin creates or repairs the local bootstrap administrator. The
// bootstrap password is consumed only when the admin row does not exist; an
// existing operator-managed password is never reset by startup repair.
func (s *Store) SeedDefaultAdmin(ctx context.Context, roleID int64, bootstrapPassword string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var existing adminModel
	err := s.db.WithContext(ctx).Where("username = ?", "admin").First(&existing).Error
	if err == nil {
		if existing.ActiveRoleID == 0 || !containsRoleID([]int64(existing.RoleIDs), existing.ActiveRoleID) {
			existing.ActiveRoleID = roleID
		}
		if !containsRoleID([]int64(existing.RoleIDs), roleID) {
			existing.RoleIDs = append(existing.RoleIDs, roleID)
		}
		existing.Active = true
		if saveErr := s.db.WithContext(ctx).Save(&existing).Error; saveErr != nil {
			return apperr.WrapDatabase(saveErr, "update seed admin")
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.WrapDatabase(err, "find seed admin")
	}
	password, err := normalizeBootstrapAdminPassword(bootstrapPassword)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	admin, err := domain.RestoreAdmin(0, "admin", "系统管理员", "admin@example.com", hash, []int64{roleID}, roleID, true, now, now)
	if err != nil {
		return err
	}
	model := adminModelFromDomain(admin)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return apperr.WrapDatabase(err, "create seed admin")
	}
	return nil
}

func normalizeBootstrapAdminPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", errors.New("admin bootstrap_password is required for first bootstrap")
	}
	if password == "123456" {
		return "", errors.New("admin bootstrap_password must not use the removed default password")
	}
	if len(password) < 8 || len(password) > 72 {
		return "", errors.New("admin bootstrap_password must be 8 to 72 characters")
	}
	return password, nil
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

// ListAll returns admins ordered by id descending.
func (s *Store) ListAll(ctx context.Context) ([]domain.Admin, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var models []adminModel
	err := s.db.WithContext(ctx).Order("id DESC").Find(&models).Error
	if err != nil {
		return nil, apperr.WrapDatabase(err, "list all admins")
	}
	admins := make([]domain.Admin, 0, len(models))
	for _, model := range models {
		admin, err := model.toDomain()
		if err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}
	return admins, nil
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

// Delete removes an administrator row by id.
func (s *Store) Delete(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	result := s.db.WithContext(ctx).Delete(&adminModel{}, "id = ?", id)
	if result.Error != nil {
		return apperr.WrapDatabase(result.Error, "delete admin")
	}
	if result.RowsAffected == 0 {
		return apperr.NewNotFound("admin")
	}
	return nil
}

// RoleAssigned reports whether any administrator still carries the role.
func (s *Store) RoleAssigned(ctx context.Context, roleID int64) (bool, error) {
	admins, err := s.ListAll(ctx)
	if err != nil {
		return false, err
	}
	for _, admin := range admins {
		if containsRoleID(admin.RoleIDs, roleID) {
			return true, nil
		}
	}
	return false, nil
}

type adminModel struct {
	ID           int64            `gorm:"primaryKey"`
	Username     string           `gorm:"type:varchar(64);not null;uniqueIndex"`
	DisplayName  string           `gorm:"type:varchar(80);not null"`
	Email        string           `gorm:"type:varchar(160);not null"`
	PasswordHash string           `gorm:"type:varchar(100);not null"`
	RoleIDs      mysqljson.Int64s `gorm:"type:json;not null"`
	ActiveRoleID int64            `gorm:"not null;index"`
	Active       bool             `gorm:"not null"`
	CreatedAt    time.Time        `gorm:"not null"`
	UpdatedAt    time.Time        `gorm:"not null"`
}

func (adminModel) TableName() string {
	return "identity_admins"
}

func adminModelFromDomain(admin domain.Admin) adminModel {
	return adminModel{
		ID:           admin.ID,
		Username:     admin.Username,
		DisplayName:  admin.DisplayName,
		Email:        admin.Email,
		PasswordHash: string(admin.PasswordHash),
		RoleIDs:      mysqljson.Int64s(admin.RoleIDs),
		ActiveRoleID: admin.ActiveRoleID,
		Active:       admin.Active,
		CreatedAt:    admin.CreatedAt,
		UpdatedAt:    admin.UpdatedAt,
	}
}

func (m adminModel) toDomain() (domain.Admin, error) {
	return domain.RestoreAdmin(m.ID, m.Username, m.DisplayName, m.Email, []byte(m.PasswordHash), []int64(m.RoleIDs), m.ActiveRoleID, m.Active, m.CreatedAt, m.UpdatedAt)
}

func containsRoleID(roleIDs []int64, want int64) bool {
	for _, roleID := range roleIDs {
		if roleID == want {
			return true
		}
	}
	return false
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
