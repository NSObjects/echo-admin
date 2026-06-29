package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
)

const testRequestID = "req-123"

func TestDefaultMiddlewareConfig(t *testing.T) {
	config := DefaultMiddlewareConfig()

	assert.NotNil(t, config)
	assert.True(t, config.EnableRecovery)
	assert.True(t, config.EnableRequestContext)
	assert.True(t, config.EnableLogger)
	assert.True(t, config.EnableGzip)
	assert.False(t, config.EnableCORS)
	assert.False(t, config.EnableAPIKey)
	assert.NotNil(t, config.APIKey)
	assert.False(t, config.EnableInstallationGate)
	assert.NotNil(t, config.InstallationGate)
	assert.False(t, config.EnableLoginSession)
	assert.NotNil(t, config.LoginSession)
	assert.False(t, config.EnableCSRF)
}

func TestDefaultLoginSessionConfig(t *testing.T) {
	config := DefaultLoginSessionConfig()

	assert.NotNil(t, config)
	assert.False(t, config.Enabled)
	assert.NotNil(t, config.SkipPaths)
	assert.Equal(t, LoginSessionCookieName, config.CookieName)
}

func TestInstallationGateBlocksPrivateRoutesWhenUninitialized(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	gate, err := InstallationGate(&InstallationGateConfig{
		Reader:  &installationStateReader{initialized: false},
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("InstallationGate() error = %v", err)
	}
	e.Use(gate)
	e.GET("/api/auth/me", func(c *echo.Context) error {
		t.Fatal("private handler reached before installation completed")
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assertErrorPayload(t, rec, apperr.ErrSystemUninitialized, "system is not initialized")
}

func TestInstallationGateSkipsSetupAndSystemRoutes(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	reader := &installationStateReader{initialized: false}
	gate, err := InstallationGate(&InstallationGateConfig{
		Reader:  reader,
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("InstallationGate() error = %v", err)
	}
	e.Use(gate)
	e.GET("/api/setup/state", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/setup/state", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	if reader.calls != 0 {
		t.Fatalf("installation state calls = %d, want 0 for skipped setup route", reader.calls)
	}
}

func TestInstallationGateAllowsPrivateRoutesAfterInitialization(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	reader := &installationStateReader{initialized: true}
	gate, err := InstallationGate(&InstallationGateConfig{
		Reader:  reader,
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("InstallationGate() error = %v", err)
	}
	e.Use(gate)
	e.GET("/api/auth/me", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	if reader.calls != 1 {
		t.Fatalf("installation state calls = %d, want 1", reader.calls)
	}
}

func TestInstallationGateRequiresReaderWhenEnabled(t *testing.T) {
	gate, err := InstallationGate(&InstallationGateConfig{Enabled: true})

	assert.Error(t, err)
	assert.Nil(t, gate)
}

func TestApplyMiddlewares(t *testing.T) {
	e := echo.New()
	config := DefaultMiddlewareConfig()

	assert.NoError(t, ApplyMiddlewares(e, config))
	assert.NotNil(t, e)
}

func TestInstallationGateRunsBeforeAPIKeyAuthentication(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	verifier := &countingAPIKeyVerifier{}
	config := DefaultMiddlewareConfig()
	config.EnableInstallationGate = true
	config.InstallationGate = &InstallationGateConfig{
		Reader:  &installationStateReader{initialized: false},
		Enabled: true,
	}
	config.EnableAPIKey = true
	config.APIKey = &APIKeyConfig{
		Header:   APIKeyHeader,
		Verifier: verifier,
		Enabled:  true,
	}

	assert.NoError(t, ApplyMiddlewares(e, config))
	e.GET("/private", func(c *echo.Context) error {
		t.Fatal("private handler reached before installation completed")
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set(APIKeyHeader, "ea_known_secret")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assertErrorPayload(t, rec, apperr.ErrSystemUninitialized, "system is not initialized")
	if verifier.calls != 0 {
		t.Fatalf("api key verifier calls = %d, want 0 before installation completes", verifier.calls)
	}
}

func TestAPIKeyAuthenticationRunsBeforeLoginSession(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	config := DefaultMiddlewareConfig()
	config.EnableAPIKey = true
	config.APIKey = &APIKeyConfig{
		Header:   APIKeyHeader,
		Verifier: staticAPIKeyVerifier{},
		Enabled:  true,
	}
	config.EnableLoginSession = true
	config.LoginSession = &LoginSessionConfig{
		CookieName:    LoginSessionCookieName,
		Authenticator: staticLoginSessionAuthenticator{identity: LoginSessionIdentity{SessionID: "session-9", UserID: "99", RoleID: "9"}},
		Enabled:       true,
	}

	assert.NoError(t, ApplyMiddlewares(e, config))
	e.GET("/private", func(c *echo.Context) error {
		if got := requestctx.GetUserID(c.Request().Context()); got != "42" {
			t.Fatalf("UserID = %q, want 42", got)
		}
		if got := requestctx.GetRoleID(c.Request().Context()); got != "7" {
			t.Fatalf("RoleID = %q, want 7", got)
		}
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set(APIKeyHeader, "ea_known_secret")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAPIKeyAuthenticationRejectsInvalidToken(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	config := DefaultMiddlewareConfig()
	config.EnableAPIKey = true
	config.APIKey = &APIKeyConfig{
		Header:   APIKeyHeader,
		Verifier: staticAPIKeyVerifier{},
		Enabled:  true,
	}

	assert.NoError(t, ApplyMiddlewares(e, config))
	e.GET("/private", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set(APIKeyHeader, "wrong")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequestLoggerPreservesRenderedApplicationErrorStatus(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	config := DefaultMiddlewareConfig()

	assert.NoError(t, ApplyMiddlewares(e, config))
	e.GET("/bad", func(_ *echo.Context) error {
		return apperr.NewBadRequest("invalid payload")
	})

	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertErrorPayload(t, rec, apperr.ErrBadRequest, "bad request: invalid payload")
}

func TestRequestContextMiddlewareStoresMetadata(t *testing.T) {
	e := echo.New()
	e.Use(RequestContext())
	e.GET("/ping", func(c *echo.Context) error {
		info, ok := requestctx.FromContext(c.Request().Context())
		if !ok {
			t.Fatal("request context metadata missing")
		}
		if info.RequestID != testRequestID {
			t.Fatalf("RequestID = %q, want %s", info.RequestID, testRequestID)
		}
		if info.TraceID != "trace-123" {
			t.Fatalf("TraceID = %q, want trace-123", info.TraceID)
		}
		if info.UserID != "" {
			t.Fatalf("UserID = %q, want empty because request metadata does not authenticate users", info.UserID)
		}
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(headerRequestID, testRequestID)
	req.Header.Set(headerTraceID, "trace-123")
	req.Header.Set("X-User-ID", "user-123")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, testRequestID, rec.Header().Get(headerRequestID))
}

func TestRequestContextMiddlewareCleansUntrustedMetadata(t *testing.T) {
	overlongRequestID := strings.Repeat("a", 129)
	overlongSpanID := strings.Repeat("b", 129)

	e := echo.New()
	e.Use(RequestContext())
	e.GET("/ping", func(c *echo.Context) error {
		info, ok := requestctx.FromContext(c.Request().Context())
		if !ok {
			t.Fatal("request context metadata missing")
		}
		if info.RequestID == "" {
			t.Fatal("RequestID is empty, want generated request id")
		}
		if info.RequestID == overlongRequestID {
			t.Fatal("RequestID used untrusted overlong header")
		}
		if info.TraceID != "" {
			t.Fatalf("TraceID = %q, want invalid trace id to be dropped", info.TraceID)
		}
		if info.SpanID != "" {
			t.Fatalf("SpanID = %q, want invalid span id to be dropped", info.SpanID)
		}
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(headerRequestID, overlongRequestID)
	req.Header.Set(headerTraceID, "trace with space")
	req.Header.Set(headerSpanID, overlongSpanID)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.NotEmpty(t, rec.Header().Get(headerRequestID))
	assert.NotEqual(t, overlongRequestID, rec.Header().Get(headerRequestID))
}

func TestRequestLoggerIncludesActiveTraceMetadata(t *testing.T) {
	tracerProvider := trace.NewTracerProvider()
	defer func() {
		assert.NoError(t, tracerProvider.Shutdown(context.Background()))
	}()

	ctx := requestctx.WithInfo(context.Background(), requestctx.Info{RequestID: testRequestID})
	base := zerolog.New(&bytes.Buffer{})
	ctx = base.WithContext(ctx)
	ctx, span := tracerProvider.Tracer("test").Start(ctx, "request")
	defer span.End()

	logger := requestLoggerFromContext(ctx)
	var buf bytes.Buffer
	logger = logger.Output(&buf)
	logger.Info().Msg("request")

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("decode log event: %v", err)
	}
	if got["request_id"] != testRequestID {
		t.Fatalf("request_id = %v, want %s", got["request_id"], testRequestID)
	}
	if got["trace_id"] == "" {
		t.Fatal("trace_id is empty, want active span trace id")
	}
	if got["span_id"] == "" {
		t.Fatal("span_id is empty, want active span id")
	}
}

func TestRequestLoggerOmitsTraceMetadataWithoutActiveSpan(t *testing.T) {
	ctx := requestctx.WithInfo(context.Background(), requestctx.Info{RequestID: testRequestID})
	base := zerolog.New(&bytes.Buffer{})
	ctx = base.WithContext(ctx)

	logger := requestLoggerFromContext(ctx)
	var buf bytes.Buffer
	logger = logger.Output(&buf)
	logger.Info().Msg("request")

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("decode log event: %v", err)
	}
	if _, ok := got["trace_id"]; ok {
		t.Fatal("trace_id present without active span")
	}
	if _, ok := got["span_id"]; ok {
		t.Fatal("span_id present without active span")
	}
}

func TestLoginSessionStoresIdentityInAuthenticatedContext(t *testing.T) {
	sessionMiddleware, err := LoginSession(&LoginSessionConfig{
		CookieName: LoginSessionCookieName,
		Authenticator: staticLoginSessionAuthenticator{identity: LoginSessionIdentity{
			SessionID: "session-123",
			UserID:    "user-123",
			RoleID:    "role-456",
		}},
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("LoginSession() error = %v", err)
	}

	e := echo.New()
	e.Use(RequestContext())
	e.Use(sessionMiddleware)
	e.GET("/me", func(c *echo.Context) error {
		if got := requestctx.GetUserID(c.Request().Context()); got != "user-123" {
			t.Fatalf("GetUserID() = %q, want user-123", got)
		}
		if got := requestctx.GetRoleID(c.Request().Context()); got != "role-456" {
			t.Fatalf("GetRoleID() = %q, want role-456", got)
		}
		if got := requestctx.GetLoginSessionID(c.Request().Context()); got != "session-123" {
			t.Fatalf("GetLoginSessionID() = %q, want session-123", got)
		}
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: LoginSessionCookieName, Value: "opaque-token"})
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestLoginSessionRejectsAuthenticatorError(t *testing.T) {
	sessionMiddleware, err := LoginSession(&LoginSessionConfig{
		CookieName:    LoginSessionCookieName,
		Authenticator: staticLoginSessionAuthenticator{err: apperr.NewUnauthorized()},
		Enabled:       true,
	})
	if err != nil {
		t.Fatalf("LoginSession() error = %v", err)
	}

	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	e.Use(RequestContext())
	e.Use(sessionMiddleware)
	e.GET("/me", func(c *echo.Context) error {
		t.Fatal("handler reached with rejected login session")
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: LoginSessionCookieName, Value: "opaque-token"})
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assertErrorPayload(t, rec, apperr.ErrUnauthorized, "unauthorized")
}

func TestLoginSessionMissingCookieReturnsGenericUnauthorized(t *testing.T) {
	sessionMiddleware, err := LoginSession(&LoginSessionConfig{
		CookieName:    LoginSessionCookieName,
		Authenticator: staticLoginSessionAuthenticator{identity: LoginSessionIdentity{SessionID: "session-1", UserID: "42", RoleID: "7"}},
		Enabled:       true,
	})
	if err != nil {
		t.Fatalf("LoginSession() error = %v", err)
	}

	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	e.Use(RequestContext())
	e.Use(sessionMiddleware)
	e.GET("/me", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assertErrorPayload(t, rec, apperr.ErrUnauthorized, "unauthorized")
}

func TestCSRFProtectsLoginSessionUnsafeRequests(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	e.Use(RequestContext())
	sessionMiddleware, err := LoginSession(&LoginSessionConfig{
		CookieName:    LoginSessionCookieName,
		Authenticator: staticLoginSessionAuthenticator{identity: LoginSessionIdentity{SessionID: "session-1", UserID: "42", RoleID: "7"}},
		Enabled:       true,
	})
	if err != nil {
		t.Fatalf("LoginSession() error = %v", err)
	}
	e.Use(sessionMiddleware)
	e.Use(middleware.CSRFWithConfig(CSRFConfig(nil, false)))
	e.POST("/change", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/change", nil)
	req.AddCookie(&http.Cookie{Name: LoginSessionCookieName, Value: "opaque-token"})
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestErrorRecovery(t *testing.T) {
	e := echo.New()
	e.Use(ErrorRecovery())
	e.HTTPErrorHandler = ErrorHandler

	// 创建一个会panic的路由
	e.GET("/panic", func(_ *echo.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	// 测试panic恢复
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assertErrorPayload(t, rec, apperr.ErrInternalServer, "Internal server error")
}

func TestErrorHandlerNormalizesApplicationAndBadRequestErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   int
		wantMsg    string
	}{
		{
			name:       "plain error becomes internal server error",
			err:        errors.New("database password leaked in raw error"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   apperr.ErrInternalServer,
			wantMsg:    "Internal server error",
		},
		{
			name:       "echo bad request uses generic public message",
			err:        echo.NewHTTPError(http.StatusBadRequest, "invalid query"),
			wantStatus: http.StatusBadRequest,
			wantCode:   apperr.ErrBadRequest,
			wantMsg:    "Bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrorHandlerNormalizes(t, tt.err, tt.wantStatus, tt.wantCode, tt.wantMsg)
		})
	}
}

func TestErrorHandlerNormalizesEchoHTTPClientErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   int
		wantMsg    string
	}{
		{
			name:       "echo method not allowed keeps client status",
			err:        echo.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed"),
			wantStatus: http.StatusMethodNotAllowed,
			wantCode:   apperr.ErrMethodNotAllowed,
			wantMsg:    "Method not allowed",
		},
		{
			name:       "echo status coder method not allowed keeps client status",
			err:        echo.ErrMethodNotAllowed,
			wantStatus: http.StatusMethodNotAllowed,
			wantCode:   apperr.ErrMethodNotAllowed,
			wantMsg:    "Method not allowed",
		},
		{
			name:       "echo conflict keeps conflict status",
			err:        echo.NewHTTPError(http.StatusConflict, "already exists"),
			wantStatus: http.StatusConflict,
			wantCode:   apperr.ErrConflict,
			wantMsg:    "Conflict",
		},
		{
			name:       "unknown echo client error remains client error",
			err:        echo.NewHTTPError(http.StatusUnsupportedMediaType, "unsupported media type"),
			wantStatus: http.StatusBadRequest,
			wantCode:   apperr.ErrBadRequest,
			wantMsg:    "Bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrorHandlerNormalizes(t, tt.err, tt.wantStatus, tt.wantCode, tt.wantMsg)
		})
	}
}

func TestErrorHandlerRecordsInternalSystemError(t *testing.T) {
	recorder := &systemErrorRecorderSpy{}
	handler := ErrorHandlerWithRecorder(recorder)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set(headerRequestID, testRequestID)
	req.Header.Set("User-Agent", "test-agent")
	req = req.WithContext(requestctx.WithUserID(req.Context(), "42"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users")

	handler(c, errors.New("database failed"))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	if len(recorder.records) != 1 {
		t.Fatalf("system error records = %d, want 1", len(recorder.records))
	}
	got := recorder.records[0]
	if got.Code != apperr.ErrInternalServer {
		t.Fatalf("Code = %d, want %d", got.Code, apperr.ErrInternalServer)
	}
	if got.Path != "/users" {
		t.Fatalf("Path = %q, want /users", got.Path)
	}
	if got.RequestID != testRequestID {
		t.Fatalf("RequestID = %q, want %s", got.RequestID, testRequestID)
	}
	if got.UserID != "42" {
		t.Fatalf("UserID = %q, want 42", got.UserID)
	}
}

func TestErrorHandlerDoesNotRecordClientErrors(t *testing.T) {
	recorder := &systemErrorRecorderSpy{}
	handler := ErrorHandlerWithRecorder(recorder)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler(c, echo.NewHTTPError(http.StatusBadRequest, "bad query"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	if len(recorder.records) != 0 {
		t.Fatalf("system error records = %d, want 0", len(recorder.records))
	}
}

func TestErrorHandlerSkipsCommittedResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().WriteHeader(http.StatusAccepted)

	ErrorHandler(c, errors.New("late error after response"))

	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestValidatorReturnsFieldValidationError(t *testing.T) {
	type request struct {
		Email string `validate:"required,email"`
	}

	err := (&Validator{Validator: validator.New()}).Validate(request{})
	if err == nil {
		t.Fatal("Validate() error = nil, want validation error")
	}

	appErr, ok := apperr.Parse(err)
	if !ok {
		t.Fatal("Validate() did not return application error")
	}
	if appErr.Code() != apperr.ErrValidation {
		t.Fatalf("Code = %d, want %d", appErr.Code(), apperr.ErrValidation)
	}
	if appErr.Message() != "Email is required" {
		t.Fatalf("Message = %q, want Email is required", appErr.Message())
	}
	if appErr.Detail() == "" {
		t.Fatal("Detail is empty, want validation detail")
	}
}

func TestValidatorReturnsInternalErrorWhenMisconfigured(t *testing.T) {
	err := (&Validator{}).Validate(struct{}{})
	if err == nil {
		t.Fatal("Validate() error = nil, want misconfiguration error")
	}

	appErr, ok := apperr.Parse(err)
	if !ok {
		t.Fatal("Validate() did not return application error")
	}
	if appErr.Code() != apperr.ErrInternalServer {
		t.Fatalf("Code = %d, want %d", appErr.Code(), apperr.ErrInternalServer)
	}
	if appErr.Message() != "Internal server error" {
		t.Fatalf("Message = %q, want Internal server error", appErr.Message())
	}
}

func TestNilValidatorReturnsInternalErrorWhenMisconfigured(t *testing.T) {
	var cv *Validator

	err := cv.Validate(struct{}{})
	if err == nil {
		t.Fatal("Validate() error = nil, want misconfiguration error")
	}
}

func assertErrorPayload(t *testing.T, rec *httptest.ResponseRecorder, wantCode int, wantMessage string) {
	t.Helper()

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if got["code"] != float64(wantCode) {
		t.Fatalf("code = %v, want %d", got["code"], wantCode)
	}
	if got["message"] != wantMessage {
		t.Fatalf("message = %v, want %q", got["message"], wantMessage)
	}
}

type staticAPIKeyVerifier struct{}

func (staticAPIKeyVerifier) VerifyAPIKey(_ context.Context, secret string) (APIKeyIdentity, error) {
	if secret != "ea_known_secret" {
		return APIKeyIdentity{}, apperr.NewUnauthorized()
	}
	return APIKeyIdentity{UserID: "42", RoleID: "7"}, nil
}

type countingAPIKeyVerifier struct {
	calls int
}

func (v *countingAPIKeyVerifier) VerifyAPIKey(_ context.Context, secret string) (APIKeyIdentity, error) {
	v.calls++
	if secret != "ea_known_secret" {
		return APIKeyIdentity{}, apperr.NewUnauthorized()
	}
	return APIKeyIdentity{UserID: "42", RoleID: "7"}, nil
}

type staticLoginSessionAuthenticator struct {
	identity LoginSessionIdentity
	err      error
}

func (a staticLoginSessionAuthenticator) AuthenticateLoginSession(context.Context, string) (LoginSessionIdentity, error) {
	if a.err != nil {
		return LoginSessionIdentity{}, a.err
	}
	return a.identity, nil
}

type installationStateReader struct {
	initialized bool
	calls       int
}

func (r *installationStateReader) Initialized(context.Context) (bool, error) {
	r.calls++
	return r.initialized, nil
}

type systemErrorRecorderSpy struct {
	records []SystemErrorInput
}

func (r *systemErrorRecorderSpy) RecordSystemError(_ context.Context, input SystemErrorInput) error {
	r.records = append(r.records, input)
	return nil
}

func assertErrorHandlerNormalizes(
	t *testing.T,
	err error,
	wantStatus int,
	wantCode int,
	wantMsg string,
) {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ErrorHandler(c, err)

	assert.Equal(t, wantStatus, rec.Code)
	assertErrorPayload(t, rec, wantCode, wantMsg)
}

func TestLoginSessionConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *LoginSessionConfig
	}{
		{
			name: "enabled login session",
			config: &LoginSessionConfig{
				CookieName:    LoginSessionCookieName,
				SkipPaths:     []string{"/api/health"},
				Authenticator: staticLoginSessionAuthenticator{identity: LoginSessionIdentity{SessionID: "session-1", UserID: "42", RoleID: "7"}},
				Enabled:       true,
			},
		},
		{
			name: "disabled login session",
			config: &LoginSessionConfig{
				CookieName: LoginSessionCookieName,
				SkipPaths:  []string{},
				Enabled:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionMiddleware, err := LoginSession(tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, sessionMiddleware)
		})
	}
}

func TestLoginSessionRequiresAuthenticatorWhenEnabled(t *testing.T) {
	sessionMiddleware, err := LoginSession(&LoginSessionConfig{Enabled: true})
	assert.Error(t, err)
	assert.Nil(t, sessionMiddleware)
}

func TestApplyMiddlewaresRequiresCORSOriginsWhenEnabled(t *testing.T) {
	e := echo.New()
	config := DefaultMiddlewareConfig()
	config.EnableCORS = true

	err := ApplyMiddlewares(e, config)

	assert.Error(t, err)
}

func TestApplyMiddlewaresRejectsWildcardCORSOriginWithCredentials(t *testing.T) {
	e := echo.New()
	config := DefaultMiddlewareConfig()
	config.EnableCORS = true
	config.CORS.AllowOrigins = []string{"*"}
	config.CORS.AllowCredentials = true

	err := ApplyMiddlewares(e, config)

	assert.Error(t, err)
}

func TestRequestPathOmitsRawQuery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users?token=secret", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users")

	if got := requestPath(c); got != "/users" {
		t.Fatalf("requestPath() = %q, want /users", got)
	}
}

func TestMiddlewareConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *MiddlewareConfig
	}{
		{
			name:   "default config",
			config: DefaultMiddlewareConfig(),
		},
		{
			name: "custom config",
			config: &MiddlewareConfig{
				EnableRecovery:       true,
				EnableRequestContext: true,
				EnableLogger:         true,
				EnableGzip:           true,
				EnableCORS:           true,
				CORS: middleware.CORSConfig{
					AllowOrigins: []string{"https://example.com"},
				},
				EnableLoginSession: false,
				LoginSession:       DefaultLoginSessionConfig(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			assert.NoError(t, ApplyMiddlewares(e, tt.config))
			assert.NotNil(t, e)
		})
	}
}
