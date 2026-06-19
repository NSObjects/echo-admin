package authhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	authhttp "github.com/NSObjects/echo-admin/internal/modules/auth/http"
	authusecase "github.com/NSObjects/echo-admin/internal/modules/auth/usecase"
	identitydomain "github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	"github.com/NSObjects/echo-admin/internal/platform/server/middlewares"
)

func TestAuthFlowReturnsCurrentUser(t *testing.T) {
	e := newTestEcho(t)
	login := doJSON(t, e, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"admin123"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", login.Code, http.StatusOK, login.Body.String())
	}
	loginBody := decodeLoginResponse(t, login)
	if loginBody.Data.Token == "" {
		t.Fatal("login token is empty, want bearer token")
	}

	me := doJSON(t, e, http.MethodGet, "/api/auth/me", "", loginBody.Data.Token)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d: %s", me.Code, http.StatusOK, me.Body.String())
	}
	body := decodeCurrentUserResponse(t, me)
	if body.Data.Username != "admin" {
		t.Fatalf("me username = %q, want admin", body.Data.Username)
	}
}

func newTestEcho(t *testing.T) *echo.Echo {
	t.Helper()
	store := newAuthStore(t)
	uc := authusecase.New(store, store, store, &loginRecorder{}, "test-secret")
	handler := authhttp.New(uc)

	e := echo.New()
	e.Validator = &middlewares.Validator{Validator: validator.New()}
	e.HTTPErrorHandler = middlewares.ErrorHandler
	e.Use(middlewares.RequestContext())
	jwtMiddleware, err := middlewares.JWT(&middlewares.JWTConfig{
		Enabled:    true,
		SigningKey: []byte("test-secret"),
		SkipPaths:  []string{"/api/auth/login"},
	})
	if err != nil {
		t.Fatalf("JWT() error = %v", err)
	}
	e.Use(jwtMiddleware)
	authhttp.Register(e.Group("/api"), handler)
	return e
}

func newAuthStore(t *testing.T) *authStore {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	now := time.Unix(1_800_000_000, 0).UTC()
	admin, err := identitydomain.RestoreAdmin(1, "admin", "系统管理员", "admin@example.com", hash, []int64{1}, 1, true, now, now)
	if err != nil {
		t.Fatalf("RestoreAdmin() error = %v", err)
	}
	role, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", []string{accessdomain.PermissionAdminRead}, []int64{1}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole() error = %v", err)
	}
	menu, err := accessdomain.RestoreMenu(1, 0, "管理员管理", "/admins", "user", accessdomain.PermissionAdminRead, 10, true, now, now)
	if err != nil {
		t.Fatalf("RestoreMenu() error = %v", err)
	}
	return &authStore{admin: admin, role: role, menu: menu}
}

func doJSON(t *testing.T, e *echo.Echo, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if token != "" {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
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

func (s *authStore) Update(_ context.Context, admin identitydomain.Admin) (identitydomain.Admin, error) {
	s.admin = admin
	return admin, nil
}

func (s *authStore) FindRoleByID(context.Context, int64) (accessdomain.Role, error) {
	return s.role, nil
}

func (s *authStore) ListMenus(context.Context) ([]accessdomain.Menu, error) {
	return []accessdomain.Menu{s.menu}, nil
}

type loginRecorder struct{}

func (r *loginRecorder) RecordLogin(context.Context, authusecase.LoginRecord) error {
	return nil
}

type loginResponse struct {
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

func decodeLoginResponse(t *testing.T, rec *httptest.ResponseRecorder) loginResponse {
	t.Helper()
	var body loginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode login response: %v\n%s", err, rec.Body.String())
	}
	return body
}

type currentUserResponse struct {
	Data struct {
		Username string `json:"username"`
	} `json:"data"`
}

func decodeCurrentUserResponse(t *testing.T, rec *httptest.ResponseRecorder) currentUserResponse {
	t.Helper()
	var body currentUserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode current user response: %v\n%s", err, rec.Body.String())
	}
	return body
}
