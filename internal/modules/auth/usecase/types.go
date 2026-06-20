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
	Update(context.Context, identitydomain.Admin) (identitydomain.Admin, error)
}

// RoleReader reads roles needed to build authorization grants.
type RoleReader interface {
	FindRoleByID(context.Context, int64) (accessdomain.Role, error)
}

// MenuReader reads menus needed to build the current-user navigation tree.
type MenuReader interface {
	ListMenus(context.Context) ([]accessdomain.Menu, error)
}

// APIReader reads managed backend routes needed for API-level authorization.
type APIReader interface {
	ListAPIs(context.Context) ([]accessdomain.API, error)
}

// LoginRecorder stores sign-in attempts without exposing audit storage details.
type LoginRecorder interface {
	RecordLogin(context.Context, LoginRecord) error
}

// JWTBlacklistStore stores revoked JWT identifiers without persisting the raw
// bearer token. Logout and middleware lookup share the same hash contract.
type JWTBlacklistStore interface {
	AddJWTBlacklist(context.Context, JWTBlacklistEntry) error
	JWTBlacklisted(context.Context, string, time.Time) (bool, error)
}

// Usecase coordinates sign-in, current user, and permission checks.
type Usecase struct {
	admins       AdminReader
	roles        RoleReader
	menus        MenuReader
	apis         APIReader
	logins       LoginRecorder
	jwtBlacklist JWTBlacklistStore
	jwtSecret    []byte
	now          func() time.Time
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

// New creates an auth usecase with its required readers, JWT revocation store,
// and JWT secret.
func New(admins AdminReader, roles RoleReader, menus MenuReader, apis APIReader, jwtBlacklist JWTBlacklistStore, logins LoginRecorder, jwtSecret string, opts ...Option) *Usecase {
	u := &Usecase{
		admins:       admins,
		roles:        roles,
		menus:        menus,
		apis:         apis,
		logins:       logins,
		jwtBlacklist: jwtBlacklist,
		jwtSecret:    []byte(jwtSecret),
		now:          func() time.Time { return time.Now().UTC() },
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
	Token string      `json:"token"`
	User  CurrentUser `json:"user"`
}

// RoleSwitchInput carries the requested active role.
type RoleSwitchInput struct {
	RoleID int64
}

// ChangePasswordInput carries current-user password rotation data.
type ChangePasswordInput struct {
	CurrentPassword string
	NewPassword     string
	RawToken        string
}

// UpdateProfileInput carries current-user profile fields.
type UpdateProfileInput struct {
	DisplayName string
	Email       string
}

// RoleSwitchOutput is returned after the active role changes.
type RoleSwitchOutput struct {
	Token string      `json:"token"`
	User  CurrentUser `json:"user"`
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

// JWTBlacklistEntry is the persistence contract for one revoked bearer token.
type JWTBlacklistEntry struct {
	TokenHash string
	ExpiresAt time.Time
}

// Admin is the current-user administrator DTO.
type Admin struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name"`
	Email        string    `json:"email"`
	RoleIDs      []int64   `json:"role_ids"`
	ActiveRoleID int64     `json:"active_role_id"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CurrentUser is the adapter-facing current-user snapshot.
type CurrentUser struct {
	ID           int64    `json:"id"`
	Username     string   `json:"username"`
	DisplayName  string   `json:"display_name"`
	Email        string   `json:"email"`
	ActiveRoleID int64    `json:"active_role_id"`
	ActiveRole   Role     `json:"active_role"`
	DefaultPath  string   `json:"default_path"`
	Roles        []Role   `json:"roles"`
	Permissions  []string `json:"permissions"`
	Menus        []Menu   `json:"menus"`
}

// Role is the current-user role DTO.
type Role struct {
	ID          int64     `json:"id"`
	ParentID    int64     `json:"parent_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	MenuIDs     []int64   `json:"menu_ids"`
	APIIDs      []int64   `json:"api_ids"`
	ButtonIDs   []int64   `json:"button_ids"`
	DataRoleIDs []int64   `json:"data_role_ids"`
	DefaultPath string    `json:"default_path"`
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
	Hidden     bool      `json:"hidden"`
	Component  string    `json:"component"`
	Meta       MenuMeta  `json:"meta"`
	Permission string    `json:"permission"`
	Sort       int       `json:"sort"`
	Active     bool      `json:"active"`
	Buttons    []Button  `json:"buttons"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MenuMeta is router metadata returned with the current-user menu tree.
type MenuMeta struct {
	ActiveName     string `json:"active_name"`
	KeepAlive      bool   `json:"keep_alive"`
	DefaultMenu    bool   `json:"default_menu"`
	CloseTab       bool   `json:"close_tab"`
	TransitionType string `json:"transition_type"`
}

// Button is a page-level operation key returned with the current user's menu.
type Button struct {
	ID          int64     `json:"id"`
	MenuID      int64     `json:"menu_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
