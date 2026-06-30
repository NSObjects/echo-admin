// Package usecase coordinates role and menu access workflows.
package usecase

import (
	"context"
	"time"

	"github.com/NSObjects/echo-admin/internal/modules/access/domain"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// Store persists roles and menus for the access usecase.
type Store interface {
	FindRoleByID(context.Context, int64) (domain.Role, error)
	FindRoleByCode(context.Context, string) (domain.Role, error)
	ListAllRoles(context.Context) ([]domain.Role, error)
	CreateRole(context.Context, domain.Role) (domain.Role, error)
	UpdateRole(context.Context, domain.Role) (domain.Role, error)
	DeleteRole(context.Context, int64) error
	FindAPIByID(context.Context, int64) (domain.API, error)
	FindAPIByRoute(context.Context, string, string) (domain.API, error)
	ListAPIs(context.Context) ([]domain.API, error)
	CreateAPI(context.Context, domain.API) (domain.API, error)
	UpdateAPI(context.Context, domain.API) (domain.API, error)
	DeleteAPI(context.Context, int64) error
	FindMenuByID(context.Context, int64) (domain.Menu, error)
	ListMenus(context.Context) ([]domain.Menu, error)
	CreateMenu(context.Context, domain.Menu) (domain.Menu, error)
	UpdateMenu(context.Context, domain.Menu) (domain.Menu, error)
	DeleteMenu(context.Context, int64) error
}

// AdminRoleReader reads the current administrator role assignment without exposing identity storage.
type AdminRoleReader interface {
	AdminRoleState(context.Context, int64) (AdminRoleState, error)
	RoleAssigned(context.Context, int64) (bool, error)
}

// AdminRoleState is the minimal identity snapshot needed for scoped role delegation.
type AdminRoleState struct {
	RoleIDs      []int64
	ActiveRoleID int64
}

// Usecase coordinates role and menu rules.
type Usecase struct {
	store  Store
	admins AdminRoleReader
}

// New creates an access usecase.
func New(store Store, admins AdminRoleReader) *Usecase {
	return &Usecase{store: store, admins: admins}
}

// RoleInput carries mutable role fields.
type RoleInput struct {
	ParentID    int64
	Code        string
	Name        string
	Permissions []string
	MenuIDs     []int64
	APIIDs      []int64
	ButtonIDs   []int64
	DataRoleIDs []int64
	DefaultPath string
	Active      bool
}

// UpdateRoleInput carries partial role updates.
type UpdateRoleInput struct {
	ID          int64
	ParentID    *int64
	Name        *string
	Permissions []string
	MenuIDs     []int64
	APIIDs      []int64
	ButtonIDs   []int64
	DataRoleIDs []int64
	DefaultPath *string
	Active      *bool
}

// CopyRoleInput carries the new identity fields for a copied role.
type CopyRoleInput struct {
	SourceID    int64
	ParentID    *int64
	Code        string
	Name        string
	DefaultPath *string
	Active      *bool
}

// MenuInput carries mutable menu fields.
type MenuInput struct {
	ParentID   int64
	Name       string
	Path       string
	Icon       string
	Hidden     bool
	Component  string
	Meta       MenuMetaInput
	Permission string
	Sort       int
	Active     bool
	Buttons    []MenuButtonInput
}

// MenuMetaInput carries router metadata for one menu.
type MenuMetaInput struct {
	ActiveName     string
	KeepAlive      bool
	DefaultMenu    bool
	CloseTab       bool
	TransitionType string
}

// MenuButtonInput carries one page-level operation key attached to a menu.
type MenuButtonInput struct {
	ID          int64
	Name        string
	Description string
}

// APIInput carries mutable API metadata.
type APIInput struct {
	Method      string
	Path        string
	Description string
	Group       string
	Permission  string
	Public      bool
}

// UpdateAPIInput carries mutable API updates.
type UpdateAPIInput struct {
	ID          int64
	Method      string
	Path        string
	Description string
	Group       string
	Permission  string
	Public      bool
}

// MenuRolesInput carries the full role assignment for one menu.
type MenuRolesInput struct {
	MenuID  int64
	RoleIDs []int64
}

// APIRolesInput carries the full role assignment for one API route.
type APIRolesInput struct {
	APIID   int64
	RoleIDs []int64
}

// UpdateMenuInput carries mutable menu updates.
type UpdateMenuInput struct {
	ID         int64
	ParentID   int64
	Name       string
	Path       string
	Icon       string
	Hidden     bool
	Component  string
	Meta       MenuMetaInput
	Permission string
	Sort       int
	Active     bool
	Buttons    []MenuButtonInput
}

// ListInput carries pagination for role lists.
type ListInput struct {
	Page     int
	PageSize int
}

// ListFilter is the validated store-facing pagination window.
type ListFilter struct {
	Offset   int
	Limit    int
	Page     int
	PageSize int
}

// RoleListOutput is a paginated role result.
type RoleListOutput struct {
	Items    []Role
	Page     int
	PageSize int
	Total    int
}

// APIListOutput is a paginated API result.
type APIListOutput struct {
	Items    []API
	Page     int
	PageSize int
	Total    int
}

// Role is the adapter-facing role DTO.
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

// Menu is the adapter-facing menu DTO.
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

// MenuMeta is the adapter-facing router metadata DTO.
type MenuMeta struct {
	ActiveName     string `json:"active_name"`
	KeepAlive      bool   `json:"keep_alive"`
	DefaultMenu    bool   `json:"default_menu"`
	CloseTab       bool   `json:"close_tab"`
	TransitionType string `json:"transition_type"`
}

// Button is the adapter-facing menu button DTO.
type Button struct {
	ID          int64     `json:"id"`
	MenuID      int64     `json:"menu_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// API is the adapter-facing API route DTO.
type API struct {
	ID          int64     `json:"id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	Description string    `json:"description"`
	Group       string    `json:"group"`
	Permission  string    `json:"permission"`
	Public      bool      `json:"public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PermissionDefinition is the adapter-facing permission metadata DTO.
type PermissionDefinition struct {
	Token    string `json:"token"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Name     string `json:"name"`
}
