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
		Component:  "./Updated",
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
	uc, ctx := scopedManagerUsecase(t)

	created, err := uc.CreateRole(ctx, usecase.RoleInput{
		ParentID:    2,
		Code:        "operator",
		Name:        "运营",
		Permissions: []string{accessdomain.PermissionAdminRead},
		MenuIDs:     []int64{1},
		APIIDs:      []int64{1},
		ButtonIDs:   []int64{1},
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
		APIIDs:      []int64{1},
		ButtonIDs:   []int64{1},
		DefaultPath: "/roles",
		Active:      true,
	})
	if err == nil {
		t.Fatal("CreateRole(extra permission) error = nil, want permission denied")
	}

	_, err = uc.CreateRole(ctx, usecase.RoleInput{
		ParentID:    2,
		Code:        "bad_api_operator",
		Name:        "越权API运营",
		Permissions: []string{accessdomain.PermissionAdminRead},
		MenuIDs:     []int64{1},
		APIIDs:      []int64{2},
		ButtonIDs:   []int64{1},
		DefaultPath: "/admins",
		Active:      true,
	})
	if err == nil {
		t.Fatal("CreateRole(extra api) error = nil, want permission denied")
	}
}

func TestCreateRoleRejectsButtonGrantOutsideActiveRoleScope(t *testing.T) {
	uc, ctx := scopedManagerUsecase(t)

	_, err := uc.CreateRole(ctx, usecase.RoleInput{
		ParentID:    2,
		Code:        "bad_button_operator",
		Name:        "越权按钮运营",
		Permissions: []string{accessdomain.PermissionAdminRead},
		MenuIDs:     []int64{1},
		APIIDs:      []int64{1},
		ButtonIDs:   []int64{2},
		DefaultPath: "/admins",
		Active:      true,
	})
	if err == nil {
		t.Fatal("CreateRole(extra button) error = nil, want permission denied")
	}
}

func scopedManagerUsecase(t *testing.T) (*usecase.Usecase, context.Context) {
	uc, ctx, _ := scopedManagerUsecaseWithStore(t)
	return uc, ctx
}

func scopedManagerUsecaseWithStore(t *testing.T) (*usecase.Usecase, context.Context, *storeSpy) {
	t.Helper()
	now := time.Unix(1_800_000_000, 0).UTC()
	manager, err := accessdomain.RestoreRole(2, 1, "manager", "经理", []string{accessdomain.PermissionAdminRead}, []int64{1}, []int64{1}, []int64{1}, []int64{2}, "/admins", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(manager) error = %v", err)
	}
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1, 2}, []int64{1, 2}, []int64{1, 2}, []int64{1, 2}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	store := &storeSpy{roles: []accessdomain.Role{root, manager}}
	uc := usecase.New(store, adminRoleReaderSpy{
		state: usecase.AdminRoleState{RoleIDs: []int64{2}, ActiveRoleID: 2},
	})
	ctx := requestctx.WithUserID(context.Background(), "42")
	return uc, requestctx.WithRoleID(ctx, "2"), store
}

func TestVisibleRoleIDsUsesDataAuthorityInsideDelegationScope(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1}, []int64{1}, []int64{1}, []int64{1, 2, 3, 4}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	manager, err := accessdomain.RestoreRole(2, 1, "manager", "经理", []string{accessdomain.PermissionAdminRead}, []int64{1}, []int64{1}, []int64{1}, []int64{3, 4}, "/admins", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(manager) error = %v", err)
	}
	operator, err := accessdomain.RestoreRole(3, 2, "operator", "运营", []string{accessdomain.PermissionAdminRead}, []int64{1}, []int64{1}, []int64{1}, []int64{3}, "/admins", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(operator) error = %v", err)
	}
	sibling, err := accessdomain.RestoreRole(4, 1, "sibling", "同级", []string{accessdomain.PermissionAdminRead}, []int64{1}, []int64{1}, []int64{1}, []int64{4}, "/admins", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(sibling) error = %v", err)
	}
	uc := usecase.New(&storeSpy{roles: []accessdomain.Role{root, manager, operator, sibling}}, adminRoleReaderSpy{
		state: usecase.AdminRoleState{RoleIDs: []int64{2}, ActiveRoleID: 2},
	})
	ctx := requestctx.WithRoleID(requestctx.WithUserID(context.Background(), "42"), "2")

	ids, err := uc.VisibleRoleIDs(ctx)
	if err != nil {
		t.Fatalf("VisibleRoleIDs() error = %v", err)
	}
	if !sameInt64s(ids, []int64{3}) {
		t.Fatalf("VisibleRoleIDs() = %v, want [3]", ids)
	}
}

func TestCopyRoleCopiesSourceGrants(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1, 2}, []int64{1, 2}, []int64{1, 2}, []int64{1, 2}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	source, err := accessdomain.RestoreRole(2, 1, "operator", "运营", []string{accessdomain.PermissionAdminRead, accessdomain.PermissionLogRead}, []int64{1, 2}, []int64{1, 2}, []int64{1, 2}, []int64{2}, "/admins", false, now, now)
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
	assertCopiedRole(t, store.createdRole, source)
}

func assertCopiedRole(t *testing.T, gotRole, wantRole accessdomain.Role) {
	t.Helper()
	if gotRole.ParentID != wantRole.ParentID {
		t.Fatalf("created ParentID = %d, want %d", gotRole.ParentID, wantRole.ParentID)
	}
	if gotRole.DefaultPath != wantRole.DefaultPath {
		t.Fatalf("created DefaultPath = %q, want %q", gotRole.DefaultPath, wantRole.DefaultPath)
	}
	if gotRole.Active != wantRole.Active {
		t.Fatalf("created Active = %v, want %v", gotRole.Active, wantRole.Active)
	}
	if got, want := gotRole.Permissions, wantRole.Permissions; !sameStrings(got, want) {
		t.Fatalf("created Permissions = %v, want %v", got, want)
	}
	if got, want := gotRole.MenuIDs, wantRole.MenuIDs; !sameInt64s(got, want) {
		t.Fatalf("created MenuIDs = %v, want %v", got, want)
	}
	if got, want := gotRole.APIIDs, wantRole.APIIDs; !sameInt64s(got, want) {
		t.Fatalf("created APIIDs = %v, want %v", got, want)
	}
	if got, want := gotRole.ButtonIDs, wantRole.ButtonIDs; !sameInt64s(got, want) {
		t.Fatalf("created ButtonIDs = %v, want %v", got, want)
	}
	if got, want := gotRole.DataRoleIDs, wantRole.DataRoleIDs; !sameInt64s(got, want) {
		t.Fatalf("created DataRoleIDs = %v, want %v", got, want)
	}
}

func TestDeleteRoleRejectsAssignedRole(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1}, []int64{1}, []int64{1}, []int64{1, 2}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	operator, err := accessdomain.RestoreRole(2, 1, "operator", "运营", []string{accessdomain.PermissionAdminRead}, []int64{1}, []int64{1}, []int64{1}, []int64{2}, "/admins", true, now, now)
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
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1}, []int64{1}, []int64{1}, []int64{1, 2}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	operator, err := accessdomain.RestoreRole(2, 1, "operator", "运营", []string{accessdomain.PermissionAdminRead}, []int64{1}, []int64{1}, []int64{1}, []int64{2}, "/admins", true, now, now)
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
	role, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{2}, []int64{1}, []int64{1}, []int64{1}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(role) error = %v", err)
	}
	menu, err := accessdomain.RestoreMenu(2, 0, "管理员管理", "/admins", "user", false, "./Admins", accessdomain.MenuMeta{}, accessdomain.PermissionAdminRead, 20, true, nil, now, now)
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

func TestDeleteMenuRejectsAssignedMenuButton(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	role, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), nil, nil, []int64{7}, []int64{1}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(role) error = %v", err)
	}
	button, err := accessdomain.RestoreMenuButton(7, 2, "delete", "删除菜单", now, now)
	if err != nil {
		t.Fatalf("RestoreMenuButton() error = %v", err)
	}
	menu, err := accessdomain.RestoreMenu(2, 0, "菜单管理", "/menus", "menu", false, "./Menus", accessdomain.MenuMeta{}, accessdomain.PermissionMenuRead, 20, true, []accessdomain.MenuButton{button}, now, now)
	if err != nil {
		t.Fatalf("RestoreMenu(menu) error = %v", err)
	}
	store := &storeSpy{roles: []accessdomain.Role{role}, menus: []accessdomain.Menu{menu}}
	uc := usecase.New(store, adminRoleReaderSpy{})

	err = uc.DeleteMenu(context.Background(), 2)
	if err == nil {
		t.Fatal("DeleteMenu(assigned menu button) error = nil, want conflict")
	}
	if store.deletedMenuID != 0 {
		t.Fatalf("deletedMenuID = %d, want 0", store.deletedMenuID)
	}
}

func TestDeleteMenuDeletesUnassignedLeafMenu(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	menu, err := accessdomain.RestoreMenu(2, 0, "临时菜单", "/scratch", "menu", false, "./Scratch", accessdomain.MenuMeta{}, "", 20, true, nil, now, now)
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

func TestSetMenuRolesUpdatesRoleGrants(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1}, []int64{1}, []int64{1}, []int64{1, 2}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	operator, err := accessdomain.RestoreRole(2, 1, "operator", "运营", []string{accessdomain.PermissionAdminRead}, []int64{1}, []int64{1}, []int64{1}, []int64{2}, "/admins", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(operator) error = %v", err)
	}
	menu, err := accessdomain.RestoreMenu(2, 0, "菜单管理", "/menus", "menu", false, "./Menus", accessdomain.MenuMeta{}, accessdomain.PermissionMenuRead, 20, true, nil, now, now)
	if err != nil {
		t.Fatalf("RestoreMenu(menu) error = %v", err)
	}
	store := &storeSpy{roles: []accessdomain.Role{root, operator}, menus: []accessdomain.Menu{menu}}
	uc := usecase.New(store, adminRoleReaderSpy{})

	roleIDs, err := uc.SetMenuRoles(superAdminContext(), usecase.MenuRolesInput{MenuID: 2, RoleIDs: []int64{2}})
	if err != nil {
		t.Fatalf("SetMenuRoles() error = %v", err)
	}
	if !sameInt64s(roleIDs, []int64{2}) {
		t.Fatalf("SetMenuRoles() = %v, want [2]", roleIDs)
	}
	updated, ok := findRoleForTest(store.roles, 2)
	if !ok || !containsIDForTest(updated.MenuIDs, 2) {
		t.Fatalf("updated operator menu ids = %v, want to include 2", updated.MenuIDs)
	}
}

func TestSetAPIRolesRejectsAPIOutsideActiveGrantScope(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1}, []int64{1, 2}, []int64{1}, []int64{1, 2, 3}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	manager, err := accessdomain.RestoreRole(2, 1, "manager", "经理", []string{accessdomain.PermissionAPIUpdate}, []int64{1}, []int64{1}, []int64{1}, []int64{3}, "/admins", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(manager) error = %v", err)
	}
	operator, err := accessdomain.RestoreRole(3, 2, "operator", "运营", []string{accessdomain.PermissionAdminRead}, []int64{1}, []int64{1}, []int64{1}, []int64{3}, "/admins", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(operator) error = %v", err)
	}
	api, err := accessdomain.RestoreAPI(2, "GET", "/api/secret", "Secret", "system", accessdomain.PermissionAPIRead, false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(api) error = %v", err)
	}
	store := &storeSpy{roles: []accessdomain.Role{root, manager, operator}, apis: []accessdomain.API{api}}
	uc := usecase.New(store, adminRoleReaderSpy{state: usecase.AdminRoleState{RoleIDs: []int64{2}, ActiveRoleID: 2}})
	ctx := requestctx.WithRoleID(requestctx.WithUserID(context.Background(), "42"), "2")

	_, err = uc.SetAPIRoles(ctx, usecase.APIRolesInput{APIID: 2, RoleIDs: []int64{3}})
	if err == nil {
		t.Fatal("SetAPIRoles() error = nil, want permission denied")
	}
	if len(store.updatedRoles) != 0 {
		t.Fatalf("updatedRoles = %d, want 0", len(store.updatedRoles))
	}
}

func TestCreateAPIRejectsPublicRouteForScopedRole(t *testing.T) {
	uc, ctx, store := scopedManagerUsecaseWithStore(t)

	_, err := uc.CreateAPI(ctx, usecase.APIInput{
		Method:      "GET",
		Path:        "/api/public-secret",
		Description: "Public secret",
		Group:       "system",
		Permission:  accessdomain.PermissionAPIRead,
		Public:      true,
	})
	if err == nil {
		t.Fatal("CreateAPI(public scoped role) error = nil, want permission denied")
	}
	if store.createdAPI.Path != "" {
		t.Fatalf("createdAPI.Path = %q, want empty", store.createdAPI.Path)
	}
}

func TestUpdateAPIRejectsAPIOutsideActiveGrantScope(t *testing.T) {
	uc, ctx, store := scopedManagerUsecaseWithStore(t)
	store.apis = []accessdomain.API{restoreAPITest(t, 2, "GET", "/api/secret", accessdomain.PermissionAPIRead, false)}

	_, err := uc.UpdateAPI(ctx, usecase.UpdateAPIInput{
		ID:          2,
		Method:      "GET",
		Path:        "/api/secret",
		Description: "Secret",
		Group:       "system",
		Permission:  accessdomain.PermissionAPIRead,
		Public:      false,
	})
	if err == nil {
		t.Fatal("UpdateAPI(outside grant) error = nil, want permission denied")
	}
	if store.createdAPI.Path != "" {
		t.Fatalf("updated API path = %q, want empty", store.createdAPI.Path)
	}
}

func TestUpdateAPIRejectsPublicRouteForScopedRole(t *testing.T) {
	uc, ctx, store := scopedManagerUsecaseWithStore(t)
	store.apis = []accessdomain.API{restoreAPITest(t, 1, "GET", "/api/managed", accessdomain.PermissionAPIRead, false)}

	_, err := uc.UpdateAPI(ctx, usecase.UpdateAPIInput{
		ID:          1,
		Method:      "GET",
		Path:        "/api/managed",
		Description: "Managed",
		Group:       "system",
		Permission:  accessdomain.PermissionAPIRead,
		Public:      true,
	})
	if err == nil {
		t.Fatal("UpdateAPI(public scoped role) error = nil, want permission denied")
	}
	if store.createdAPI.Path != "" {
		t.Fatalf("updated API path = %q, want empty", store.createdAPI.Path)
	}
}

func TestAPIGroupsReturnsSortedUniqueGroups(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	adminAPI, err := accessdomain.RestoreAPI(1, "GET", "/api/admins", "管理员", "admin", accessdomain.PermissionAdminRead, false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(admin) error = %v", err)
	}
	logAPI, err := accessdomain.RestoreAPI(2, "GET", "/api/logs", "日志", "log", accessdomain.PermissionLogRead, false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(log) error = %v", err)
	}
	anotherAdminAPI, err := accessdomain.RestoreAPI(3, "POST", "/api/admins", "创建管理员", "admin", accessdomain.PermissionAdminCreate, false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(another admin) error = %v", err)
	}
	store := &storeSpy{apis: []accessdomain.API{logAPI, adminAPI, anotherAdminAPI}}
	uc := usecase.New(store, adminRoleReaderSpy{})

	groups, err := uc.APIGroups(context.Background())
	if err != nil {
		t.Fatalf("APIGroups() error = %v", err)
	}
	if !sameStrings(groups, []string{"admin", "log"}) {
		t.Fatalf("APIGroups() = %v, want [admin log]", groups)
	}
}

func TestDeleteAPIRejectsAssignedAPI(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	role, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1}, []int64{2}, []int64{1}, []int64{1}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(role) error = %v", err)
	}
	api, err := accessdomain.RestoreAPI(2, "GET", "/api/example", "示例API", "example", accessdomain.PermissionLogRead, false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(api) error = %v", err)
	}
	store := &storeSpy{roles: []accessdomain.Role{role}, apis: []accessdomain.API{api}}
	uc := usecase.New(store, adminRoleReaderSpy{})

	err = uc.DeleteAPI(context.Background(), 2)
	if err == nil {
		t.Fatal("DeleteAPI(assigned api) error = nil, want conflict")
	}
	if store.deletedAPIID != 0 {
		t.Fatalf("deletedAPIID = %d, want 0", store.deletedAPIID)
	}
}

func TestDeleteAPIsRejectsAssignedAPIWithoutPartialDelete(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	role, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), []int64{1}, []int64{3}, []int64{1}, []int64{1}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(role) error = %v", err)
	}
	firstAPI, err := accessdomain.RestoreAPI(2, "GET", "/api/free", "未授权API", "example", accessdomain.PermissionLogRead, false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(first) error = %v", err)
	}
	secondAPI, err := accessdomain.RestoreAPI(3, "GET", "/api/assigned", "已授权API", "example", accessdomain.PermissionLogRead, false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(second) error = %v", err)
	}
	store := &storeSpy{roles: []accessdomain.Role{role}, apis: []accessdomain.API{firstAPI, secondAPI}}
	uc := usecase.New(store, adminRoleReaderSpy{})

	err = uc.DeleteAPIs(context.Background(), []int64{2, 3})
	if err == nil {
		t.Fatal("DeleteAPIs(assigned api) error = nil, want conflict")
	}
	if len(store.deletedAPIIDs) != 0 {
		t.Fatalf("deletedAPIIDs = %v, want none", store.deletedAPIIDs)
	}
}

func TestDeleteAPIDeletesUnassignedAPI(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	api, err := accessdomain.RestoreAPI(2, "GET", "/api/scratch", "临时API", "example", "", false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(api) error = %v", err)
	}
	store := &storeSpy{apis: []accessdomain.API{api}}
	uc := usecase.New(store, adminRoleReaderSpy{})

	if err := uc.DeleteAPI(context.Background(), 2); err != nil {
		t.Fatalf("DeleteAPI() error = %v", err)
	}
	if store.deletedAPIID != 2 {
		t.Fatalf("deletedAPIID = %d, want 2", store.deletedAPIID)
	}
}

func superAdminContext() context.Context {
	ctx := requestctx.WithUserID(context.Background(), "42")
	return requestctx.WithRoleID(ctx, "1")
}

func restoreAPITest(t *testing.T, id int64, method, path, permission string, public bool) accessdomain.API {
	t.Helper()
	api, err := accessdomain.RestoreAPI(id, method, path, "API", "system", permission, public, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("RestoreAPI() error = %v", err)
	}
	return api
}

type storeSpy struct {
	menuCreatedAt time.Time
	updatedMenu   accessdomain.Menu
	createdAPI    accessdomain.API
	createdRole   accessdomain.Role
	updatedRoles  []accessdomain.Role
	roles         []accessdomain.Role
	menus         []accessdomain.Menu
	apis          []accessdomain.API
	deletedRoleID int64
	deletedMenuID int64
	deletedAPIID  int64
	deletedAPIIDs []int64
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
	return accessdomain.RestoreRole(10, role.ParentID, role.Code, role.Name, role.Permissions, role.MenuIDs, role.APIIDs, role.ButtonIDs, role.DataRoleIDs, role.DefaultPath, role.Active, time.Now(), time.Now())
}

func (s *storeSpy) UpdateRole(_ context.Context, role accessdomain.Role) (accessdomain.Role, error) {
	s.updatedRoles = append(s.updatedRoles, role)
	for index, existing := range s.roles {
		if existing.ID == role.ID {
			s.roles[index] = role
			return role, nil
		}
	}
	s.roles = append(s.roles, role)
	return role, nil
}

func (s *storeSpy) DeleteRole(_ context.Context, id int64) error {
	s.deletedRoleID = id
	return nil
}

func (s *storeSpy) FindAPIByID(_ context.Context, id int64) (accessdomain.API, error) {
	for _, api := range s.apis {
		if api.ID == id {
			return api, nil
		}
	}
	return accessdomain.RestoreAPI(id, "GET", "/api/existing", "Existing", "example", accessdomain.PermissionLogRead, false, time.Now(), time.Now())
}

func (s *storeSpy) FindAPIByRoute(_ context.Context, method, path string) (accessdomain.API, error) {
	for _, api := range s.apis {
		if api.Method == method && api.Path == path {
			return api, nil
		}
	}
	return accessdomain.API{}, apperr.NewNotFound("api")
}

func (s *storeSpy) ListAPIs(context.Context) ([]accessdomain.API, error) {
	return s.apis, nil
}

func (s *storeSpy) CreateAPI(_ context.Context, api accessdomain.API) (accessdomain.API, error) {
	s.createdAPI = api
	return accessdomain.RestoreAPI(10, api.Method, api.Path, api.Description, api.Group, api.Permission, api.Public, time.Now(), time.Now())
}

func (s *storeSpy) UpdateAPI(_ context.Context, api accessdomain.API) (accessdomain.API, error) {
	s.createdAPI = api
	return api, nil
}

func (s *storeSpy) DeleteAPI(_ context.Context, id int64) error {
	s.deletedAPIID = id
	s.deletedAPIIDs = append(s.deletedAPIIDs, id)
	return nil
}

func (s *storeSpy) FindMenuByID(_ context.Context, id int64) (accessdomain.Menu, error) {
	for _, menu := range s.menus {
		if menu.ID == id {
			return menu, nil
		}
	}
	return accessdomain.RestoreMenu(1, 0, "Existing", "/existing", "menu", false, "./Existing", accessdomain.MenuMeta{}, accessdomain.PermissionLogRead, 10, true, nil, s.menuCreatedAt, s.menuCreatedAt)
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

func findRoleForTest(roles []accessdomain.Role, roleID int64) (accessdomain.Role, bool) {
	for _, role := range roles {
		if role.ID == roleID {
			return role, true
		}
	}
	return accessdomain.Role{}, false
}

func containsIDForTest(ids []int64, want int64) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
