package usecase_test

import (
	"context"
	"testing"
	"time"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	"github.com/NSObjects/echo-admin/internal/modules/access/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
)

func TestUpdateMenuPreservesCreatedAt(t *testing.T) {
	createdAt := time.Unix(1_700_000_000, 0).UTC()
	store := &storeSpy{menuCreatedAt: createdAt}
	uc := usecase.New(store, adminRoleReaderSpy{})

	_, err := uc.UpdateMenu(context.Background(), usecase.UpdateMenuInput{
		ID:         1,
		ParentID:   0,
		Name:       "Updated Menu",
		Path:       "/updated",
		Icon:       "menu",
		Permission: accessdomain.PermissionLogRead,
		Sort:       20,
		Active:     true,
	})
	if err != nil {
		t.Fatalf("UpdateMenu() error = %v", err)
	}
	if got := store.updatedMenu.CreatedAt; !got.Equal(createdAt) {
		t.Fatalf("updated CreatedAt = %s, want %s", got, createdAt)
	}
}

func TestCreateRoleEnforcesActiveRoleGrantScope(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	manager, err := accessdomain.RestoreRole(2, 1, "manager", "经理", []string{accessdomain.PermissionAdminRead}, []int64{1}, "/admins", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(manager) error = %v", err)
	}
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1, 2}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	store := &storeSpy{roles: []accessdomain.Role{root, manager}}
	uc := usecase.New(store, adminRoleReaderSpy{
		state: usecase.AdminRoleState{RoleIDs: []int64{2}, ActiveRoleID: 2},
	})
	ctx := requestctx.WithUserID(context.Background(), "42")
	ctx = requestctx.WithRoleID(ctx, "2")

	created, err := uc.CreateRole(ctx, usecase.RoleInput{
		ParentID:    2,
		Code:        "operator",
		Name:        "运营",
		Permissions: []string{accessdomain.PermissionAdminRead},
		MenuIDs:     []int64{1},
		DefaultPath: "/admins",
		Active:      true,
	})
	if err != nil {
		t.Fatalf("CreateRole(allowed) error = %v", err)
	}
	if created.ParentID != 2 {
		t.Fatalf("created ParentID = %d, want 2", created.ParentID)
	}

	_, err = uc.CreateRole(ctx, usecase.RoleInput{
		ParentID:    2,
		Code:        "bad_operator",
		Name:        "越权运营",
		Permissions: []string{accessdomain.PermissionRoleRead},
		MenuIDs:     []int64{1},
		DefaultPath: "/roles",
		Active:      true,
	})
	if err == nil {
		t.Fatal("CreateRole(extra permission) error = nil, want permission denied")
	}
}

func TestCopyRoleCopiesSourceGrants(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1, 2}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	source, err := accessdomain.RestoreRole(2, 1, "operator", "运营", []string{accessdomain.PermissionAdminRead, accessdomain.PermissionLogRead}, []int64{1, 2}, "/admins", false, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(source) error = %v", err)
	}
	store := &storeSpy{roles: []accessdomain.Role{root, source}}
	uc := usecase.New(store, adminRoleReaderSpy{})

	copied, err := uc.CopyRole(superAdminContext(), usecase.CopyRoleInput{
		SourceID: 2,
		Code:     "operator_copy",
		Name:     "运营副本",
	})
	if err != nil {
		t.Fatalf("CopyRole() error = %v", err)
	}
	if copied.Code != "operator_copy" {
		t.Fatalf("copied Code = %q, want operator_copy", copied.Code)
	}
	if store.createdRole.ParentID != source.ParentID {
		t.Fatalf("created ParentID = %d, want %d", store.createdRole.ParentID, source.ParentID)
	}
	if store.createdRole.DefaultPath != source.DefaultPath {
		t.Fatalf("created DefaultPath = %q, want %q", store.createdRole.DefaultPath, source.DefaultPath)
	}
	if store.createdRole.Active != source.Active {
		t.Fatalf("created Active = %v, want %v", store.createdRole.Active, source.Active)
	}
	if got, want := store.createdRole.Permissions, source.Permissions; !sameStrings(got, want) {
		t.Fatalf("created Permissions = %v, want %v", got, want)
	}
	if got, want := store.createdRole.MenuIDs, source.MenuIDs; !sameInt64s(got, want) {
		t.Fatalf("created MenuIDs = %v, want %v", got, want)
	}
}

func TestDeleteRoleRejectsAssignedRole(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	operator, err := accessdomain.RestoreRole(2, 1, "operator", "运营", []string{accessdomain.PermissionAdminRead}, []int64{1}, "/admins", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(operator) error = %v", err)
	}
	store := &storeSpy{roles: []accessdomain.Role{root, operator}}
	uc := usecase.New(store, adminRoleReaderSpy{assignedRoles: map[int64]bool{2: true}})

	err = uc.DeleteRole(superAdminContext(), 2)
	if err == nil {
		t.Fatal("DeleteRole(assigned role) error = nil, want conflict")
	}
	if store.deletedRoleID != 0 {
		t.Fatalf("deletedRoleID = %d, want 0", store.deletedRoleID)
	}
}

func TestDeleteRoleDeletesLeafRole(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	operator, err := accessdomain.RestoreRole(2, 1, "operator", "运营", []string{accessdomain.PermissionAdminRead}, []int64{1}, "/admins", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(operator) error = %v", err)
	}
	store := &storeSpy{roles: []accessdomain.Role{root, operator}}
	uc := usecase.New(store, adminRoleReaderSpy{})

	if err := uc.DeleteRole(superAdminContext(), 2); err != nil {
		t.Fatalf("DeleteRole() error = %v", err)
	}
	if store.deletedRoleID != 2 {
		t.Fatalf("deletedRoleID = %d, want 2", store.deletedRoleID)
	}
}

func TestDeleteMenuRejectsAssignedMenu(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	role, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{2}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(role) error = %v", err)
	}
	menu, err := accessdomain.RestoreMenu(2, 0, "管理员管理", "/admins", "user", accessdomain.PermissionAdminRead, 20, true, now, now)
	if err != nil {
		t.Fatalf("RestoreMenu(menu) error = %v", err)
	}
	store := &storeSpy{roles: []accessdomain.Role{role}, menus: []accessdomain.Menu{menu}}
	uc := usecase.New(store, adminRoleReaderSpy{})

	err = uc.DeleteMenu(context.Background(), 2)
	if err == nil {
		t.Fatal("DeleteMenu(assigned menu) error = nil, want conflict")
	}
	if store.deletedMenuID != 0 {
		t.Fatalf("deletedMenuID = %d, want 0", store.deletedMenuID)
	}
}

func TestDeleteMenuDeletesUnassignedLeafMenu(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	menu, err := accessdomain.RestoreMenu(2, 0, "临时菜单", "/scratch", "menu", "", 20, true, now, now)
	if err != nil {
		t.Fatalf("RestoreMenu(menu) error = %v", err)
	}
	store := &storeSpy{menus: []accessdomain.Menu{menu}}
	uc := usecase.New(store, adminRoleReaderSpy{})

	if err := uc.DeleteMenu(context.Background(), 2); err != nil {
		t.Fatalf("DeleteMenu() error = %v", err)
	}
	if store.deletedMenuID != 2 {
		t.Fatalf("deletedMenuID = %d, want 2", store.deletedMenuID)
	}
}

func superAdminContext() context.Context {
	ctx := requestctx.WithUserID(context.Background(), "42")
	return requestctx.WithRoleID(ctx, "1")
}

type storeSpy struct {
	menuCreatedAt time.Time
	updatedMenu   accessdomain.Menu
	createdRole   accessdomain.Role
	roles         []accessdomain.Role
	menus         []accessdomain.Menu
	deletedRoleID int64
	deletedMenuID int64
}

func (s *storeSpy) FindRoleByID(_ context.Context, id int64) (accessdomain.Role, error) {
	for _, role := range s.roles {
		if role.ID == id {
			return role, nil
		}
	}
	return accessdomain.Role{}, apperr.NewNotFound("role")
}

func (s *storeSpy) FindRoleByCode(context.Context, string) (accessdomain.Role, error) {
	return accessdomain.Role{}, apperr.NewNotFound("role")
}

func (s *storeSpy) ListAllRoles(context.Context) ([]accessdomain.Role, error) {
	return s.roles, nil
}

func (s *storeSpy) CreateRole(_ context.Context, role accessdomain.Role) (accessdomain.Role, error) {
	s.createdRole = role
	return accessdomain.RestoreRole(10, role.ParentID, role.Code, role.Name, role.Permissions, role.MenuIDs, role.DefaultPath, role.Active, time.Now(), time.Now())
}

func (s *storeSpy) UpdateRole(context.Context, accessdomain.Role) (accessdomain.Role, error) {
	return accessdomain.Role{}, nil
}

func (s *storeSpy) DeleteRole(_ context.Context, id int64) error {
	s.deletedRoleID = id
	return nil
}

func (s *storeSpy) FindMenuByID(_ context.Context, id int64) (accessdomain.Menu, error) {
	for _, menu := range s.menus {
		if menu.ID == id {
			return menu, nil
		}
	}
	return accessdomain.RestoreMenu(1, 0, "Existing", "/existing", "menu", accessdomain.PermissionLogRead, 10, true, s.menuCreatedAt, s.menuCreatedAt)
}

func (s *storeSpy) ListMenus(context.Context) ([]accessdomain.Menu, error) {
	return s.menus, nil
}

func (s *storeSpy) CreateMenu(context.Context, accessdomain.Menu) (accessdomain.Menu, error) {
	return accessdomain.Menu{}, nil
}

func (s *storeSpy) UpdateMenu(_ context.Context, menu accessdomain.Menu) (accessdomain.Menu, error) {
	s.updatedMenu = menu
	return menu, nil
}

func (s *storeSpy) DeleteMenu(_ context.Context, id int64) error {
	s.deletedMenuID = id
	return nil
}

type adminRoleReaderSpy struct {
	state         usecase.AdminRoleState
	assignedRoles map[int64]bool
}

func (s adminRoleReaderSpy) AdminRoleState(context.Context, int64) (usecase.AdminRoleState, error) {
	if len(s.state.RoleIDs) > 0 {
		return s.state, nil
	}
	return usecase.AdminRoleState{RoleIDs: []int64{1}, ActiveRoleID: 1}, nil
}

func (s adminRoleReaderSpy) RoleAssigned(_ context.Context, roleID int64) (bool, error) {
	return s.assignedRoles[roleID], nil
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func sameInt64s(got, want []int64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
