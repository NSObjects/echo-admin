package authhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
	"golang.org/x/crypto/bcrypt"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	authdomain "github.com/NSObjects/echo-admin/internal/modules/auth/domain"
	authhttp "github.com/NSObjects/echo-admin/internal/modules/auth/http"
	authusecase "github.com/NSObjects/echo-admin/internal/modules/auth/usecase"
	identitydomain "github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/server/middlewares"
)

func TestAuthFlowReturnsCurrentUser(t *testing.T) {
	e := newTestEcho(t)
	client := newSessionClient(e)

	login := client.doJSON(t, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", login.Code, http.StatusOK, login.Body.String())
	}
	loginBody := decodeLoginResponse(t, login)
	if loginBody.Data.User.Username != "admin" {
		t.Fatalf("login username = %q, want admin", loginBody.Data.User.Username)
	}
	assertLoginCookies(t, client)

	me := client.doJSON(t, http.MethodGet, "/api/auth/me", "")
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

func TestLogoutRevokesCurrentLoginSession(t *testing.T) {
	e := newTestEcho(t)
	client := newSessionClient(e)
	login := client.doJSON(t, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", login.Code, http.StatusOK, login.Body.String())
	}

	logout := client.doJSON(t, http.MethodPost, "/api/auth/logout", "")
	if logout.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want %d: %s", logout.Code, http.StatusOK, logout.Body.String())
	}
	me := client.doJSON(t, http.MethodGet, "/api/auth/me", "")
	if me.Code != http.StatusUnauthorized {
		t.Fatalf("me status after logout = %d, want %d: %s", me.Code, http.StatusUnauthorized, me.Body.String())
	}
}

func TestLogoutOthersRevokesOtherLoginSessions(t *testing.T) {
	e := newTestEcho(t)
	current := newSessionClient(e)
	other := newSessionClient(e)
	if login := current.doJSON(t, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`); login.Code != http.StatusOK {
		t.Fatalf("current login status = %d, want %d: %s", login.Code, http.StatusOK, login.Body.String())
	}
	if login := other.doJSON(t, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`); login.Code != http.StatusOK {
		t.Fatalf("other login status = %d, want %d: %s", login.Code, http.StatusOK, login.Body.String())
	}

	logout := current.doJSON(t, http.MethodPost, "/api/auth/logout-others", "")
	if logout.Code != http.StatusOK {
		t.Fatalf("logout others status = %d, want %d: %s", logout.Code, http.StatusOK, logout.Body.String())
	}
	if me := current.doJSON(t, http.MethodGet, "/api/auth/me", ""); me.Code != http.StatusOK {
		t.Fatalf("current me status = %d, want %d: %s", me.Code, http.StatusOK, me.Body.String())
	}
	if me := other.doJSON(t, http.MethodGet, "/api/auth/me", ""); me.Code != http.StatusUnauthorized {
		t.Fatalf("other me status = %d, want %d: %s", me.Code, http.StatusUnauthorized, me.Body.String())
	}
}

func TestChangePasswordKeepsCurrentSession(t *testing.T) {
	e := newTestEcho(t)
	client := newSessionClient(e)
	login := client.doJSON(t, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", login.Code, http.StatusOK, login.Body.String())
	}

	change := client.doJSON(t, http.MethodPost, "/api/auth/password", `{"current_password":"123456","new_password":"changed123"}`)
	if change.Code != http.StatusOK {
		t.Fatalf("change password status = %d, want %d: %s", change.Code, http.StatusOK, change.Body.String())
	}
	me := client.doJSON(t, http.MethodGet, "/api/auth/me", "")
	if me.Code != http.StatusOK {
		t.Fatalf("me status after password change = %d, want %d: %s", me.Code, http.StatusOK, me.Body.String())
	}
	oldLogin := client.doJSON(t, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`)
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("old login status = %d, want %d: %s", oldLogin.Code, http.StatusUnauthorized, oldLogin.Body.String())
	}
	newLogin := client.doJSON(t, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"changed123"}`)
	if newLogin.Code != http.StatusOK {
		t.Fatalf("new login status = %d, want %d: %s", newLogin.Code, http.StatusOK, newLogin.Body.String())
	}
}

func TestUpdateProfileReturnsCurrentUser(t *testing.T) {
	e := newTestEcho(t)
	client := newSessionClient(e)
	login := client.doJSON(t, http.MethodPost, "/api/auth/login", `{"username":"admin","password":"123456"}`)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", login.Code, http.StatusOK, login.Body.String())
	}

	update := client.doJSON(t, http.MethodPatch, "/api/auth/me", `{"display_name":"平台管理员","email":"ops@example.com"}`)
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
	uc := authusecase.New(store, store, store, store, store, store, &loginRecorder{})
	handler := authhttp.New(uc, false)

	e := echo.New()
	e.Validator = &middlewares.Validator{Validator: validator.New()}
	e.HTTPErrorHandler = middlewares.ErrorHandler
	e.Use(middlewares.RequestContext())
	sessionMiddleware, err := middlewares.LoginSession(&middlewares.LoginSessionConfig{
		Enabled:       true,
		CookieName:    middlewares.LoginSessionCookieName,
		SkipPaths:     []string{"/api/auth/login"},
		Authenticator: sessionAuthenticator{auth: uc},
	})
	if err != nil {
		t.Fatalf("LoginSession() error = %v", err)
	}
	e.Use(sessionMiddleware)
	e.Use(echomiddleware.CSRFWithConfig(middlewares.CSRFConfig([]string{"/api/auth/login"}, false)))
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
		admin:         admin,
		role:          role,
		menu:          menu,
		nextSessionID: 1,
		sessions:      map[int64]authdomain.LoginSession{},
		sessionByHash: map[string]int64{},
	}
}

type authStore struct {
	admin         identitydomain.Admin
	role          accessdomain.Role
	menu          accessdomain.Menu
	nextSessionID int64
	sessions      map[int64]authdomain.LoginSession
	sessionByHash map[string]int64
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

func (s *authStore) FindAPIByRoute(context.Context, string, string) (accessdomain.API, error) {
	return accessdomain.API{}, apperr.NewNotFound("api")
}

func (s *authStore) CreateLoginSession(_ context.Context, session authdomain.LoginSession) (authdomain.LoginSession, error) {
	session.ID = s.nextSessionID
	s.nextSessionID++
	s.sessions[session.ID] = session
	s.sessionByHash[session.TokenHash] = session.ID
	return session, nil
}

func (s *authStore) FindLoginSessionByTokenHash(_ context.Context, tokenHash string) (authdomain.LoginSession, bool, error) {
	id, ok := s.sessionByHash[tokenHash]
	if !ok {
		return authdomain.LoginSession{}, false, nil
	}
	return s.sessions[id], true, nil
}

func (s *authStore) RefreshLoginSession(_ context.Context, session authdomain.LoginSession) error {
	s.sessions[session.ID] = session
	return nil
}

func (s *authStore) UpdateLoginSessionRole(_ context.Context, sessionID, roleID int64, now time.Time) error {
	session := s.sessions[sessionID]
	session.ActiveRoleID = roleID
	session.UpdatedAt = now
	s.sessions[sessionID] = session
	return nil
}

func (s *authStore) RevokeLoginSession(_ context.Context, sessionID int64, reason string, now time.Time) error {
	session := s.sessions[sessionID]
	session.RevokedAt = &now
	session.RevokedReason = reason
	session.UpdatedAt = now
	s.sessions[sessionID] = session
	return nil
}

func (s *authStore) RevokeOtherLoginSessions(_ context.Context, adminID, keepSessionID int64, reason string, now time.Time) error {
	for id, session := range s.sessions {
		if session.AdminID != adminID || id == keepSessionID || session.RevokedAt != nil {
			continue
		}
		session.RevokedAt = &now
		session.RevokedReason = reason
		session.UpdatedAt = now
		s.sessions[id] = session
	}
	return nil
}

func (s *authStore) RevokeLoginSessions(_ context.Context, adminID int64, reason string, now time.Time) error {
	for id, session := range s.sessions {
		if session.AdminID != adminID || session.RevokedAt != nil {
			continue
		}
		session.RevokedAt = &now
		session.RevokedReason = reason
		session.UpdatedAt = now
		s.sessions[id] = session
	}
	return nil
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

type sessionAuthenticator struct {
	auth *authusecase.Usecase
}

func (a sessionAuthenticator) AuthenticateLoginSession(ctx context.Context, token string) (middlewares.LoginSessionIdentity, error) {
	identity, err := a.auth.AuthenticateLoginSession(ctx, token)
	if err != nil {
		return middlewares.LoginSessionIdentity{}, err
	}
	return middlewares.LoginSessionIdentity{
		SessionID: strconv.FormatInt(identity.SessionID, 10),
		UserID:    strconv.FormatInt(identity.AdminID, 10),
		RoleID:    strconv.FormatInt(identity.RoleID, 10),
	}, nil
}

type sessionClient struct {
	echo    *echo.Echo
	cookies map[string]*http.Cookie
}

func newSessionClient(e *echo.Echo) *sessionClient {
	return &sessionClient{echo: e, cookies: map[string]*http.Cookie{}}
}

func (c *sessionClient) doJSON(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for _, cookie := range c.cookies {
		if cookie.MaxAge < 0 {
			continue
		}
		req.AddCookie(cookie)
	}
	if csrf, ok := c.cookies[middlewares.CSRFCookieName]; ok && unsafeMethod(method) {
		req.Header.Set(echo.HeaderXCSRFToken, csrf.Value)
	}
	rec := httptest.NewRecorder()
	c.echo.ServeHTTP(rec, req)
	response := rec.Result()
	defer response.Body.Close()
	for _, cookie := range response.Cookies() {
		if cookie.MaxAge < 0 {
			delete(c.cookies, cookie.Name)
			continue
		}
		c.cookies[cookie.Name] = cookie
	}
	return rec
}

func unsafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return false
	default:
		return true
	}
}

func assertLoginCookies(t *testing.T, client *sessionClient) {
	t.Helper()
	sessionCookie, ok := client.cookies[middlewares.LoginSessionCookieName]
	if !ok {
		t.Fatalf("%s cookie missing", middlewares.LoginSessionCookieName)
	}
	if !sessionCookie.HttpOnly {
		t.Fatalf("%s HttpOnly = false, want true", middlewares.LoginSessionCookieName)
	}
	csrfCookie, ok := client.cookies[middlewares.CSRFCookieName]
	if !ok {
		t.Fatalf("%s cookie missing", middlewares.CSRFCookieName)
	}
	if csrfCookie.HttpOnly {
		t.Fatalf("%s HttpOnly = true, want false", middlewares.CSRFCookieName)
	}
}

type loginResponse struct {
	Data struct {
		User struct {
			Username string `json:"username"`
		} `json:"user"`
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
