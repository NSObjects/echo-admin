package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/NSObjects/echo-admin/internal/modules/access/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/pagination"
)

// AllPermissions returns the current foundation permission set.
func AllPermissions() []string {
	return []string{
		domain.PermissionAdminManage,
		domain.PermissionRoleManage,
		domain.PermissionMenuManage,
		domain.PermissionConfigManage,
		domain.PermissionDictManage,
		domain.PermissionFileUpload,
		domain.PermissionLogRead,
	}
}

// ListRoles returns paginated roles.
func (u *Usecase) ListRoles(ctx context.Context, input ListInput) (RoleListOutput, error) {
	if err := u.ready(); err != nil {
		return RoleListOutput{}, err
	}
	filter, err := normalizeListInput(input)
	if err != nil {
		return RoleListOutput{}, err
	}
	roles, total, err := u.store.ListRoles(ctx, filter)
	if err != nil {
		return RoleListOutput{}, err
	}
	return RoleListOutput{
		Items:    mapRoles(roles),
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Total:    total,
	}, nil
}

// CreateRole validates and stores a new role.
func (u *Usecase) CreateRole(ctx context.Context, input RoleInput) (Role, error) {
	if err := u.ready(); err != nil {
		return Role{}, err
	}
	role, err := domain.RestoreRole(0, input.Code, input.Name, input.Permissions, input.MenuIDs, input.Active, time.Time{}, time.Time{})
	if err != nil {
		return Role{}, mapDomainError(err)
	}
	created, err := u.store.CreateRole(ctx, role)
	if err != nil {
		return Role{}, err
	}
	return fromRole(created), nil
}

// UpdateRole applies mutable role changes.
func (u *Usecase) UpdateRole(ctx context.Context, input UpdateRoleInput) (Role, error) {
	if err := u.ready(); err != nil {
		return Role{}, err
	}
	existing, err := u.store.FindRoleByID(ctx, input.ID)
	if err != nil {
		return Role{}, err
	}
	name := existing.Name()
	if input.Name != nil {
		name = *input.Name
	}
	active := existing.Active()
	if input.Active != nil {
		active = *input.Active
	}
	role, err := domain.RestoreRole(
		existing.ID(),
		existing.Code(),
		name,
		coalesceStrings(input.Permissions, existing.Permissions()),
		coalesceIDs(input.MenuIDs, existing.MenuIDs()),
		active,
		existing.CreatedAt(),
		time.Time{},
	)
	if err != nil {
		return Role{}, mapDomainError(err)
	}
	saved, err := u.store.UpdateRole(ctx, role)
	if err != nil {
		return Role{}, err
	}
	return fromRole(saved), nil
}

// ListMenus returns all menus in display order.
func (u *Usecase) ListMenus(ctx context.Context) ([]Menu, error) {
	if err := u.ready(); err != nil {
		return nil, err
	}
	menus, err := u.store.ListMenus(ctx)
	if err != nil {
		return nil, err
	}
	return mapMenus(menus), nil
}

// CreateMenu validates and stores a new menu.
func (u *Usecase) CreateMenu(ctx context.Context, input MenuInput) (Menu, error) {
	if err := u.ready(); err != nil {
		return Menu{}, err
	}
	menu, err := domain.RestoreMenu(0, input.ParentID, input.Name, input.Path, input.Icon, input.Permission, input.Sort, input.Active, time.Time{}, time.Time{})
	if err != nil {
		return Menu{}, mapDomainError(err)
	}
	created, err := u.store.CreateMenu(ctx, menu)
	if err != nil {
		return Menu{}, err
	}
	return fromMenu(created), nil
}

// UpdateMenu replaces mutable menu fields.
func (u *Usecase) UpdateMenu(ctx context.Context, input UpdateMenuInput) (Menu, error) {
	if err := u.ready(); err != nil {
		return Menu{}, err
	}
	existing, err := u.store.FindMenuByID(ctx, input.ID)
	if err != nil {
		return Menu{}, err
	}
	menu, err := domain.RestoreMenu(input.ID, input.ParentID, input.Name, input.Path, input.Icon, input.Permission, input.Sort, input.Active, existing.CreatedAt(), time.Time{})
	if err != nil {
		return Menu{}, mapDomainError(err)
	}
	updated, err := u.store.UpdateMenu(ctx, menu)
	if err != nil {
		return Menu{}, err
	}
	return fromMenu(updated), nil
}

func (u *Usecase) ready() error {
	if u == nil || u.store == nil {
		return apperr.New(apperr.ErrInternalServer, "access store is not configured")
	}
	return nil
}

func normalizeListInput(input ListInput) (ListFilter, error) {
	window, err := pagination.Normalize(input.Page, input.PageSize, pagination.Options{
		DefaultPageSize: defaultPageSize,
		MaxPageSize:     maxPageSize,
	})
	if err != nil {
		return ListFilter{}, apperr.NewBadRequest("invalid pagination")
	}
	return ListFilter{Offset: window.Offset, Limit: window.Limit, Page: window.Page, PageSize: window.PageSize}, nil
}

func mapRoles(roles []domain.Role) []Role {
	out := make([]Role, 0, len(roles))
	for _, role := range roles {
		out = append(out, fromRole(role))
	}
	return out
}

func mapMenus(menus []domain.Menu) []Menu {
	out := make([]Menu, 0, len(menus))
	for _, menu := range menus {
		out = append(out, fromMenu(menu))
	}
	return out
}

func fromRole(role domain.Role) Role {
	return Role{
		ID:          role.ID(),
		Code:        role.Code(),
		Name:        role.Name(),
		Permissions: role.Permissions(),
		MenuIDs:     role.MenuIDs(),
		Active:      role.Active(),
		CreatedAt:   role.CreatedAt(),
		UpdatedAt:   role.UpdatedAt(),
	}
}

func fromMenu(menu domain.Menu) Menu {
	return Menu{
		ID:         menu.ID(),
		ParentID:   menu.ParentID(),
		Name:       menu.Name(),
		Path:       menu.Path(),
		Icon:       menu.Icon(),
		Permission: menu.Permission(),
		Sort:       menu.Sort(),
		Active:     menu.Active(),
		CreatedAt:  menu.CreatedAt(),
		UpdatedAt:  menu.UpdatedAt(),
	}
}

func coalesceIDs(next, fallback []int64) []int64 {
	if next == nil {
		return fallback
	}
	return next
}

func coalesceStrings(next, fallback []string) []string {
	if next == nil {
		return fallback
	}
	return next
}

func mapDomainError(err error) error {
	for _, entry := range domainErrorMessages {
		if errors.Is(err, entry.err) {
			return apperr.NewBadRequest(entry.message)
		}
	}
	return err
}

var domainErrorMessages = []struct {
	err     error
	message string
}{
	{domain.ErrInvalidRoleID, "invalid role id"},
	{domain.ErrInvalidRoleCode, "invalid role code"},
	{domain.ErrInvalidRoleName, "invalid role name"},
	{domain.ErrRoleNeedsPerms, "role needs permissions"},
	{domain.ErrInvalidMenuID, "invalid menu id"},
	{domain.ErrInvalidMenuName, "invalid menu name"},
	{domain.ErrInvalidMenuPath, "invalid menu path"},
	{domain.ErrInvalidPermission, "invalid permission"},
}
