// Package usecase coordinates authentication and authorization workflows.
package usecase

import (
	"context"
	"time"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	identitydomain "github.com/NSObjects/echo-admin/internal/modules/identity/domain"
)

// AdminReader reads administrator credentials and profiles for authentication.
type AdminReader interface {
	FindByUsername(context.Context, string) (identitydomain.Admin, error)
	FindByID(context.Context, int64) (identitydomain.Admin, error)
}

// RoleReader reads roles needed to build authorization grants.
type RoleReader interface {
	FindRoleByID(context.Context, int64) (accessdomain.Role, error)
}

// MenuReader reads menus needed to build the current-user navigation tree.
type MenuReader interface {
	ListMenus(context.Context) ([]accessdomain.Menu, error)
}

// LoginRecorder stores sign-in attempts without exposing audit storage details.
type LoginRecorder interface {
	RecordLogin(context.Context, LoginRecord) error
}

// Usecase coordinates sign-in, current user, and permission checks.
type Usecase struct {
	admins    AdminReader
	roles     RoleReader
	menus     MenuReader
	logins    LoginRecorder
	jwtSecret []byte
	now       func() time.Time
}

// Option customizes the auth usecase.
type Option func(*Usecase)

// WithClock replaces the clock used for tokens and records.
func WithClock(now func() time.Time) Option {
	return func(u *Usecase) {
		if now != nil {
			u.now = now
		}
	}
}

// New creates an auth usecase with its required readers and JWT secret.
func New(admins AdminReader, roles RoleReader, menus MenuReader, logins LoginRecorder, jwtSecret string, opts ...Option) *Usecase {
	u := &Usecase{
		admins:    admins,
		roles:     roles,
		menus:     menus,
		logins:    logins,
		jwtSecret: []byte(jwtSecret),
		now:       func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		if opt != nil {
			opt(u)
		}
	}
	return u
}

// LoginInput carries administrator credentials from a delivery adapter.
type LoginInput struct {
	Username  string
	Password  string
	IP        string
	UserAgent string
}

// LoginOutput is returned after successful authentication.
type LoginOutput struct {
	Token string `json:"token"`
	User  Admin  `json:"user"`
}

// LoginRecord carries a safe sign-in audit event.
type LoginRecord struct {
	AdminID   int64
	Username  string
	IP        string
	UserAgent string
	Success   bool
	Reason    string
}

// Admin is the current-user administrator DTO.
type Admin struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	RoleIDs     []int64   `json:"role_ids"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CurrentUser is the adapter-facing current-user snapshot.
type CurrentUser struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Roles       []Role   `json:"roles"`
	Permissions []string `json:"permissions"`
	Menus       []Menu   `json:"menus"`
}

// Role is the current-user role DTO.
type Role struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	MenuIDs     []int64   `json:"menu_ids"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Menu is the current-user menu DTO.
type Menu struct {
	ID         int64     `json:"id"`
	ParentID   int64     `json:"parent_id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Icon       string    `json:"icon"`
	Permission string    `json:"permission"`
	Sort       int       `json:"sort"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
