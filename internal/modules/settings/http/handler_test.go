package settingshttp_test

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
	settingsdomain "github.com/NSObjects/echo-admin/internal/modules/settings/domain"
	settingshttp "github.com/NSObjects/echo-admin/internal/modules/settings/http"
	settingsusecase "github.com/NSObjects/echo-admin/internal/modules/settings/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
	"github.com/NSObjects/echo-admin/internal/platform/server/middlewares"
)

func TestUpsertConfigRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newSettingsEcho(nil)

	rec := doJSON(t, e, http.MethodPut, "/api/system/configs/site_name", `{"name":"站点名称","value":"Echo Admin","public":true}`, "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("upsert config status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionConfigUpdate {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionConfigUpdate)
	}
	if store.upsertCalls != 1 {
		t.Fatalf("upsertCalls = %d, want 1", store.upsertCalls)
	}
	if got := store.config.Key; got != "site_name" {
		t.Fatalf("config key = %q, want site_name", got)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].ActorID; got != 42 {
		t.Fatalf("operation actor id = %d, want 42", got)
	}
}

func TestListDictionariesRejectsUnauthorizedBeforeStore(t *testing.T) {
	e, store, recorder, _ := newSettingsEcho(apperr.NewPermissionDenied("dict", "read"))

	rec := doJSON(t, e, http.MethodGet, "/api/dictionaries", "", "42")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("list dictionaries status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if store.listDictionaryCalls != 0 {
		t.Fatalf("listDictionaryCalls = %d, want 0", store.listDictionaryCalls)
	}
	if len(recorder.records) != 0 {
		t.Fatalf("operation records = %d, want 0", len(recorder.records))
	}
}

func TestDeleteDictionaryRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newSettingsEcho(nil)

	rec := doJSON(t, e, http.MethodDelete, "/api/dictionaries/color", "", "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete dictionary status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionDictDelete {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionDictDelete)
	}
	if got := store.deletedDictionaryCode; got != "color" {
		t.Fatalf("deletedDictionaryCode = %q, want color", got)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].Resource; got != "dictionary" {
		t.Fatalf("operation resource = %q, want dictionary", got)
	}
}

func newSettingsEcho(authErr error) (*echo.Echo, *settingsStore, *operationRecorder, *settingsAuthorizer) {
	store := &settingsStore{}
	uc := settingsusecase.New(store)
	auth := &settingsAuthorizer{err: authErr}
	recorder := &operationRecorder{}
	handler := settingshttp.New(uc, auth, recorder)

	e := echo.New()
	e.Validator = &middlewares.Validator{Validator: validator.New()}
	e.HTTPErrorHandler = middlewares.ErrorHandler
	settingshttp.Register(e.Group("/api"), handler)
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

type settingsStore struct {
	upsertCalls           int
	listDictionaryCalls   int
	config                settingsdomain.SystemConfig
	deletedDictionaryCode string
}

func (s *settingsStore) ListConfigs(ctx context.Context) ([]settingsdomain.SystemConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *settingsStore) UpsertConfig(ctx context.Context, config settingsdomain.SystemConfig) (settingsdomain.SystemConfig, error) {
	if err := ctx.Err(); err != nil {
		return settingsdomain.SystemConfig{}, err
	}
	s.upsertCalls++
	s.config = config
	return settingsdomain.RestoreSystemConfig(config.Key, config.Name, config.Value, config.Public, fixedTime())
}

func (s *settingsStore) ListDictionaries(ctx context.Context) ([]settingsdomain.Dictionary, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.listDictionaryCalls++
	return nil, nil
}

func (s *settingsStore) CreateDictionary(ctx context.Context, dictionary settingsdomain.Dictionary) (settingsdomain.Dictionary, error) {
	if err := ctx.Err(); err != nil {
		return settingsdomain.Dictionary{}, err
	}
	return dictionary, nil
}

func (s *settingsStore) UpdateDictionary(ctx context.Context, dictionary settingsdomain.Dictionary) (settingsdomain.Dictionary, error) {
	if err := ctx.Err(); err != nil {
		return settingsdomain.Dictionary{}, err
	}
	return dictionary, nil
}

func (s *settingsStore) DeleteDictionary(ctx context.Context, code string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.deletedDictionaryCode = code
	return nil
}

func (s *settingsStore) AddDictionaryItem(ctx context.Context, _ string, item settingsdomain.DictionaryItem) (settingsdomain.Dictionary, error) {
	if err := ctx.Err(); err != nil {
		return settingsdomain.Dictionary{}, err
	}
	return settingsdomain.RestoreDictionary(1, "status", "状态", []settingsdomain.DictionaryItem{item}, fixedTime(), fixedTime())
}

func (s *settingsStore) UpdateDictionaryItem(ctx context.Context, _ string, item settingsdomain.DictionaryItem) (settingsdomain.Dictionary, error) {
	if err := ctx.Err(); err != nil {
		return settingsdomain.Dictionary{}, err
	}
	return settingsdomain.RestoreDictionary(1, "status", "状态", []settingsdomain.DictionaryItem{item}, fixedTime(), fixedTime())
}

func (s *settingsStore) DeleteDictionaryItem(ctx context.Context, _ string, itemID int64) (settingsdomain.Dictionary, error) {
	if err := ctx.Err(); err != nil {
		return settingsdomain.Dictionary{}, err
	}
	item, err := settingsdomain.RestoreDictionaryItem(itemID, "启用", "enabled", 10, true)
	if err != nil {
		return settingsdomain.Dictionary{}, err
	}
	return settingsdomain.RestoreDictionary(1, "status", "状态", []settingsdomain.DictionaryItem{item}, fixedTime(), fixedTime())
}

type settingsAuthorizer struct {
	err         error
	permissions []string
}

func (a *settingsAuthorizer) RequirePermission(_ context.Context, permission string) error {
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

func fixedTime() time.Time {
	return time.Unix(1_800_000_000, 0).UTC()
}
