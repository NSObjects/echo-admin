// Package domain contains role, permission, and menu business rules.
package domain

import (
	"errors"
	"strings"
	"time"
)

// Permission names used by the back-office foundation modules.
const (
	PermissionAdminManage  = "admin:manage"
	PermissionRoleManage   = "role:manage"
	PermissionMenuManage   = "menu:manage"
	PermissionConfigManage = "config:manage"
	PermissionDictManage   = "dict:manage"
	PermissionFileUpload   = "file:upload"
	PermissionLogRead      = "log:read"
)

// Access validation errors.
var (
	ErrInvalidRoleID     = errors.New("invalid role id")
	ErrInvalidRoleCode   = errors.New("invalid role code")
	ErrInvalidRoleName   = errors.New("invalid role name")
	ErrRoleNeedsPerms    = errors.New("role needs at least one permission")
	ErrInvalidMenuID     = errors.New("invalid menu id")
	ErrInvalidMenuName   = errors.New("invalid menu name")
	ErrInvalidMenuPath   = errors.New("invalid menu path")
	ErrInvalidPermission = errors.New("invalid permission")
)

// Role groups permissions and menu visibility for administrators.
type Role struct {
	id          int64
	code        string
	name        string
	permissions []string
	menuIDs     []int64
	active      bool
	createdAt   time.Time
	updatedAt   time.Time
}

// RestoreRole rebuilds a role from a trusted store representation.
func RestoreRole(id int64, code, name string, permissions []string, menuIDs []int64, active bool, createdAt, updatedAt time.Time) (Role, error) {
	code = normalizeToken(code)
	name = strings.TrimSpace(name)
	perms, permissionsOK := uniquePermissionTokens(permissions)

	if id < 0 {
		return Role{}, ErrInvalidRoleID
	}
	if code == "" || len(code) > 64 {
		return Role{}, ErrInvalidRoleCode
	}
	if name == "" || len(name) > 80 {
		return Role{}, ErrInvalidRoleName
	}
	if !permissionsOK {
		return Role{}, ErrInvalidPermission
	}
	if len(perms) == 0 {
		return Role{}, ErrRoleNeedsPerms
	}
	return Role{
		id:          id,
		code:        code,
		name:        name,
		permissions: perms,
		menuIDs:     uniquePositiveIDs(menuIDs),
		active:      active,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

// ID returns the persisted role id.
func (r Role) ID() int64 { return r.id }

// Code returns the unique role code.
func (r Role) Code() string { return r.code }

// Name returns the operator-facing role name.
func (r Role) Name() string { return r.name }

// Permissions returns the permissions granted by the role.
func (r Role) Permissions() []string { return append([]string(nil), r.permissions...) }

// MenuIDs returns the menu ids visible to the role.
func (r Role) MenuIDs() []int64 { return append([]int64(nil), r.menuIDs...) }

// Active reports whether the role grants permissions.
func (r Role) Active() bool { return r.active }

// CreatedAt returns the creation timestamp.
func (r Role) CreatedAt() time.Time { return r.createdAt }

// UpdatedAt returns the last update timestamp.
func (r Role) UpdatedAt() time.Time { return r.updatedAt }

// Menu represents one visible navigation entry in the back-office UI.
type Menu struct {
	id         int64
	parentID   int64
	name       string
	path       string
	icon       string
	permission string
	sort       int
	active     bool
	createdAt  time.Time
	updatedAt  time.Time
}

// RestoreMenu rebuilds a menu from a trusted store representation.
func RestoreMenu(id, parentID int64, name, path, icon, permission string, sort int, active bool, createdAt, updatedAt time.Time) (Menu, error) {
	name = strings.TrimSpace(name)
	path = strings.TrimSpace(path)
	icon = strings.TrimSpace(icon)
	permission = normalizeToken(permission)

	if id < 0 || parentID < 0 {
		return Menu{}, ErrInvalidMenuID
	}
	if name == "" || len(name) > 80 {
		return Menu{}, ErrInvalidMenuName
	}
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || len(path) > 160 {
		return Menu{}, ErrInvalidMenuPath
	}
	if permission != "" && !isPermissionToken(permission) {
		return Menu{}, ErrInvalidPermission
	}
	return Menu{
		id:         id,
		parentID:   parentID,
		name:       name,
		path:       path,
		icon:       icon,
		permission: permission,
		sort:       sort,
		active:     active,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}, nil
}

// ID returns the persisted menu id.
func (m Menu) ID() int64 { return m.id }

// ParentID returns the parent menu id, or zero for root menus.
func (m Menu) ParentID() int64 { return m.parentID }

// Name returns the visible menu name.
func (m Menu) Name() string { return m.name }

// Path returns the frontend route path.
func (m Menu) Path() string { return m.path }

// Icon returns the frontend icon token.
func (m Menu) Icon() string { return m.icon }

// Permission returns the permission required to see the menu.
func (m Menu) Permission() string { return m.permission }

// Sort returns the menu ordering key.
func (m Menu) Sort() int { return m.sort }

// Active reports whether the menu is visible when permissions allow it.
func (m Menu) Active() bool { return m.active }

// CreatedAt returns the creation timestamp.
func (m Menu) CreatedAt() time.Time { return m.createdAt }

// UpdatedAt returns the last update timestamp.
func (m Menu) UpdatedAt() time.Time { return m.updatedAt }

func uniquePermissionTokens(values []string) ([]string, bool) {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		token := normalizeToken(value)
		if !isPermissionToken(token) {
			return nil, false
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	return out, true
}

func isPermissionToken(value string) bool {
	if value == "" || len(value) > 80 {
		return false
	}
	parts := strings.Split(value, ":")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
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
