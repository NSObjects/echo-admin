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
	ListRoles(context.Context, ListFilter) ([]domain.Role, int, error)
	CreateRole(context.Context, domain.Role) (domain.Role, error)
	UpdateRole(context.Context, domain.Role) (domain.Role, error)
	FindMenuByID(context.Context, int64) (domain.Menu, error)
	ListMenus(context.Context) ([]domain.Menu, error)
	CreateMenu(context.Context, domain.Menu) (domain.Menu, error)
	UpdateMenu(context.Context, domain.Menu) (domain.Menu, error)
}

// Usecase coordinates role and menu rules.
type Usecase struct {
	store Store
}

// New creates an access usecase.
func New(store Store) *Usecase {
	return &Usecase{store: store}
}

// RoleInput carries mutable role fields.
type RoleInput struct {
	Code        string
	Name        string
	Permissions []string
	MenuIDs     []int64
	Active      bool
}

// UpdateRoleInput carries partial role updates.
type UpdateRoleInput struct {
	ID          int64
	Name        *string
	Permissions []string
	MenuIDs     []int64
	Active      *bool
}

// MenuInput carries mutable menu fields.
type MenuInput struct {
	ParentID   int64
	Name       string
	Path       string
	Icon       string
	Permission string
	Sort       int
	Active     bool
}

// UpdateMenuInput carries mutable menu updates.
type UpdateMenuInput struct {
	ID         int64
	ParentID   int64
	Name       string
	Path       string
	Icon       string
	Permission string
	Sort       int
	Active     bool
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

// Role is the adapter-facing role DTO.
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

// Menu is the adapter-facing menu DTO.
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
