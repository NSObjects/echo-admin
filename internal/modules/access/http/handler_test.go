package accesshttp_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	accesshttp "github.com/NSObjects/echo-admin/internal/modules/access/http"
	accessusecase "github.com/NSObjects/echo-admin/internal/modules/access/usecase"
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
	"github.com/NSObjects/echo-admin/internal/platform/server/middlewares"
)

func TestCreateRoleRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newAccessEcho(nil)

	rec := doJSON(t, e, http.MethodPost, "/api/roles", `{"code":"operator","name":"运营","permissions":["log:read"],"menu_ids":[1],"active":true}`, "42")
	if rec.Code != http.StatusCreated {
		t.Fatalf("create role status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionRoleCreate {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionRoleCreate)
	}
	if store.createRoleCalls != 1 {
		t.Fatalf("createRoleCalls = %d, want 1", store.createRoleCalls)
	}
	if got := store.createdRole.Code; got != "operator" {
		t.Fatalf("created role code = %q, want operator", got)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].ActorID; got != 42 {
		t.Fatalf("operation actor id = %d, want 42", got)
	}
	if got := recorder.records[0].Resource; got != "role" {
		t.Fatalf("operation resource = %q, want role", got)
	}
}

func TestListMenusRejectsUnauthorizedBeforeStore(t *testing.T) {
	e, store, recorder, _ := newAccessEcho(apperr.NewPermissionDenied("menu", "read"))

	rec := doJSON(t, e, http.MethodGet, "/api/menus", "", "42")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("list menus status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if store.listMenuCalls != 0 {
		t.Fatalf("listMenuCalls = %d, want 0", store.listMenuCalls)
	}
	if len(recorder.records) != 0 {
		t.Fatalf("operation records = %d, want 0", len(recorder.records))
	}
}

func newAccessEcho(authErr error) (*echo.Echo, *accessStore, *operationRecorder, *accessAuthorizer) {
	store := &accessStore{}
	uc := accessusecase.New(store, accessAdminRoleReader{})
	auth := &accessAuthorizer{err: authErr}
	recorder := &operationRecorder{}
	handler := accesshttp.New(uc, auth, recorder)

	e := echo.New()
	e.Validator = &middlewares.Validator{Validator: validator.New()}
	e.HTTPErrorHandler = middlewares.ErrorHandler
	accesshttp.Register(e.Group("/api"), handler)
	return e, store, recorder, auth
}

func doJSON(t *testing.T, e *echo.Echo, method, path, body, userID string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	req = req.WithContext(requestctx.WithUserID(req.Context(), userID))
	req = req.WithContext(requestctx.WithRoleID(req.Context(), "1"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

type accessStore struct {
	createRoleCalls int
	listMenuCalls   int
	createdRole     accessdomain.Role
}

func (s *accessStore) FindRoleByID(context.Context, int64) (accessdomain.Role, error) {
	return accessdomain.Role{}, apperr.NewNotFound("role")
}

func (s *accessStore) FindRoleByCode(context.Context, string) (accessdomain.Role, error) {
	return accessdomain.Role{}, apperr.NewNotFound("role")
}

func (s *accessStore) ListAllRoles(ctx context.Context) ([]accessdomain.Role, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	role, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", accessusecase.AllPermissions(), []int64{1}, accessdomain.DefaultRolePath, true, fixedTime(), fixedTime())
	if err != nil {
		return nil, err
	}
	return []accessdomain.Role{role}, nil
}

func (s *accessStore) CreateRole(ctx context.Context, role accessdomain.Role) (accessdomain.Role, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.Role{}, err
	}
	s.createRoleCalls++
	s.createdRole = role
	return accessdomain.RestoreRole(1, role.ParentID, role.Code, role.Name, role.Permissions, role.MenuIDs, role.DefaultPath, role.Active, fixedTime(), fixedTime())
}

func (s *accessStore) UpdateRole(ctx context.Context, role accessdomain.Role) (accessdomain.Role, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.Role{}, err
	}
	return role, nil
}

func (s *accessStore) ListMenus(ctx context.Context) ([]accessdomain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.listMenuCalls++
	return nil, nil
}

func (s *accessStore) FindMenuByID(ctx context.Context, id int64) (accessdomain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.Menu{}, err
	}
	return accessdomain.RestoreMenu(id, 0, "Existing", "/existing", "menu", "log:read", 10, true, fixedTime(), fixedTime())
}

func (s *accessStore) CreateMenu(ctx context.Context, menu accessdomain.Menu) (accessdomain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.Menu{}, err
	}
	return menu, nil
}

func (s *accessStore) UpdateMenu(ctx context.Context, menu accessdomain.Menu) (accessdomain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.Menu{}, err
	}
	return menu, nil
}

type accessAuthorizer struct {
	err         error
	permissions []string
}

func (a *accessAuthorizer) RequirePermission(_ context.Context, permission string) error {
	a.permissions = append(a.permissions, permission)
	return a.err
}

type operationRecorder struct {
	records []auditusecase.OperationInput
}

func (r *operationRecorder) RecordOperation(_ context.Context, input auditusecase.OperationInput) (auditusecase.OperationLog, error) {
	r.records = append(r.records, input)
	return auditusecase.OperationLog{}, nil
}

type accessAdminRoleReader struct{}

func (accessAdminRoleReader) AdminRoleState(context.Context, int64) (accessusecase.AdminRoleState, error) {
	return accessusecase.AdminRoleState{RoleIDs: []int64{1}, ActiveRoleID: 1}, nil
}

func fixedTime() time.Time {
	return time.Unix(1_800_000_000, 0).UTC()
}
