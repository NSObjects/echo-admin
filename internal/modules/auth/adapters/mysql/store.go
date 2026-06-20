// Package mysql persists authentication runtime state in MySQL.
package mysql

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/NSObjects/echo-admin/internal/modules/auth/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
)

// Store persists auth-owned runtime data such as revoked JWTs.
type Store struct {
	db *gorm.DB
}

// NewStore migrates the auth runtime tables.
func NewStore(ctx context.Context, db *gorm.DB) (*Store, error) {
	if ctx == nil {
		return nil, errors.New("create auth store: nil context")
	}
	if db == nil {
		return nil, errors.New("create auth store: nil db")
	}
	if err := db.WithContext(ctx).AutoMigrate(&jwtBlacklistModel{}); err != nil {
		return nil, apperr.WrapDatabase(err, "migrate auth tables")
	}
	return &Store{db: db}, nil
}

// AddJWTBlacklist records a revoked JWT hash. The operation is idempotent so
// repeated logout requests keep the revocation active until the latest expiry.
func (s *Store) AddJWTBlacklist(ctx context.Context, entry usecase.JWTBlacklistEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	hash := strings.TrimSpace(entry.TokenHash)
	if hash == "" {
		return apperr.NewBadRequest("invalid jwt token")
	}
	now := time.Now().UTC()
	model := jwtBlacklistModel{
		TokenHash: hash,
		ExpiresAt: entry.ExpiresAt.UTC(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "token_hash"}},
			DoUpdates: clause.AssignmentColumns([]string{"expires_at", "updated_at"}),
		}).
		Create(&model).Error
	if err != nil {
		return apperr.WrapDatabase(err, "record jwt blacklist")
	}
	return nil
}

// JWTBlacklisted reports whether a JWT hash is still revoked at the supplied time.
func (s *Store) JWTBlacklisted(ctx context.Context, tokenHash string, now time.Time) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	tokenHash = strings.TrimSpace(tokenHash)
	if tokenHash == "" {
		return false, nil
	}
	var count int64
	err := s.db.WithContext(ctx).
		Model(&jwtBlacklistModel{}).
		Where("token_hash = ? AND expires_at > ?", tokenHash, now.UTC()).
		Count(&count).Error
	if err != nil {
		return false, apperr.WrapDatabase(err, "check jwt blacklist")
	}
	return count > 0, nil
}

type jwtBlacklistModel struct {
	ID        int64     `gorm:"primaryKey"`
	TokenHash string    `gorm:"type:char(64);not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (jwtBlacklistModel) TableName() string {
	return "auth_jwt_blacklists"
}
