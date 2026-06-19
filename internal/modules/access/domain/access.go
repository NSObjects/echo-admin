// Package domain contains role, permission, and menu business rules.
package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	// RoleCodeSuperAdmin is the seeded root role with every platform grant.
	RoleCodeSuperAdmin = "super_admin"
	// DefaultRolePath is used when a role has not chosen a more specific home page.
	DefaultRolePath = "/dashboard"
)

// Permission names used by the back-office foundation modules.
const (
	PermissionAdminRead   = "admin:read"
	PermissionAdminCreate = "admin:create"
	PermissionAdminUpdate = "admin:update"
	PermissionAdminDelete = "admin:delete"

	PermissionRoleRead   = "role:read"
	PermissionRoleCreate = "role:create"
	PermissionRoleUpdate = "role:update"
	PermissionRoleDelete = "role:delete"

	PermissionMenuRead   = "menu:read"
	PermissionMenuCreate = "menu:create"
	PermissionMenuUpdate = "menu:update"
	PermissionMenuDelete = "menu:delete"

	PermissionConfigRead   = "config:read"
	PermissionConfigUpdate = "config:update"

	PermissionDictRead   = "dict:read"
	PermissionDictCreate = "dict:create"
	PermissionDictUpdate = "dict:update"
	PermissionDictDelete = "dict:delete"

	PermissionFileRead   = "file:read"
	PermissionFileUpload = "file:upload"

	PermissionLogRead = "log:read"
)

// PermissionDefinition describes one grantable action exposed to role managers.
type PermissionDefinition struct {
	Token    string
	Resource string
	Action   string
	Name     string
}

// PermissionCatalog returns the canonical grant list used by seed and UI metadata.
func PermissionCatalog() []PermissionDefinition {
	return append([]PermissionDefinition(nil), permissionCatalog...)
}

// Access validation errors.
var (
	ErrInvalidRoleID      = errors.New("invalid role id")
	ErrInvalidRoleParent  = errors.New("invalid role parent")
	ErrInvalidRoleCode    = errors.New("invalid role code")
	ErrInvalidRoleName    = errors.New("invalid role name")
	ErrRoleNeedsPerms     = errors.New("role needs at least one permission")
	ErrInvalidDefaultPath = errors.New("invalid default path")
	ErrInvalidMenuID      = errors.New("invalid menu id")
	ErrInvalidMenuName    = errors.New("invalid menu name")
	ErrInvalidMenuPath    = errors.New("invalid menu path")
	ErrInvalidPermission  = errors.New("invalid permission")
)

// Role groups permissions, menu visibility, and delegation hierarchy for administrators.
type Role struct {
	ID          int64
	ParentID    int64
	Code        string
	Name        string
	Permissions []string
	MenuIDs     []int64
	DefaultPath string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RestoreRole rebuilds a role from a trusted store representation.
func RestoreRole(id, parentID int64, code, name string, permissions []string, menuIDs []int64, defaultPath string, active bool, createdAt, updatedAt time.Time) (Role, error) {
	code = normalizeToken(code)
	name = strings.TrimSpace(name)
	defaultPath = normalizeDefaultPath(defaultPath)
	perms, permissionsOK := uniquePermissionTokens(permissions)

	if id < 0 {
		return Role{}, ErrInvalidRoleID
	}
	if parentID < 0 || (id > 0 && parentID == id) {
		return Role{}, ErrInvalidRoleParent
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
	if !isSafePath(defaultPath) {
		return Role{}, ErrInvalidDefaultPath
	}
	return Role{
		ID:          id,
		ParentID:    parentID,
		Code:        code,
		Name:        name,
		Permissions: perms,
		MenuIDs:     uniquePositiveIDs(menuIDs),
		DefaultPath: defaultPath,
		Active:      active,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// IsSuperAdmin reports whether the role is the seeded root administrator role.
func (r Role) IsSuperAdmin() bool { return r.Code == RoleCodeSuperAdmin }

// Menu represents one visible navigation entry in the back-office UI.
type Menu struct {
	ID         int64
	ParentID   int64
	Name       string
	Path       string
	Icon       string
	Permission string
	Sort       int
	Active     bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// RestoreMenu rebuilds a menu from a trusted store representation.
func RestoreMenu(id, parentID int64, name, path, icon, permission string, sort int, active bool, createdAt, updatedAt time.Time) (Menu, error) {
	name = strings.TrimSpace(name)
	path = strings.TrimSpace(path)
	icon = strings.TrimSpace(icon)
	permission = normalizeToken(permission)

	if id < 0 || parentID < 0 || (id > 0 && parentID == id) {
		return Menu{}, ErrInvalidMenuID
	}
	if name == "" || len(name) > 80 {
		return Menu{}, ErrInvalidMenuName
	}
	if !isSafePath(path) || len(path) > 160 {
		return Menu{}, ErrInvalidMenuPath
	}
	if permission != "" && !isPermissionToken(permission) {
		return Menu{}, ErrInvalidPermission
	}
	return Menu{
		ID:         id,
		ParentID:   parentID,
		Name:       name,
		Path:       path,
		Icon:       icon,
		Permission: permission,
		Sort:       sort,
		Active:     active,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

var permissionCatalog = []PermissionDefinition{
	{Token: PermissionAdminRead, Resource: "管理员", Action: "查看", Name: "查看管理员"},
	{Token: PermissionAdminCreate, Resource: "管理员", Action: "创建", Name: "创建管理员"},
	{Token: PermissionAdminUpdate, Resource: "管理员", Action: "更新", Name: "更新管理员"},
	{Token: PermissionAdminDelete, Resource: "管理员", Action: "删除", Name: "删除管理员"},
	{Token: PermissionRoleRead, Resource: "角色", Action: "查看", Name: "查看角色"},
	{Token: PermissionRoleCreate, Resource: "角色", Action: "创建", Name: "创建角色"},
	{Token: PermissionRoleUpdate, Resource: "角色", Action: "更新", Name: "更新角色"},
	{Token: PermissionRoleDelete, Resource: "角色", Action: "删除", Name: "删除角色"},
	{Token: PermissionMenuRead, Resource: "菜单", Action: "查看", Name: "查看菜单"},
	{Token: PermissionMenuCreate, Resource: "菜单", Action: "创建", Name: "创建菜单"},
	{Token: PermissionMenuUpdate, Resource: "菜单", Action: "更新", Name: "更新菜单"},
	{Token: PermissionMenuDelete, Resource: "菜单", Action: "删除", Name: "删除菜单"},
	{Token: PermissionConfigRead, Resource: "系统配置", Action: "查看", Name: "查看系统配置"},
	{Token: PermissionConfigUpdate, Resource: "系统配置", Action: "更新", Name: "更新系统配置"},
	{Token: PermissionDictRead, Resource: "数据字典", Action: "查看", Name: "查看数据字典"},
	{Token: PermissionDictCreate, Resource: "数据字典", Action: "创建", Name: "创建数据字典"},
	{Token: PermissionDictUpdate, Resource: "数据字典", Action: "更新", Name: "更新数据字典"},
	{Token: PermissionDictDelete, Resource: "数据字典", Action: "删除", Name: "删除数据字典"},
	{Token: PermissionFileRead, Resource: "文件", Action: "查看", Name: "查看文件"},
	{Token: PermissionFileUpload, Resource: "文件", Action: "上传", Name: "上传文件"},
	{Token: PermissionLogRead, Resource: "审计日志", Action: "查看", Name: "查看日志"},
}

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

func isSafePath(value string) bool {
	return value != "" && strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//")
}

func normalizeDefaultPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultRolePath
	}
	return value
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
