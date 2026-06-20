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
		Password: "123456",
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
	if current.Menus[0].Component != "./Admins" {
		t.Fatalf("CurrentUser().Menus[0].Component = %q, want ./Admins", current.Menus[0].Component)
	}
	if len(current.Menus[0].Buttons) != 2 {
		t.Fatalf("CurrentUser().Menus[0].Buttons = %d, want 2 for super admin", len(current.Menus[0].Buttons))
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

func TestRequireRoutePermissionUsesAssignedAPIs(t *testing.T) {
	uc, _ := newUsecase(t)
	ctx := requestctx.WithRoleID(requestctx.WithUserID(context.Background(), "1"), "2")

	if err := uc.RequireRoutePermission(ctx, accessdomain.PermissionRoleRead, "GET", "/api/roles"); err != nil {
		t.Fatalf("RequireRoutePermission(assigned api) error = %v", err)
	}
	if err := uc.RequireRoutePermission(ctx, accessdomain.PermissionRoleRead, "GET", "/api/roles/:id"); err == nil {
		t.Fatal("RequireRoutePermission(unassigned api) error = nil, want permission denied")
	}
	if err := uc.RequireRoutePermission(ctx, accessdomain.PermissionRoleRead, "GET", "/api/missing"); err == nil {
		t.Fatal("RequireRoutePermission(missing api) error = nil, want permission denied")
	}
}

func TestRequireRoutePermissionKeepsSuperAdminUnblocked(t *testing.T) {
	uc, _ := newUsecase(t)
	ctx := requestctx.WithUserID(context.Background(), "1")

	if err := uc.RequireRoutePermission(ctx, accessdomain.PermissionAdminRead, "DELETE", "/api/admins/:id"); err != nil {
		t.Fatalf("RequireRoutePermission(super admin api not listed on role) error = %v", err)
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
	if len(output.User.Menus) != 1 {
		t.Fatalf("switched menus = %d, want 1", len(output.User.Menus))
	}
	if len(output.User.Menus[0].Buttons) != 1 || output.User.Menus[0].Buttons[0].Name != "update" {
		t.Fatalf("switched menu buttons = %#v, want only update", output.User.Menus[0].Buttons)
	}

	switchedCtx := requestctx.WithRoleID(ctx, "2")
	if err := uc.RequirePermission(switchedCtx, accessdomain.PermissionRoleRead); err != nil {
		t.Fatalf("RequirePermission(switched allowed) error = %v", err)
	}
	if err := uc.RequirePermission(switchedCtx, accessdomain.PermissionAdminRead); err == nil {
		t.Fatal("RequirePermission(old role permission) error = nil, want permission denied")
	}
}

func TestLogoutBlacklistsCurrentJWT(t *testing.T) {
	uc, _ := newUsecase(t)
	output, err := uc.Login(context.Background(), authusecase.LoginInput{
		Username: "admin",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	ctx := requestctx.WithUserID(context.Background(), "1")
	if logoutErr := uc.Logout(ctx, output.Token); logoutErr != nil {
		t.Fatalf("Logout() error = %v", logoutErr)
	}
	blocked, err := uc.TokenBlocked(context.Background(), output.Token)
	if err != nil {
		t.Fatalf("TokenBlocked() error = %v", err)
	}
	if !blocked {
		t.Fatal("TokenBlocked() = false, want true after logout")
	}
}

func TestLogoutRejectsInvalidJWTWithoutWritingBlacklist(t *testing.T) {
	uc, _ := newUsecase(t)

	ctx := requestctx.WithUserID(context.Background(), "1")
	if err := uc.Logout(ctx, "not-a-jwt"); err == nil {
		t.Fatal("Logout(invalid token) error = nil, want unauthorized")
	}
	blocked, err := uc.TokenBlocked(context.Background(), "not-a-jwt")
	if err != nil {
		t.Fatalf("TokenBlocked() error = %v", err)
	}
	if blocked {
		t.Fatal("TokenBlocked(invalid token) = true, want false")
	}
}

func TestChangePasswordRotatesPasswordAndRevokesCurrentToken(t *testing.T) {
	uc, _ := newUsecase(t)
	login, err := uc.Login(context.Background(), authusecase.LoginInput{
		Username: "admin",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	ctx := requestctx.WithUserID(context.Background(), "1")

	err = uc.ChangePassword(ctx, authusecase.ChangePasswordInput{
		CurrentPassword: "123456",
		NewPassword:     "changed123",
		RawToken:        login.Token,
	})
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	blocked, err := uc.TokenBlocked(context.Background(), login.Token)
	if err != nil {
		t.Fatalf("TokenBlocked() error = %v", err)
	}
	if !blocked {
		t.Fatal("TokenBlocked(old token) = false, want true")
	}
	if _, err := uc.Login(context.Background(), authusecase.LoginInput{Username: "admin", Password: "123456"}); err == nil {
		t.Fatal("Login(old password) error = nil, want invalid credentials")
	}
	if _, err := uc.Login(context.Background(), authusecase.LoginInput{Username: "admin", Password: "changed123"}); err != nil {
		t.Fatalf("Login(new password) error = %v", err)
	}
}

func TestChangePasswordRejectsWrongCurrentPassword(t *testing.T) {
	uc, _ := newUsecase(t)
	login, err := uc.Login(context.Background(), authusecase.LoginInput{
		Username: "admin",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	ctx := requestctx.WithUserID(context.Background(), "1")

	err = uc.ChangePassword(ctx, authusecase.ChangePasswordInput{
		CurrentPassword: "wrong-password",
		NewPassword:     "changed123",
		RawToken:        login.Token,
	})
	if err == nil {
		t.Fatal("ChangePassword(wrong current password) error = nil, want unauthorized")
	}
	if _, err := uc.Login(context.Background(), authusecase.LoginInput{Username: "admin", Password: "123456"}); err != nil {
		t.Fatalf("Login(original password) error = %v", err)
	}
}

func newUsecase(t *testing.T) (*authusecase.Usecase, *loginRecorder) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	now := time.Unix(1_800_000_000, 0).UTC()
	admin, err := identitydomain.RestoreAdmin(1, "admin", "系统管理员", "admin@example.com", hash, []int64{1, 2}, 1, true, now, now)
	if err != nil {
		t.Fatalf("RestoreAdmin() error = %v", err)
	}
	roles := authRoles(t, now)
	menus := authMenus(t, now)
	apis := authAPIs(t, now)
	store := &authStore{
		admin:       admin,
		roles:       roles,
		menus:       menus,
		apis:        apis,
		blacklisted: map[string]time.Time{},
	}
	recorder := &loginRecorder{}
	uc := authusecase.New(store, store, store, store, store, recorder, "test-secret", authusecase.WithClock(func() time.Time {
		return now
	}))
	return uc, recorder
}

func authRoles(t *testing.T, now time.Time) map[int64]accessdomain.Role {
	t.Helper()
	superRole, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", []string{accessdomain.PermissionAdminRead}, []int64{1}, []int64{1}, nil, []int64{1, 2}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole() error = %v", err)
	}
	operatorRole, err := accessdomain.RestoreRole(2, 1, "operator", "运营", []string{accessdomain.PermissionRoleRead}, []int64{2}, []int64{2}, []int64{22}, []int64{2}, "/roles", true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole(operator) error = %v", err)
	}
	return map[int64]accessdomain.Role{
		1: superRole,
		2: operatorRole,
	}
}

func authMenus(t *testing.T, now time.Time) []accessdomain.Menu {
	t.Helper()
	adminCreateButton, err := accessdomain.RestoreMenuButton(11, 1, "create", "新增管理员", now, now)
	if err != nil {
		t.Fatalf("RestoreMenuButton(admin create) error = %v", err)
	}
	adminDeleteButton, err := accessdomain.RestoreMenuButton(12, 1, "delete", "删除管理员", now, now)
	if err != nil {
		t.Fatalf("RestoreMenuButton(admin delete) error = %v", err)
	}
	adminMenu, err := accessdomain.RestoreMenu(1, 0, "管理员管理", "/admins", "user", false, "./Admins", accessdomain.MenuMeta{}, accessdomain.PermissionAdminRead, 10, true, []accessdomain.MenuButton{adminCreateButton, adminDeleteButton}, now, now)
	if err != nil {
		t.Fatalf("RestoreMenu() error = %v", err)
	}
	roleCreateButton, err := accessdomain.RestoreMenuButton(21, 2, "create", "新增角色", now, now)
	if err != nil {
		t.Fatalf("RestoreMenuButton(role create) error = %v", err)
	}
	roleUpdateButton, err := accessdomain.RestoreMenuButton(22, 2, "update", "编辑角色", now, now)
	if err != nil {
		t.Fatalf("RestoreMenuButton(role update) error = %v", err)
	}
	roleMenu, err := accessdomain.RestoreMenu(2, 0, "角色权限", "/roles", "safety", false, "./Roles", accessdomain.MenuMeta{KeepAlive: true}, accessdomain.PermissionRoleRead, 20, true, []accessdomain.MenuButton{roleCreateButton, roleUpdateButton}, now, now)
	if err != nil {
		t.Fatalf("RestoreMenu(role) error = %v", err)
	}
	return []accessdomain.Menu{adminMenu, roleMenu}
}

func authAPIs(t *testing.T, now time.Time) []accessdomain.API {
	t.Helper()
	adminAPI, err := accessdomain.RestoreAPI(1, "GET", "/api/admins", "管理员列表", "admin", accessdomain.PermissionAdminRead, false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(admin) error = %v", err)
	}
	roleAPI, err := accessdomain.RestoreAPI(2, "GET", "/api/roles", "角色列表", "role", accessdomain.PermissionRoleRead, false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(role) error = %v", err)
	}
	deleteAdminAPI, err := accessdomain.RestoreAPI(3, "DELETE", "/api/admins/:id", "删除管理员", "admin", accessdomain.PermissionAdminRead, false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(delete admin) error = %v", err)
	}
	roleDetailAPI, err := accessdomain.RestoreAPI(4, "GET", "/api/roles/:id", "角色详情", "role", accessdomain.PermissionRoleRead, false, now, now)
	if err != nil {
		t.Fatalf("RestoreAPI(role detail) error = %v", err)
	}
	return []accessdomain.API{adminAPI, roleAPI, deleteAdminAPI, roleDetailAPI}
}

type authStore struct {
	admin       identitydomain.Admin
	roles       map[int64]accessdomain.Role
	menus       []accessdomain.Menu
	apis        []accessdomain.API
	blacklisted map[string]time.Time
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

func (s *authStore) ListAPIs(context.Context) ([]accessdomain.API, error) {
	return s.apis, nil
}

func (s *authStore) AddJWTBlacklist(_ context.Context, entry authusecase.JWTBlacklistEntry) error {
	s.blacklisted[entry.TokenHash] = entry.ExpiresAt
	return nil
}

func (s *authStore) JWTBlacklisted(_ context.Context, tokenHash string, now time.Time) (bool, error) {
	expiresAt, ok := s.blacklisted[tokenHash]
	return ok && now.Before(expiresAt), nil
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
