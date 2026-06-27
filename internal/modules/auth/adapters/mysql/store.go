// Package mysql persists authentication runtime state in MySQL.
package mysql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/NSObjects/echo-admin/internal/modules/auth/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
)

const (
	loginAttemptWindow      = 15 * time.Minute
	loginAttemptLockout     = 15 * time.Minute
	maxFailedLoginAttempts  = 5
	loginAttemptKeyMaxBytes = 128
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
	if err := db.WithContext(ctx).AutoMigrate(&jwtBlacklistModel{}, &loginAttemptModel{}); err != nil {
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

// CheckLoginAttempt rejects a login when the hashed username/client key is
// still locked by previous failures.
func (s *Store) CheckLoginAttempt(ctx context.Context, key string, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	key, err := normalizeLoginAttemptKey(key)
	if err != nil {
		return err
	}
	attempt, found, err := s.findLoginAttempt(ctx, key)
	if err != nil || !found {
		return err
	}
	now = now.UTC()
	if attempt.LockedUntil != nil && attempt.LockedUntil.After(now) {
		return apperr.New(apperr.ErrTooManyAttempts, "too many login attempts")
	}
	if loginAttemptExpired(attempt, now) {
		return s.ResetLoginAttempts(ctx, key)
	}
	return nil
}

// RecordLoginFailure increments the failure counter for one hashed login key.
func (s *Store) RecordLoginFailure(ctx context.Context, key string, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	key, err := normalizeLoginAttemptKey(key)
	if err != nil {
		return err
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.recordLoginFailure(ctx, tx, key, now.UTC())
	})
	if err != nil {
		return apperr.WrapDatabase(err, "record login attempt")
	}
	return nil
}

// ResetLoginAttempts clears failed-attempt state after a successful login.
func (s *Store) ResetLoginAttempts(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	key, err := normalizeLoginAttemptKey(key)
	if err != nil {
		return err
	}
	err = s.db.WithContext(ctx).Where("attempt_key = ?", key).Delete(&loginAttemptModel{}).Error
	if err != nil {
		return apperr.WrapDatabase(err, "reset login attempts")
	}
	return nil
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

type loginAttemptModel struct {
	AttemptKey    string     `gorm:"primaryKey;type:varchar(128)"`
	Failures      int        `gorm:"not null"`
	FirstFailedAt time.Time  `gorm:"not null"`
	LastFailedAt  time.Time  `gorm:"not null"`
	LockedUntil   *time.Time `gorm:"index"`
	CreatedAt     time.Time  `gorm:"not null"`
	UpdatedAt     time.Time  `gorm:"not null"`
}

func (loginAttemptModel) TableName() string {
	return "auth_login_attempts"
}

func (s *Store) findLoginAttempt(ctx context.Context, key string) (loginAttemptModel, bool, error) {
	var attempt loginAttemptModel
	err := s.db.WithContext(ctx).First(&attempt, "attempt_key = ?", key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return loginAttemptModel{}, false, nil
	}
	if err != nil {
		return loginAttemptModel{}, false, apperr.WrapDatabase(err, "find login attempt")
	}
	return attempt, true, nil
}

func (s *Store) recordLoginFailure(ctx context.Context, tx *gorm.DB, key string, now time.Time) error {
	var attempt loginAttemptModel
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&attempt, "attempt_key = ?", key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		attempt = newLoginAttempt(key, now)
		if createErr := tx.WithContext(ctx).Create(&attempt).Error; createErr != nil {
			if isDuplicateKey(createErr) {
				return s.incrementExistingLoginAttempt(ctx, tx, key, now)
			}
			return createErr
		}
		return nil
	}
	if err != nil {
		return err
	}
	incrementLoginAttempt(&attempt, now)
	return tx.WithContext(ctx).Save(&attempt).Error
}

func (s *Store) incrementExistingLoginAttempt(ctx context.Context, tx *gorm.DB, key string, now time.Time) error {
	var attempt loginAttemptModel
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&attempt, "attempt_key = ?", key).Error
	if err != nil {
		return err
	}
	incrementLoginAttempt(&attempt, now)
	return tx.WithContext(ctx).Save(&attempt).Error
}

func newLoginAttempt(key string, now time.Time) loginAttemptModel {
	attempt := loginAttemptModel{
		AttemptKey:    key,
		Failures:      1,
		FirstFailedAt: now,
		LastFailedAt:  now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return attempt
}

func incrementLoginAttempt(attempt *loginAttemptModel, now time.Time) {
	if loginAttemptExpired(*attempt, now) {
		attempt.Failures = 0
		attempt.FirstFailedAt = now
		attempt.LockedUntil = nil
	}
	attempt.Failures++
	attempt.LastFailedAt = now
	attempt.UpdatedAt = now
	if attempt.Failures >= maxFailedLoginAttempts {
		lockedUntil := now.Add(loginAttemptLockout)
		attempt.LockedUntil = &lockedUntil
	}
}

func loginAttemptExpired(attempt loginAttemptModel, now time.Time) bool {
	return attempt.LastFailedAt.Before(now.Add(-loginAttemptWindow))
}

func normalizeLoginAttemptKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" || len(key) > loginAttemptKeyMaxBytes {
		return "", fmt.Errorf("invalid login attempt key")
	}
	return key, nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *drivermysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
