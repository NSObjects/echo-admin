// Package usecase coordinates authentication and authorization workflows.
package usecase

import (
	"context"
	"time"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	authdomain "github.com/NSObjects/echo-admin/internal/modules/auth/domain"
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

// LoginSessionStore persists browser login sessions without storing the raw
// cookie credential.
type LoginSessionStore interface {
	CreateLoginSession(context.Context, authdomain.LoginSession) (authdomain.LoginSession, error)
	FindLoginSessionByTokenHash(context.Context, string) (authdomain.LoginSession, bool, error)
	RefreshLoginSession(context.Context, authdomain.LoginSession) error
	UpdateLoginSessionRole(context.Context, int64, int64, time.Time) error
	RevokeLoginSession(context.Context, int64, string, time.Time) error
	RevokeOtherLoginSessions(context.Context, int64, int64, string, time.Time) error
	RevokeLoginSessions(context.Context, int64, string, time.Time) error
}

// LoginAttemptLimiter blocks repeated failed sign-in attempts across app
// instances. The key is already hashed by the usecase so stores do not need to
// persist raw usernames or client addresses.
type LoginAttemptLimiter interface {
	CheckLoginAttempt(context.Context, string, time.Time) error
	RecordLoginFailure(context.Context, string, time.Time) error
	ResetLoginAttempts(context.Context, string) error
}

// Usecase coordinates sign-in, current user, and permission checks.
type Usecase struct {
	admins       AdminReader
	roles        RoleReader
	menus        MenuReader
	apis         APIReader
	logins       LoginRecorder
	sessions     LoginSessionStore
	loginLimiter LoginAttemptLimiter
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

// New creates an auth usecase with its required readers and login-session
// store.
func New(admins AdminReader, roles RoleReader, menus MenuReader, apis APIReader, sessions LoginSessionStore, loginLimiter LoginAttemptLimiter, logins LoginRecorder, opts ...Option) *Usecase {
	u := &Usecase{
		admins:       admins,
		roles:        roles,
		menus:        menus,
		apis:         apis,
		logins:       logins,
		sessions:     sessions,
		loginLimiter: loginLimiter,
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
	SessionToken     string      `json:"-"`
	SessionExpiresAt time.Time   `json:"-"`
	User             CurrentUser `json:"user"`
}

// RoleSwitchInput carries the requested active role.
type RoleSwitchInput struct {
	RoleID int64
}

// ChangePasswordInput carries current-user password rotation data.
type ChangePasswordInput struct {
	CurrentPassword string
	NewPassword     string
}

// UpdateProfileInput carries current-user profile fields.
type UpdateProfileInput struct {
	DisplayName string
	Email       string
}

// RoleSwitchOutput is returned after the active role changes.
type RoleSwitchOutput struct {
	User CurrentUser `json:"user"`
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

// LoginSessionIdentity is the authenticated browser login identity stored on
// request context by HTTP middleware.
type LoginSessionIdentity struct {
	SessionID int64
	AdminID   int64
	RoleID    int64
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
