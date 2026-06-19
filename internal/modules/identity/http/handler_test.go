package identityhttp_test

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
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	identitydomain "github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	identityhttp "github.com/NSObjects/echo-admin/internal/modules/identity/http"
	identityusecase "github.com/NSObjects/echo-admin/internal/modules/identity/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
	"github.com/NSObjects/echo-admin/internal/platform/server/middlewares"
)

func TestCreateAdminRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newIdentityEcho(nil)

	rec := doJSON(t, e, http.MethodPost, "/api/admins", `{"username":"operator","display_name":"运营","email":"operator@example.com","password":"operator123","role_ids":[1],"active":true}`, "42")
	if rec.Code != http.StatusCreated {
		t.Fatalf("create admin status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionAdminCreate {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionAdminCreate)
	}
	if store.createCalls != 1 {
		t.Fatalf("createCalls = %d, want 1", store.createCalls)
	}
	if got := store.created.Username; got != "operator" {
		t.Fatalf("created username = %q, want operator", got)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].ActorID; got != 42 {
		t.Fatalf("operation actor id = %d, want 42", got)
	}
	if got := recorder.records[0].Resource; got != "admin" {
		t.Fatalf("operation resource = %q, want admin", got)
	}
}

func TestListAdminsRejectsUnauthorizedBeforeStore(t *testing.T) {
	e, store, recorder, _ := newIdentityEcho(apperr.NewPermissionDenied("admin", "read"))

	rec := doJSON(t, e, http.MethodGet, "/api/admins", "", "42")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("list admins status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if store.listCalls != 0 {
		t.Fatalf("listCalls = %d, want 0", store.listCalls)
	}
	if len(recorder.records) != 0 {
		t.Fatalf("operation records = %d, want 0", len(recorder.records))
	}
}

func TestDeleteAdminRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newIdentityEcho(nil)
	store.admins = map[int64]identitydomain.Admin{
		7: newAdmin(t, 7, "operator"),
	}

	rec := doJSON(t, e, http.MethodDelete, "/api/admins/7", "", "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete admin status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionAdminDelete {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionAdminDelete)
	}
	if store.deletedID != 7 {
		t.Fatalf("deletedID = %d, want 7", store.deletedID)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].Action; got != "delete" {
		t.Fatalf("operation action = %q, want delete", got)
	}
}

func newIdentityEcho(authErr error) (*echo.Echo, *identityStore, *operationRecorder, *identityAuthorizer) {
	store := &identityStore{}
	uc := identityusecase.New(store, identityRoleScope{})
	auth := &identityAuthorizer{err: authErr}
	recorder := &operationRecorder{}
	handler := identityhttp.New(uc, auth, recorder)

	e := echo.New()
	e.Validator = &middlewares.Validator{Validator: validator.New()}
	e.HTTPErrorHandler = middlewares.ErrorHandler
	identityhttp.Register(e.Group("/api"), handler)
	return e, store, recorder, auth
}

func doJSON(t *testing.T, e *echo.Echo, method, path, body, userID string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	req = req.WithContext(requestctx.WithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

type identityStore struct {
	createCalls int
	listCalls   int
	created     identitydomain.Admin
	admins      map[int64]identitydomain.Admin
	deletedID   int64
}

func (s *identityStore) FindByUsername(context.Context, string) (identitydomain.Admin, error) {
	return identitydomain.Admin{}, apperr.NewNotFound("admin")
}

func (s *identityStore) FindByID(_ context.Context, id int64) (identitydomain.Admin, error) {
	admin, ok := s.admins[id]
	if ok {
		return admin, nil
	}
	return identitydomain.Admin{}, apperr.NewNotFound("admin")
}

func (s *identityStore) ListAll(ctx context.Context) ([]identitydomain.Admin, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.listCalls++
	return nil, nil
}

func (s *identityStore) Create(ctx context.Context, admin identitydomain.Admin) (identitydomain.Admin, error) {
	if err := ctx.Err(); err != nil {
		return identitydomain.Admin{}, err
	}
	s.createCalls++
	s.created = admin
	return identitydomain.RestoreAdmin(1, admin.Username, admin.DisplayName, admin.Email, admin.PasswordHash, admin.RoleIDs, admin.ActiveRoleID, admin.Active, fixedTime(), fixedTime())
}

func (s *identityStore) Update(ctx context.Context, admin identitydomain.Admin) (identitydomain.Admin, error) {
	if err := ctx.Err(); err != nil {
		return identitydomain.Admin{}, err
	}
	return admin, nil
}

func (s *identityStore) Delete(_ context.Context, id int64) error {
	s.deletedID = id
	return nil
}

type identityAuthorizer struct {
	err         error
	permissions []string
}

func (a *identityAuthorizer) RequirePermission(_ context.Context, permission string) error {
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

type identityRoleScope struct{}

func (identityRoleScope) AssignableRoleIDs(context.Context) ([]int64, error) {
	return []int64{1}, nil
}

func (identityRoleScope) EnsureAssignableRoles(_ context.Context, _ []int64) error {
	return nil
}

func newAdmin(t *testing.T, id int64, username string) identitydomain.Admin {
	t.Helper()
	admin, err := identitydomain.RestoreAdmin(id, username, username, username+"@example.com", []byte("hash"), []int64{1}, 1, true, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreAdmin() error = %v", err)
	}
	return admin
}

func fixedTime() time.Time {
	return time.Unix(1_800_000_000, 0).UTC()
}
