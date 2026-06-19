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
	if !contains(current.Permissions, accessdomain.PermissionAdminManage) {
		t.Fatalf("permissions = %v, want %s", current.Permissions, accessdomain.PermissionAdminManage)
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

	if err := uc.RequirePermission(ctx, accessdomain.PermissionAdminManage); err != nil {
		t.Fatalf("RequirePermission(allowed) error = %v", err)
	}
	if err := uc.RequirePermission(ctx, accessdomain.PermissionRoleManage); err == nil {
		t.Fatal("RequirePermission(denied) error = nil, want permission denied")
	}
}

func newUsecase(t *testing.T) (*authusecase.Usecase, *loginRecorder) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	now := time.Unix(1_800_000_000, 0).UTC()
	admin, err := identitydomain.RestoreAdmin(1, "admin", "系统管理员", "admin@example.com", hash, []int64{1}, true, now, now)
	if err != nil {
		t.Fatalf("RestoreAdmin() error = %v", err)
	}
	role, err := accessdomain.RestoreRole(1, "super_admin", "超级管理员", []string{accessdomain.PermissionAdminManage}, []int64{1}, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole() error = %v", err)
	}
	menu, err := accessdomain.RestoreMenu(1, 0, "管理员管理", "/admins", "user", accessdomain.PermissionAdminManage, 10, true, now, now)
	if err != nil {
		t.Fatalf("RestoreMenu() error = %v", err)
	}
	store := &authStore{admin: admin, role: role, menu: menu}
	recorder := &loginRecorder{}
	uc := authusecase.New(store, store, store, recorder, "test-secret", authusecase.WithClock(func() time.Time {
		return now
	}))
	return uc, recorder
}

type authStore struct {
	admin identitydomain.Admin
	role  accessdomain.Role
	menu  accessdomain.Menu
}

func (s *authStore) FindByUsername(context.Context, string) (identitydomain.Admin, error) {
	return s.admin, nil
}

func (s *authStore) FindByID(context.Context, int64) (identitydomain.Admin, error) {
	return s.admin, nil
}

func (s *authStore) FindRoleByID(context.Context, int64) (accessdomain.Role, error) {
	return s.role, nil
}

func (s *authStore) ListMenus(context.Context) ([]accessdomain.Menu, error) {
	return []accessdomain.Menu{s.menu}, nil
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
