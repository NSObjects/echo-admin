package boot

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestFinalizeRouteExposureAllowsSystemAndBootstrapRoutesWithoutAuthorizer(t *testing.T) {
	e := echo.New()
	handler := func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}
	for _, route := range []apiRouteKey{
		{method: http.MethodGet, pattern: "/api/health"},
		{method: http.MethodGet, pattern: "/api/info"},
		{method: http.MethodGet, pattern: "/api/ready"},
		{method: http.MethodGet, pattern: "/api/capabilities"},
		{method: http.MethodGet, pattern: "/api/setup/state"},
		{method: http.MethodPost, pattern: "/api/setup"},
		{method: http.MethodPost, pattern: "/api/auth/login"},
		{method: http.MethodOptions, pattern: "/api/things"},
	} {
		e.Add(route.method, route.pattern, handler)
	}

	manifest, err := finalizeRouteExposure(e, nil)
	if err != nil {
		t.Fatalf("finalizeRouteExposure() error = %v, want nil", err)
	}
	if got := len(manifest.managed); got != 0 {
		t.Fatalf("managed route count = %d, want 0", got)
	}
}

func TestFinalizeRouteExposureTreatsExemptionVariantsAsManaged(t *testing.T) {
	tests := []apiRouteKey{
		{method: http.MethodGet, pattern: "/api/auth/login"},
		{method: http.MethodPost, pattern: "/api/auth/login/:suffix"},
		{method: http.MethodPost, pattern: "/api/setup/state"},
	}
	for _, route := range tests {
		t.Run(route.String(), func(t *testing.T) {
			e := echo.New()
			e.Add(route.method, route.pattern, func(c *echo.Context) error {
				return c.NoContent(http.StatusNoContent)
			})

			_, err := finalizeRouteExposure(e, nil)
			if !errors.Is(err, errRouteAuthorizationRequired) {
				t.Fatalf("finalizeRouteExposure() error = %v, want %v", err, errRouteAuthorizationRequired)
			}
		})
	}
}

func TestFinalizeRouteExposureLeavesOptionsAndUnmatchedRequestsOutsideAuthorization(t *testing.T) {
	e := echo.New()
	e.GET("/api/things/:id", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	authorizer := &testRouteAuthorizer{}
	if _, err := finalizeRouteExposure(e, authorizer); err != nil {
		t.Fatalf("finalizeRouteExposure() error = %v", err)
	}

	for _, tt := range []struct {
		name   string
		method string
		path   string
	}{
		{name: "options", method: http.MethodOptions, path: "/api/things/42"},
		{name: "not found", method: http.MethodGet, path: "/api/missing"},
		{name: "method not allowed", method: http.MethodDelete, path: "/api/things/42"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, nil)
			e.ServeHTTP(recorder, request)
			if got := len(authorizer.calls); got != 0 {
				t.Fatalf("authorizer calls = %d, want 0", got)
			}
		})
	}
}
