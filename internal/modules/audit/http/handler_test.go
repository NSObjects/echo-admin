package audithttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	auditdomain "github.com/NSObjects/echo-admin/internal/modules/audit/domain"
	audithttp "github.com/NSObjects/echo-admin/internal/modules/audit/http"
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/server/middlewares"
)

func TestListOperationLogsRequiresPermission(t *testing.T) {
	e, store, auth := newAuditEcho(t, nil)

	rec := doGET(e, "/api/logs/operations")
	if rec.Code != http.StatusOK {
		t.Fatalf("operation logs status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionLogRead {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionLogRead)
	}
	if store.listOperationCalls != 1 {
		t.Fatalf("listOperationCalls = %d, want 1", store.listOperationCalls)
	}
}

func TestListLoginLogsRejectsUnauthorizedBeforeStore(t *testing.T) {
	e, store, _ := newAuditEcho(t, apperr.NewPermissionDenied("log", "read"))

	rec := doGET(e, "/api/logs/logins")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("login logs status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if store.listLoginCalls != 0 {
		t.Fatalf("listLoginCalls = %d, want 0", store.listLoginCalls)
	}
}

func newAuditEcho(t *testing.T, authErr error) (*echo.Echo, *auditStore, *auditAuthorizer) {
	t.Helper()
	store := newAuditStore(t)
	uc := auditusecase.New(store)
	auth := &auditAuthorizer{err: authErr}
	handler := audithttp.New(uc, auth)

	e := echo.New()
	e.HTTPErrorHandler = middlewares.ErrorHandler
	audithttp.Register(e.Group("/api"), handler)
	return e, store, auth
}

func doGET(e *echo.Echo, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func newAuditStore(t *testing.T) *auditStore {
	t.Helper()
	log, err := auditdomain.RestoreOperationLog(1, 42, "create", "admin", "1", http.MethodPost, "/api/admins", "127.0.0.1", "test", true, "created admin", fixedTime())
	if err != nil {
		t.Fatalf("RestoreOperationLog() error = %v", err)
	}
	login, err := auditdomain.RestoreLoginLog(1, 42, "admin", "127.0.0.1", "test", true, "login succeeded", fixedTime())
	if err != nil {
		t.Fatalf("RestoreLoginLog() error = %v", err)
	}
	return &auditStore{operationLog: log, loginLog: login}
}

type auditStore struct {
	listOperationCalls int
	listLoginCalls     int
	operationLog       auditdomain.OperationLog
	loginLog           auditdomain.LoginLog
}

func (s *auditStore) RecordOperation(ctx context.Context, log auditdomain.OperationLog) (auditdomain.OperationLog, error) {
	if err := ctx.Err(); err != nil {
		return auditdomain.OperationLog{}, err
	}
	return log, nil
}

func (s *auditStore) ListOperationLogs(ctx context.Context, _ auditusecase.ListFilter) ([]auditdomain.OperationLog, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	s.listOperationCalls++
	return []auditdomain.OperationLog{s.operationLog}, 1, nil
}

func (s *auditStore) RecordLogin(ctx context.Context, log auditdomain.LoginLog) (auditdomain.LoginLog, error) {
	if err := ctx.Err(); err != nil {
		return auditdomain.LoginLog{}, err
	}
	return log, nil
}

func (s *auditStore) ListLoginLogs(ctx context.Context, _ auditusecase.ListFilter) ([]auditdomain.LoginLog, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	s.listLoginCalls++
	return []auditdomain.LoginLog{s.loginLog}, 1, nil
}

type auditAuthorizer struct {
	err         error
	permissions []string
}

func (a *auditAuthorizer) RequirePermission(_ context.Context, permission string) error {
	a.permissions = append(a.permissions, permission)
	return a.err
}

func fixedTime() time.Time {
	return time.Unix(1_800_000_000, 0).UTC()
}
