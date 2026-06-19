package fileassethttp_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	filedomain "github.com/NSObjects/echo-admin/internal/modules/fileasset/domain"
	filehttp "github.com/NSObjects/echo-admin/internal/modules/fileasset/http"
	fileusecase "github.com/NSObjects/echo-admin/internal/modules/fileasset/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
	"github.com/NSObjects/echo-admin/internal/platform/server/middlewares"
)

func TestUploadFileStoresBytesMetadataAndOperation(t *testing.T) {
	e, store, recorder, auth, uploadDir := newFileEcho(t, nil)

	rec := doMultipart(t, e, "/api/files", "hello.txt", "hello", "7")
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionFileUpload {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionFileUpload)
	}
	if store.createCalls != 1 {
		t.Fatalf("createCalls = %d, want 1", store.createCalls)
	}
	if store.created.Name != "hello.txt" {
		t.Fatalf("created name = %q, want hello.txt", store.created.Name)
	}
	storedName := strings.TrimPrefix(store.created.URL, "/api/uploads/")
	if storedName == store.created.URL || storedName == "" {
		t.Fatalf("created URL = %q, want /api/uploads/<file>", store.created.URL)
	}
	if _, err := os.Stat(filepath.Join(uploadDir, storedName)); err != nil {
		t.Fatalf("stat uploaded file: %v", err)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].ActorID; got != 7 {
		t.Fatalf("operation actor id = %d, want 7", got)
	}
}

func TestUploadFileRejectsUnauthorizedBeforeStore(t *testing.T) {
	e, store, recorder, _, _ := newFileEcho(t, apperr.NewPermissionDenied("file", "upload"))

	rec := doMultipart(t, e, "/api/files", "hello.txt", "hello", "7")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("upload status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if store.createCalls != 0 {
		t.Fatalf("createCalls = %d, want 0", store.createCalls)
	}
	if len(recorder.records) != 0 {
		t.Fatalf("operation records = %d, want 0", len(recorder.records))
	}
}

func newFileEcho(t *testing.T, authErr error) (*echo.Echo, *fileStore, *operationRecorder, *fileAuthorizer, string) {
	t.Helper()
	store := &fileStore{}
	uc := fileusecase.New(store)
	auth := &fileAuthorizer{err: authErr}
	recorder := &operationRecorder{}
	uploadDir := t.TempDir()
	handler := filehttp.New(uc, auth, recorder, uploadDir)

	e := echo.New()
	e.Validator = &middlewares.Validator{Validator: validator.New()}
	e.HTTPErrorHandler = middlewares.ErrorHandler
	filehttp.Register(e.Group("/api"), handler)
	return e, store, recorder, auth, uploadDir
}

func doMultipart(t *testing.T, e *echo.Echo, path, filename, body, userID string) *httptest.ResponseRecorder {
	t.Helper()
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write([]byte(body)); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, &payload)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	req = req.WithContext(requestctx.WithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

type fileStore struct {
	createCalls int
	created     filedomain.FileObject
}

func (s *fileStore) CreateFile(ctx context.Context, file filedomain.FileObject) (filedomain.FileObject, error) {
	if err := ctx.Err(); err != nil {
		return filedomain.FileObject{}, err
	}
	s.createCalls++
	s.created = file
	return filedomain.RestoreFileObject(1, file.Name, file.URL, file.Size, file.ContentType, time.Unix(1_800_000_000, 0).UTC())
}

func (s *fileStore) ListFiles(ctx context.Context, _ fileusecase.ListFilter) ([]filedomain.FileObject, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	return nil, 0, nil
}

type fileAuthorizer struct {
	err         error
	permissions []string
}

func (a *fileAuthorizer) RequirePermission(_ context.Context, permission string) error {
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
