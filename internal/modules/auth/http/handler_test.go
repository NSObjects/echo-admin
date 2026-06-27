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
	login := doJSON(t, e, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`, "")
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
	if len(body.Data.Menus) != 1 || body.Data.Menus[0].Component != "./Admins" {
		t.Fatalf("me menus = %#v, want one admins menu with component", body.Data.Menus)
	}
}

func TestLogoutBlacklistsCurrentToken(t *testing.T) {
	e := newTestEcho(t)
	login := doJSON(t, e, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", login.Code, http.StatusOK, login.Body.String())
	}
	token := decodeLoginResponse(t, login).Data.Token

	logout := doJSON(t, e, http.MethodPost, "/api/auth/logout", "", token)
	if logout.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want %d: %s", logout.Code, http.StatusOK, logout.Body.String())
	}
	me := doJSON(t, e, http.MethodGet, "/api/auth/me", "", token)
	if me.Code != http.StatusUnauthorized {
		t.Fatalf("me status after logout = %d, want %d: %s", me.Code, http.StatusUnauthorized, me.Body.String())
	}
}

func TestChangePasswordRevokesCurrentToken(t *testing.T) {
	e := newTestEcho(t)
	login := doJSON(t, e, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", login.Code, http.StatusOK, login.Body.String())
	}
	token := decodeLoginResponse(t, login).Data.Token

	change := doJSON(t, e, http.MethodPost, "/api/auth/password", `{"current_password":"123456","new_password":"changed123"}`, token)
	if change.Code != http.StatusOK {
		t.Fatalf("change password status = %d, want %d: %s", change.Code, http.StatusOK, change.Body.String())
	}
	me := doJSON(t, e, http.MethodGet, "/api/auth/me", "", token)
	if me.Code != http.StatusUnauthorized {
		t.Fatalf("me status after password change = %d, want %d: %s", me.Code, http.StatusUnauthorized, me.Body.String())
	}
	oldLogin := doJSON(t, e, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`, "")
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("old login status = %d, want %d: %s", oldLogin.Code, http.StatusUnauthorized, oldLogin.Body.String())
	}
	newLogin := doJSON(t, e, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"changed123"}`, "")
	if newLogin.Code != http.StatusOK {
		t.Fatalf("new login status = %d, want %d: %s", newLogin.Code, http.StatusOK, newLogin.Body.String())
	}
}

func TestUpdateProfileReturnsCurrentUser(t *testing.T) {
	e := newTestEcho(t)
	login := doJSON(t, e, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", login.Code, http.StatusOK, login.Body.String())
	}
	token := decodeLoginResponse(t, login).Data.Token

	update := doJSON(t, e, http.MethodPatch, "/api/auth/me", `{"display_name":"平台管理员","email":"ops@example.com"}`, token)
	if update.Code != http.StatusOK {
		t.Fatalf("update profile status = %d, want %d: %s", update.Code, http.StatusOK, update.Body.String())
	}
	body := decodeCurrentUserResponse(t, update)
	if body.Data.DisplayName != "平台管理员" {
		t.Fatalf("display name = %q, want 平台管理员", body.Data.DisplayName)
	}
	if body.Data.Email != "ops@example.com" {
		t.Fatalf("email = %q, want ops@example.com", body.Data.Email)
	}
}

func newTestEcho(t *testing.T) *echo.Echo {
	t.Helper()
	store := newAuthStore(t)
	uc := authusecase.New(store, store, store, store, store, store, &loginRecorder{}, "test-secret")
	handler := authhttp.New(uc)

	e := echo.New()
	e.Validator = &middlewares.Validator{Validator: validator.New()}
	e.HTTPErrorHandler = middlewares.ErrorHandler
	e.Use(middlewares.RequestContext())
	jwtMiddleware, err := middlewares.JWT(&middlewares.JWTConfig{
		Enabled:    true,
		SigningKey: []byte("test-secret"),
		SkipPaths:  []string{"/api/auth/login"},
		Blocklist:  uc,
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
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	now := time.Unix(1_800_000_000, 0).UTC()
	admin, err := identitydomain.RestoreAdmin(1, "admin", "系统管理员", "admin@example.com", hash, []int64{1}, 1, true, now, now)
	if err != nil {
		t.Fatalf("RestoreAdmin() error = %v", err)
	}
	role, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", []string{accessdomain.PermissionAdminRead}, []int64{1}, []int64{1}, nil, []int64{1}, accessdomain.DefaultRolePath, true, now, now)
	if err != nil {
		t.Fatalf("RestoreRole() error = %v", err)
	}
	menu, err := accessdomain.RestoreMenu(1, 0, "管理员管理", "/admins", "user", false, "./Admins", accessdomain.MenuMeta{}, accessdomain.PermissionAdminRead, 10, true, nil, now, now)
	if err != nil {
		t.Fatalf("RestoreMenu() error = %v", err)
	}
	return &authStore{
		admin:       admin,
		role:        role,
		menu:        menu,
		blacklisted: map[string]time.Time{},
	}
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
	admin       identitydomain.Admin
	role        accessdomain.Role
	menu        accessdomain.Menu
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

func (s *authStore) FindRoleByID(context.Context, int64) (accessdomain.Role, error) {
	return s.role, nil
}

func (s *authStore) ListMenus(context.Context) ([]accessdomain.Menu, error) {
	return []accessdomain.Menu{s.menu}, nil
}

func (s *authStore) ListAPIs(context.Context) ([]accessdomain.API, error) {
	return nil, nil
}

func (s *authStore) AddJWTBlacklist(_ context.Context, entry authusecase.JWTBlacklistEntry) error {
	s.blacklisted[entry.TokenHash] = entry.ExpiresAt
	return nil
}

func (s *authStore) JWTBlacklisted(_ context.Context, tokenHash string, now time.Time) (bool, error) {
	expiresAt, ok := s.blacklisted[tokenHash]
	return ok && now.Before(expiresAt), nil
}

func (s *authStore) CheckLoginAttempt(context.Context, string, time.Time) error {
	return nil
}

func (s *authStore) RecordLoginFailure(context.Context, string, time.Time) error {
	return nil
}

func (s *authStore) ResetLoginAttempts(context.Context, string) error {
	return nil
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
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
		Menus       []struct {
			Component string `json:"component"`
		} `json:"menus"`
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
