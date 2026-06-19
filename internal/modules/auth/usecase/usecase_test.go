package usecase_test

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	authusecase "github.com/NSObjects/echo-admin/internal/modules/auth/usecase"
	identitydomain "github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
)

func TestLoginReturnsTokenAndCurrentUserGrants(t *testing.T) {
	uc, recorder := newUsecase(t)
	output, err := uc.Login(context.Background(), authusecase.LoginInput{
		Username: "admin",
		Password: "admin123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if output.Token == "" {
		t.Fatal("Login() token is empty, want signed token")
	}
	if output.User.ActiveRoleID != 1 {
		t.Fatalf("Login().User.ActiveRoleID = %d, want 1", output.User.ActiveRoleID)
	}
	if len(recorder.records) != 1 || !recorder.records[0].Success {
		t.Fatalf("login records = %#v, want one success record", recorder.records)
	}

	ctx := requestctx.WithUserID(context.Background(), "1")
	current, err := uc.CurrentUser(ctx)
	if err != nil {
		t.Fatalf("CurrentUser() error = %v", err)
	}
	if current.Username != "admin" {
		t.Fatalf("CurrentUser().Username = %q, want admin", current.Username)
	}
	if !contains(current.Permissions, accessdomain.PermissionAdminRead) {
		t.Fatalf("permissions = %v, want %s", current.Permissions, accessdomain.PermissionAdminRead)
	}
	if len(current.Menus) == 0 {
		t.Fatal("CurrentUser().Menus is empty, want visible menus")
	}
}

func TestFailedLoginRecordsLoginLog(t *testing.T) {
	uc, recorder := newUsecase(t)
	_, err := uc.Login(context.Background(), authusecase.LoginInput{
		Username: "admin",
		Password: "wrong-password",
	})
	if err == nil {
		t.Fatal("Login() error = nil, want invalid credentials")
	}
	if len(recorder.records) != 1 {
		t.Fatalf("login record count = %d, want 1", len(recorder.records))
	}
	if recorder.records[0].Success {
		t.Fatal("login record success = true, want false")
	}
}

func TestRequirePermissionUsesCasbinRBAC(t *testing.T) {
	uc, _ := newUsecase(t)
	ctx := requestctx.WithUserID(context.Background(), "1")

	if err := uc.RequirePermission(ctx, accessdomain.PermissionAdminRead); err != nil {
		t.Fatalf("RequirePermission(allowed) error = %v", err)
	}
	if err := uc.RequirePermission(ctx, accessdomain.PermissionRoleRead); err == nil {
		t.Fatal("RequirePermission(denied) error = nil, want permission denied")
	}
}

func TestSwitchRolePersistsActiveRoleAndScopesGrants(t *testing.T) {
	uc, _ := newUsecase(t)
	ctx := requestctx.WithUserID(context.Background(), "1")

	output, err := uc.SwitchRole(ctx, authusecase.RoleSwitchInput{RoleID: 2})
	if err != nil {
		t.Fatalf("SwitchRole() error = %v", err)
	}
	if output.Token == "" {
		t.Fatal("SwitchRole() token is empty, want signed token")
	}
	if output.User.ActiveRoleID != 2 {
		t.Fatalf("SwitchRole().User.ActiveRoleID = %d, want 2", output.User.ActiveRoleID)
	}
	if !contains(output.User.Permissions, accessdomain.PermissionRoleRead) {
		t.Fatalf("permissions = %v, want %s", output.User.Permissions, accessdomain.PermissionRoleRead)
	}
	if contains(output.User.Permissions, accessdomain.PermissionAdminRead) {
		t.Fatalf("permissions = %v, want active role only", output.User.Permissions)
	}

	switchedCtx := requestctx.WithRoleID(ctx, "2")
	if err := uc.RequirePermission(switchedCtx, accessdomain.PermissionRoleRead); err != nil {
		t.Fatalf("RequirePermission(switched allowed) error = %v", err)
	}
	if err := uc.RequirePermission(switchedCtx, accessdomain.PermissionAdminRead); err == nil {
		t.Fatal("RequirePermission(old role permission) error = nil, want permission denied")
	}
}

func newUsecase(t *testing.T) (*authusecase.Usecase, *loginRecorder) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	now := time.Unix(1_800_000_000, 0).UTC()
	admin, err := identitydomain.RestoreAdmin(1, "admin", "系统管理员", "admin@example.com", hash, []int64{1, 2}, 1, true, now, now)
	if err != nil {
		t.Fatalf("RestoreAdmin() error = %v", err)
	}
	superRole, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", []string{accessdomain.PermissionAdminRead}, []int64{1}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole() error = %v", err)
	}
	operatorRole, err := accessdomain.RestoreRole(2, 1, "operator", "运营", []string{accessdomain.PermissionRoleRead}, []int64{2}, "/roles", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(operator) error = %v", err)
	}
	adminMenu, err := accessdomain.RestoreMenu(1, 0, "管理员管理", "/admins", "user", accessdomain.PermissionAdminRead, 10, true, now, now)
	if err != nil {
		t.Fatalf("RestoreMenu() error = %v", err)
	}
	roleMenu, err := accessdomain.RestoreMenu(2, 0, "角色权限", "/roles", "safety", accessdomain.PermissionRoleRead, 20, true, now, now)
	if err != nil {
		t.Fatalf("RestoreMenu(role) error = %v", err)
	}
	store := &authStore{
		admin: admin,
		roles: map[int64]accessdomain.Role{
			1: superRole,
			2: operatorRole,
		},
		menus: []accessdomain.Menu{adminMenu, roleMenu},
	}
	recorder := &loginRecorder{}
	uc := authusecase.New(store, store, store, recorder, "test-secret", authusecase.WithClock(func() time.Time {
		return now
	}))
	return uc, recorder
}

type authStore struct {
	admin identitydomain.Admin
	roles map[int64]accessdomain.Role
	menus []accessdomain.Menu
}

func (s *authStore) FindByUsername(context.Context, string) (identitydomain.Admin, error) {
	return s.admin, nil
}

func (s *authStore) FindByID(context.Context, int64) (identitydomain.Admin, error) {
	return s.admin, nil
}

func (s *authStore) Update(_ context.Context, admin identitydomain.Admin) (identitydomain.Admin, error) {
	s.admin = admin
	return admin, nil
}

func (s *authStore) FindRoleByID(_ context.Context, id int64) (accessdomain.Role, error) {
	role, ok := s.roles[id]
	if !ok {
		return accessdomain.Role{}, accessdomain.ErrInvalidRoleID
	}
	return role, nil
}

func (s *authStore) ListMenus(context.Context) ([]accessdomain.Menu, error) {
	return s.menus, nil
}

type loginRecorder struct {
	records []authusecase.LoginRecord
}

func (r *loginRecorder) RecordLogin(_ context.Context, record authusecase.LoginRecord) error {
	r.records = append(r.records, record)
	return nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
