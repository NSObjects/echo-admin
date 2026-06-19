// Package domain contains the administrator management business rules.
package domain

import (
	"errors"
	"strings"
	"time"
)

// Administrator validation errors.
var (
	ErrInvalidAdminID      = errors.New("invalid admin id")
	ErrInvalidUsername     = errors.New("invalid username")
	ErrInvalidDisplayName  = errors.New("invalid display name")
	ErrInvalidPasswordHash = errors.New("invalid password hash")
	ErrAdminNeedsRole      = errors.New("admin needs at least one role")
)

// Admin is a back-office user that can sign in and operate management APIs.
type Admin struct {
	id           int64
	username     string
	displayName  string
	email        string
	passwordHash []byte
	roleIDs      []int64
	active       bool
	createdAt    time.Time
	updatedAt    time.Time
}

// NewAdmin validates and creates an admin before it is persisted.
func NewAdmin(username, displayName, email string, passwordHash []byte, roleIDs []int64) (Admin, error) {
	return RestoreAdmin(0, username, displayName, email, passwordHash, roleIDs, true, time.Time{}, time.Time{})
}

// RestoreAdmin rebuilds an admin from a trusted store representation.
func RestoreAdmin(id int64, username, displayName, email string, passwordHash []byte, roleIDs []int64, active bool, createdAt, updatedAt time.Time) (Admin, error) {
	username = normalizeToken(username)
	displayName = strings.TrimSpace(displayName)
	email = strings.TrimSpace(email)

	if id < 0 {
		return Admin{}, ErrInvalidAdminID
	}
	if username == "" || len(username) > 64 {
		return Admin{}, ErrInvalidUsername
	}
	if displayName == "" || len(displayName) > 80 {
		return Admin{}, ErrInvalidDisplayName
	}
	if len(passwordHash) == 0 {
		return Admin{}, ErrInvalidPasswordHash
	}
	copiedRoles := uniquePositiveIDs(roleIDs)
	if len(copiedRoles) == 0 {
		return Admin{}, ErrAdminNeedsRole
	}
	return Admin{
		id:           id,
		username:     username,
		displayName:  displayName,
		email:        email,
		passwordHash: append([]byte(nil), passwordHash...),
		roleIDs:      copiedRoles,
		active:       active,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}

// ID returns the persisted admin id.
func (a Admin) ID() int64 { return a.id }

// Username returns the unique login name.
func (a Admin) Username() string { return a.username }

// DisplayName returns the operator-facing admin name.
func (a Admin) DisplayName() string { return a.displayName }

// Email returns the optional contact email.
func (a Admin) Email() string { return a.email }

// PasswordHash returns a copy of the stored password hash.
func (a Admin) PasswordHash() []byte { return append([]byte(nil), a.passwordHash...) }

// RoleIDs returns a copy of the role ids assigned to the admin.
func (a Admin) RoleIDs() []int64 { return append([]int64(nil), a.roleIDs...) }

// Active reports whether the admin can sign in and operate APIs.
func (a Admin) Active() bool { return a.active }

// CreatedAt returns the creation timestamp.
func (a Admin) CreatedAt() time.Time { return a.createdAt }

// UpdatedAt returns the last update timestamp.
func (a Admin) UpdatedAt() time.Time { return a.updatedAt }

// UpdateProfile changes mutable admin fields while preserving identity and password.
func (a Admin) UpdateProfile(displayName, email string, roleIDs []int64, active bool) (Admin, error) {
	return RestoreAdmin(a.id, a.username, displayName, email, a.passwordHash, roleIDs, active, a.createdAt, time.Time{})
}

// ReplacePassword changes the admin password hash.
func (a Admin) ReplacePassword(passwordHash []byte) (Admin, error) {
	return RestoreAdmin(a.id, a.username, a.displayName, a.email, passwordHash, a.roleIDs, a.active, a.createdAt, time.Time{})
}

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func uniquePositiveIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
