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

type storeSpy struct {
	menuCreatedAt time.Time
	updatedMenu   accessdomain.Menu
	createdRole   accessdomain.Role
	roles         []accessdomain.Role
}

func (s *storeSpy) FindRoleByID(context.Context, int64) (accessdomain.Role, error) {
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

func (s *storeSpy) FindMenuByID(context.Context, int64) (accessdomain.Menu, error) {
	return accessdomain.RestoreMenu(1, 0, "Existing", "/existing", "menu", accessdomain.PermissionLogRead, 10, true, s.menuCreatedAt, s.menuCreatedAt)
}

func (s *storeSpy) ListMenus(context.Context) ([]accessdomain.Menu, error) {
	return nil, nil
}

func (s *storeSpy) CreateMenu(context.Context, accessdomain.Menu) (accessdomain.Menu, error) {
	return accessdomain.Menu{}, nil
}

func (s *storeSpy) UpdateMenu(_ context.Context, menu accessdomain.Menu) (accessdomain.Menu, error) {
	s.updatedMenu = menu
	return menu, nil
}

type adminRoleReaderSpy struct {
	state usecase.AdminRoleState
}

func (s adminRoleReaderSpy) AdminRoleState(context.Context, int64) (usecase.AdminRoleState, error) {
	if len(s.state.RoleIDs) > 0 {
		return s.state, nil
	}
	return usecase.AdminRoleState{RoleIDs: []int64{1}, ActiveRoleID: 1}, nil
}
