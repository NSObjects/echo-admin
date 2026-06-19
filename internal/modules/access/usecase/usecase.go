package usecase

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/NSObjects/echo-admin/internal/modules/access/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/pagination"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
)

// AllPermissions returns the current foundation permission set.
func AllPermissions() []string {
	catalog := domain.PermissionCatalog()
	out := make([]string, 0, len(catalog))
	for _, permission := range catalog {
		out = append(out, permission.Token)
	}
	return out
}

// ListPermissions returns grant metadata for role editing UIs.
func (u *Usecase) ListPermissions(ctx context.Context) ([]PermissionDefinition, error) {
	if err := u.ready(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	catalog := domain.PermissionCatalog()
	out := make([]PermissionDefinition, 0, len(catalog))
	for _, permission := range catalog {
		out = append(out, PermissionDefinition{
			Token:    permission.Token,
			Resource: permission.Resource,
			Action:   permission.Action,
			Name:     permission.Name,
		})
	}
	return out, nil
}

// ListRoles returns paginated roles visible to the active role.
func (u *Usecase) ListRoles(ctx context.Context, input ListInput) (RoleListOutput, error) {
	if err := u.ready(); err != nil {
		return RoleListOutput{}, err
	}
	filter, err := normalizeListInput(input)
	if err != nil {
		return RoleListOutput{}, err
	}
	scope, err := u.roleScope(ctx)
	if err != nil {
		return RoleListOutput{}, err
	}
	roles := scope.visibleRoles()
	pageRoles := paginateRoles(roles, filter)
	return RoleListOutput{
		Items:    mapRoles(pageRoles),
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Total:    len(roles),
	}, nil
}

// CreateRole validates and stores a delegated role.
func (u *Usecase) CreateRole(ctx context.Context, input RoleInput) (Role, error) {
	if err := u.ready(); err != nil {
		return Role{}, err
	}
	scope, err := u.roleScope(ctx)
	if err != nil {
		return Role{}, err
	}
	if checkErr := scope.ensureParentAllowed(input.ParentID, 0); checkErr != nil {
		return Role{}, checkErr
	}
	if checkErr := scope.ensureGrantSubset(input.Permissions, input.MenuIDs); checkErr != nil {
		return Role{}, checkErr
	}
	role, err := domain.RestoreRole(0, input.ParentID, input.Code, input.Name, input.Permissions, input.MenuIDs, input.DefaultPath, input.Active, time.Time{}, time.Time{})
	if err != nil {
		return Role{}, mapDomainError(err)
	}
	created, err := u.store.CreateRole(ctx, role)
	if err != nil {
		return Role{}, err
	}
	return fromRole(created), nil
}

// UpdateRole applies mutable role changes within the active role delegation scope.
func (u *Usecase) UpdateRole(ctx context.Context, input UpdateRoleInput) (Role, error) {
	if err := u.ready(); err != nil {
		return Role{}, err
	}
	scope, err := u.roleScope(ctx)
	if err != nil {
		return Role{}, err
	}
	existing, err := u.store.FindRoleByID(ctx, input.ID)
	if err != nil {
		return Role{}, err
	}
	if checkErr := scope.ensureRoleMutable(existing); checkErr != nil {
		return Role{}, checkErr
	}
	draft := roleUpdateDraftFrom(existing, input)
	if checkErr := scope.ensureParentAllowed(draft.parentID, existing.ID); checkErr != nil {
		return Role{}, checkErr
	}
	if roleParentWouldCycle(scope.allRoles, existing.ID, draft.parentID) {
		return Role{}, apperr.NewBadRequest("invalid role parent")
	}
	if existing.IsSuperAdmin() && !draft.active {
		return Role{}, apperr.NewBadRequest("super admin role must stay active")
	}
	if checkErr := scope.ensureGrantSubset(draft.permissions, draft.menuIDs); checkErr != nil {
		return Role{}, checkErr
	}
	role, err := domain.RestoreRole(
		existing.ID,
		draft.parentID,
		existing.Code,
		draft.name,
		draft.permissions,
		draft.menuIDs,
		draft.defaultPath,
		draft.active,
		existing.CreatedAt,
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

// DeleteRole removes a role only when the active role owns it and no admin or child role still depends on it.
func (u *Usecase) DeleteRole(ctx context.Context, id int64) error {
	if err := u.ready(); err != nil {
		return err
	}
	scope, err := u.roleScope(ctx)
	if err != nil {
		return err
	}
	existing, err := u.store.FindRoleByID(ctx, id)
	if err != nil {
		return err
	}
	if checkErr := scope.ensureRoleMutable(existing); checkErr != nil {
		return checkErr
	}
	if existing.IsSuperAdmin() {
		return apperr.NewBadRequest("super admin role cannot be deleted")
	}
	if roleHasChildren(scope.allRoles, existing.ID) {
		return apperr.NewConflict("role has child roles")
	}
	assigned, err := u.admins.RoleAssigned(ctx, existing.ID)
	if err != nil {
		return err
	}
	if assigned {
		return apperr.NewConflict("role is assigned to admins")
	}
	return u.store.DeleteRole(ctx, existing.ID)
}

// CopyRole creates a new role by copying grants from an existing role inside the active delegation scope.
func (u *Usecase) CopyRole(ctx context.Context, input CopyRoleInput) (Role, error) {
	if err := u.ready(); err != nil {
		return Role{}, err
	}
	scope, err := u.roleScope(ctx)
	if err != nil {
		return Role{}, err
	}
	source, err := u.store.FindRoleByID(ctx, input.SourceID)
	if err != nil {
		return Role{}, err
	}
	if checkErr := scope.ensureRoleMutable(source); checkErr != nil {
		return Role{}, checkErr
	}
	if source.IsSuperAdmin() {
		return Role{}, apperr.NewBadRequest("super admin role cannot be copied")
	}
	draft := copyRoleDraftFrom(source, input)
	if checkErr := scope.ensureParentAllowed(draft.parentID, 0); checkErr != nil {
		return Role{}, checkErr
	}
	if checkErr := scope.ensureGrantSubset(source.Permissions, source.MenuIDs); checkErr != nil {
		return Role{}, checkErr
	}
	role, err := domain.RestoreRole(0, draft.parentID, input.Code, input.Name, source.Permissions, source.MenuIDs, draft.defaultPath, draft.active, time.Time{}, time.Time{})
	if err != nil {
		return Role{}, mapDomainError(err)
	}
	created, err := u.store.CreateRole(ctx, role)
	if err != nil {
		return Role{}, err
	}
	return fromRole(created), nil
}

type roleUpdateDraft struct {
	parentID    int64
	name        string
	permissions []string
	menuIDs     []int64
	defaultPath string
	active      bool
}

func roleUpdateDraftFrom(existing domain.Role, input UpdateRoleInput) roleUpdateDraft {
	draft := roleUpdateDraft{
		parentID:    existing.ParentID,
		name:        existing.Name,
		permissions: coalesceStrings(input.Permissions, existing.Permissions),
		menuIDs:     coalesceIDs(input.MenuIDs, existing.MenuIDs),
		defaultPath: existing.DefaultPath,
		active:      existing.Active,
	}
	if input.ParentID != nil {
		draft.parentID = *input.ParentID
	}
	if input.Name != nil {
		draft.name = *input.Name
	}
	if input.DefaultPath != nil {
		draft.defaultPath = *input.DefaultPath
	}
	if input.Active != nil {
		draft.active = *input.Active
	}
	return draft
}

type copyRoleDraft struct {
	parentID    int64
	defaultPath string
	active      bool
}

func copyRoleDraftFrom(source domain.Role, input CopyRoleInput) copyRoleDraft {
	draft := copyRoleDraft{
		parentID:    source.ParentID,
		defaultPath: source.DefaultPath,
		active:      source.Active,
	}
	if input.ParentID != nil {
		draft.parentID = *input.ParentID
	}
	if input.DefaultPath != nil {
		draft.defaultPath = *input.DefaultPath
	}
	if input.Active != nil {
		draft.active = *input.Active
	}
	return draft
}

// AssignableRoleIDs returns role ids the active role may assign to administrators.
func (u *Usecase) AssignableRoleIDs(ctx context.Context) ([]int64, error) {
	if err := u.ready(); err != nil {
		return nil, err
	}
	scope, err := u.roleScope(ctx)
	if err != nil {
		return nil, err
	}
	return sortedRoleIDs(scope.assignableRoleIDs), nil
}

// EnsureAssignableRoles rejects role assignments outside the active role delegation scope.
func (u *Usecase) EnsureAssignableRoles(ctx context.Context, roleIDs []int64) error {
	if err := u.ready(); err != nil {
		return err
	}
	scope, err := u.roleScope(ctx)
	if err != nil {
		return err
	}
	for _, roleID := range uniquePositiveIDs(roleIDs) {
		if _, ok := scope.assignableRoleIDs[roleID]; !ok {
			return apperr.NewPermissionDenied("role", strconv.FormatInt(roleID, 10))
		}
	}
	return nil
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
	menu, err := domain.RestoreMenu(input.ID, input.ParentID, input.Name, input.Path, input.Icon, input.Permission, input.Sort, input.Active, existing.CreatedAt, time.Time{})
	if err != nil {
		return Menu{}, mapDomainError(err)
	}
	updated, err := u.store.UpdateMenu(ctx, menu)
	if err != nil {
		return Menu{}, err
	}
	return fromMenu(updated), nil
}

// DeleteMenu removes a menu only when no child menu or role grant still references it.
func (u *Usecase) DeleteMenu(ctx context.Context, id int64) error {
	if err := u.ready(); err != nil {
		return err
	}
	existing, err := u.store.FindMenuByID(ctx, id)
	if err != nil {
		return err
	}
	menus, err := u.store.ListMenus(ctx)
	if err != nil {
		return err
	}
	if menuHasChildren(menus, existing.ID) {
		return apperr.NewConflict("menu has child menus")
	}
	roles, err := u.store.ListAllRoles(ctx)
	if err != nil {
		return err
	}
	if menuAssignedToRole(roles, existing.ID) {
		return apperr.NewConflict("menu is assigned to roles")
	}
	return u.store.DeleteMenu(ctx, existing.ID)
}

func (u *Usecase) ready() error {
	if u == nil || u.store == nil || u.admins == nil {
		return apperr.New(apperr.ErrInternalServer, "access dependencies are not configured")
	}
	return nil
}

func (u *Usecase) roleScope(ctx context.Context) (roleScope, error) {
	adminID, err := currentAdminID(ctx)
	if err != nil {
		return roleScope{}, err
	}
	state, err := u.admins.AdminRoleState(ctx, adminID)
	if err != nil {
		return roleScope{}, err
	}
	activeRoleID := state.ActiveRoleID
	if roleID, parseErr := currentRoleID(ctx); parseErr == nil && roleID > 0 {
		activeRoleID = roleID
	}
	if !containsID(state.RoleIDs, activeRoleID) {
		return roleScope{}, apperr.NewUnauthorized()
	}
	roles, err := u.store.ListAllRoles(ctx)
	if err != nil {
		return roleScope{}, err
	}
	activeRole, ok := findRole(roles, activeRoleID)
	if !ok || !activeRole.Active {
		return roleScope{}, apperr.NewPermissionDenied("role", strconv.FormatInt(activeRoleID, 10))
	}
	allRoleIDs := roleIDSet(roles)
	if activeRole.IsSuperAdmin() {
		return roleScope{
			activeRole:        activeRole,
			allRoles:          roles,
			visibleRoleIDs:    allRoleIDs,
			assignableRoleIDs: allRoleIDs,
			super:             true,
		}, nil
	}
	descendants := descendantRoleIDs(roles, activeRole.ID)
	visible := copyIDSet(descendants)
	visible[activeRole.ID] = struct{}{}
	return roleScope{
		activeRole:        activeRole,
		allRoles:          roles,
		visibleRoleIDs:    visible,
		assignableRoleIDs: descendants,
	}, nil
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

func paginateRoles(roles []domain.Role, filter ListFilter) []domain.Role {
	if filter.Offset >= len(roles) {
		return []domain.Role{}
	}
	end := filter.Offset + filter.Limit
	if end > len(roles) {
		end = len(roles)
	}
	return roles[filter.Offset:end]
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
		ID:          role.ID,
		ParentID:    role.ParentID,
		Code:        role.Code,
		Name:        role.Name,
		Permissions: role.Permissions,
		MenuIDs:     role.MenuIDs,
		DefaultPath: role.DefaultPath,
		Active:      role.Active,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func fromMenu(menu domain.Menu) Menu {
	return Menu{
		ID:         menu.ID,
		ParentID:   menu.ParentID,
		Name:       menu.Name,
		Path:       menu.Path,
		Icon:       menu.Icon,
		Permission: menu.Permission,
		Sort:       menu.Sort,
		Active:     menu.Active,
		CreatedAt:  menu.CreatedAt,
		UpdatedAt:  menu.UpdatedAt,
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

func currentAdminID(ctx context.Context) (int64, error) {
	raw := requestctx.GetUserID(ctx)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperr.NewUnauthorized()
	}
	return id, nil
}

func currentRoleID(ctx context.Context) (int64, error) {
	raw := requestctx.GetRoleID(ctx)
	if raw == "" {
		return 0, apperr.NewUnauthorized()
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperr.NewUnauthorized()
	}
	return id, nil
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
	{domain.ErrInvalidRoleParent, "invalid role parent"},
	{domain.ErrInvalidRoleCode, "invalid role code"},
	{domain.ErrInvalidRoleName, "invalid role name"},
	{domain.ErrRoleNeedsPerms, "role needs permissions"},
	{domain.ErrInvalidDefaultPath, "invalid default path"},
	{domain.ErrInvalidMenuID, "invalid menu id"},
	{domain.ErrInvalidMenuName, "invalid menu name"},
	{domain.ErrInvalidMenuPath, "invalid menu path"},
	{domain.ErrInvalidPermission, "invalid permission"},
}

type roleScope struct {
	activeRole        domain.Role
	allRoles          []domain.Role
	visibleRoleIDs    map[int64]struct{}
	assignableRoleIDs map[int64]struct{}
	super             bool
}

func (s roleScope) visibleRoles() []domain.Role {
	out := make([]domain.Role, 0, len(s.visibleRoleIDs))
	for _, role := range s.allRoles {
		if _, ok := s.visibleRoleIDs[role.ID]; ok {
			out = append(out, role)
		}
	}
	return out
}

func (s roleScope) ensureRoleMutable(role domain.Role) error {
	if s.super {
		return nil
	}
	if role.IsSuperAdmin() {
		return apperr.NewPermissionDenied("role", role.Code)
	}
	if _, ok := s.assignableRoleIDs[role.ID]; !ok {
		return apperr.NewPermissionDenied("role", strconv.FormatInt(role.ID, 10))
	}
	return nil
}

func (s roleScope) ensureParentAllowed(parentID, roleID int64) error {
	if s.super {
		if parentID == roleID && roleID > 0 {
			return apperr.NewBadRequest("invalid role parent")
		}
		return nil
	}
	if parentID <= 0 {
		return apperr.NewPermissionDenied("role", "root")
	}
	if parentID == roleID && roleID > 0 {
		return apperr.NewBadRequest("invalid role parent")
	}
	if parentID == s.activeRole.ID {
		return nil
	}
	if _, ok := s.assignableRoleIDs[parentID]; ok {
		return nil
	}
	return apperr.NewPermissionDenied("role", strconv.FormatInt(parentID, 10))
}

func (s roleScope) ensureGrantSubset(permissions []string, menuIDs []int64) error {
	if s.super {
		return nil
	}
	if !isStringSubset(permissions, s.activeRole.Permissions) {
		return apperr.NewPermissionDenied("permission", "grant")
	}
	if !isInt64Subset(menuIDs, s.activeRole.MenuIDs) {
		return apperr.NewPermissionDenied("menu", "grant")
	}
	return nil
}

func roleParentWouldCycle(roles []domain.Role, roleID, parentID int64) bool {
	for parentID > 0 {
		if parentID == roleID {
			return true
		}
		parent, ok := findRole(roles, parentID)
		if !ok {
			return false
		}
		parentID = parent.ParentID
	}
	return false
}

func roleHasChildren(roles []domain.Role, roleID int64) bool {
	for _, role := range roles {
		if role.ParentID == roleID {
			return true
		}
	}
	return false
}

func menuHasChildren(menus []domain.Menu, menuID int64) bool {
	for _, menu := range menus {
		if menu.ParentID == menuID {
			return true
		}
	}
	return false
}

func menuAssignedToRole(roles []domain.Role, menuID int64) bool {
	for _, role := range roles {
		if containsID(role.MenuIDs, menuID) {
			return true
		}
	}
	return false
}

func findRole(roles []domain.Role, roleID int64) (domain.Role, bool) {
	for _, role := range roles {
		if role.ID == roleID {
			return role, true
		}
	}
	return domain.Role{}, false
}

func roleIDSet(roles []domain.Role) map[int64]struct{} {
	out := make(map[int64]struct{}, len(roles))
	for _, role := range roles {
		out[role.ID] = struct{}{}
	}
	return out
}

func descendantRoleIDs(roles []domain.Role, rootID int64) map[int64]struct{} {
	children := make(map[int64][]int64, len(roles))
	for _, role := range roles {
		children[role.ParentID] = append(children[role.ParentID], role.ID)
	}
	out := map[int64]struct{}{}
	var walk func(int64)
	walk = func(parentID int64) {
		for _, childID := range children[parentID] {
			out[childID] = struct{}{}
			walk(childID)
		}
	}
	walk(rootID)
	return out
}

func copyIDSet(values map[int64]struct{}) map[int64]struct{} {
	out := make(map[int64]struct{}, len(values))
	for value := range values {
		out[value] = struct{}{}
	}
	return out
}

func sortedRoleIDs(values map[int64]struct{}) []int64 {
	out := make([]int64, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
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

func containsID(ids []int64, want int64) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

func isStringSubset(values, allowed []string) bool {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, value := range allowed {
		allowedSet[value] = struct{}{}
	}
	for _, value := range values {
		if _, ok := allowedSet[value]; !ok {
			return false
		}
	}
	return true
}

func isInt64Subset(values, allowed []int64) bool {
	allowedSet := make(map[int64]struct{}, len(allowed))
	for _, value := range allowed {
		allowedSet[value] = struct{}{}
	}
	for _, value := range values {
		if _, ok := allowedSet[value]; !ok {
			return false
		}
	}
	return true
}
