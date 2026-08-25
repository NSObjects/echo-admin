package boot

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/labstack/echo/v5"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
)

var (
	errRouteAuthorizationRequired = errors.New("route authorizer is required")
	errRouteExposureMismatch      = errors.New("route exposure policy mismatch")
	errManagedAPIRouteCoverage    = errors.New("managed api route coverage mismatch")
)

type apiRouteKey struct {
	method  string
	pattern string
}

func (r apiRouteKey) String() string {
	return r.method + " " + r.pattern
}

type apiRouteClass uint8

const (
	apiRouteOutside apiRouteClass = iota
	apiRouteSystem
	apiRouteBootstrap
	apiRouteManaged
)

type routeManifest struct {
	managed map[apiRouteKey]struct{}
}

func finalizeRouteExposure(e *echo.Echo, authorizer routeAuthorizer) (routeManifest, error) {
	if e == nil {
		return routeManifest{}, errors.New("finalize route exposure: nil echo")
	}
	manifest := routeManifest{managed: make(map[apiRouteKey]struct{})}
	for _, route := range e.Router().Routes() {
		key := apiRouteKey{method: route.Method, pattern: route.Path}
		if classifyAPIRoute(key) == apiRouteManaged {
			manifest.managed[key] = struct{}{}
		}
	}
	if len(manifest.managed) > 0 && authorizer == nil {
		return routeManifest{}, fmt.Errorf("%w for managed routes: %s", errRouteAuthorizationRequired, strings.Join(manifest.managedKeys(), ", "))
	}

	// Echo's global middleware chain is rebuilt by Use, so finalization can
	// happen after all route registrars without relying on registration order.
	e.Use(routeExposureMiddleware(manifest, authorizer))
	return manifest, nil
}

func routeExposureMiddleware(manifest routeManifest, authorizer routeAuthorizer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c == nil || c.Request() == nil || c.Request().Method == http.MethodOptions {
				return next(c)
			}
			info := c.RouteInfo()
			key := apiRouteKey{method: info.Method, pattern: info.Path}
			if classifyAPIRoute(key) != apiRouteManaged {
				return next(c)
			}
			if _, ok := manifest.managed[key]; !ok || authorizer == nil {
				return apperr.WrapInternal(errRouteExposureMismatch, "route exposure policy mismatch")
			}
			if err := authorizer.AuthorizeRoute(c.Request().Context(), key.method, key.pattern); err != nil {
				return err
			}
			return next(c)
		}
	}
}

func classifyAPIRoute(route apiRouteKey) apiRouteClass {
	if route.method == "" || route.pattern == "" || route.method == http.MethodOptions || !isAPIPath(route.pattern) {
		return apiRouteOutside
	}
	switch route {
	case apiRouteKey{method: http.MethodGet, pattern: "/api/health"},
		apiRouteKey{method: http.MethodGet, pattern: "/api/info"},
		apiRouteKey{method: http.MethodGet, pattern: "/api/ready"},
		apiRouteKey{method: http.MethodGet, pattern: "/api/capabilities"}:
		return apiRouteSystem
	case apiRouteKey{method: http.MethodGet, pattern: "/api/setup/state"},
		apiRouteKey{method: http.MethodPost, pattern: "/api/setup"},
		apiRouteKey{method: http.MethodPost, pattern: "/api/auth/login"}:
		return apiRouteBootstrap
	default:
		return apiRouteManaged
	}
}

func isAPIPath(pattern string) bool {
	return pattern == "/api" || strings.HasPrefix(pattern, "/api/")
}

func (m routeManifest) managedKeys() []string {
	keys := make([]string, 0, len(m.managed))
	for route := range m.managed {
		keys = append(keys, route.String())
	}
	sort.Strings(keys)
	return keys
}

// CheckManagedCatalog verifies the test-time contract between registered
// Managed API Routes and the access-owned deployment definition.
func (m routeManifest) CheckManagedCatalog(definitions []accessdomain.ManagedAPIRouteDefinition) error {
	catalog := make(map[apiRouteKey]struct{}, len(definitions))
	diff := routeCoverageDiff{}
	for _, definition := range definitions {
		route := apiRouteKey{method: definition.Method, pattern: definition.Pattern}
		if _, exists := catalog[route]; exists {
			diff.duplicates = append(diff.duplicates, route.String())
			continue
		}
		catalog[route] = struct{}{}
		if classifyAPIRoute(route) != apiRouteManaged {
			diff.wronglyClassified = append(diff.wronglyClassified, route.String())
			delete(catalog, route)
		}
	}
	for route := range m.managed {
		if _, ok := catalog[route]; !ok {
			diff.missing = append(diff.missing, route.String())
		}
	}
	for route := range catalog {
		if _, ok := m.managed[route]; !ok {
			diff.stale = append(diff.stale, route.String())
		}
	}
	return diff.err()
}

type routeCoverageDiff struct {
	duplicates        []string
	wronglyClassified []string
	missing           []string
	stale             []string
}

func (d routeCoverageDiff) err() error {
	if len(d.duplicates)+len(d.wronglyClassified)+len(d.missing)+len(d.stale) == 0 {
		return nil
	}
	sort.Strings(d.duplicates)
	sort.Strings(d.wronglyClassified)
	sort.Strings(d.missing)
	sort.Strings(d.stale)
	parts := make([]string, 0, 4)
	parts = appendRouteCoveragePart(parts, "duplicate", d.duplicates)
	parts = appendRouteCoveragePart(parts, "wrongly_classified", d.wronglyClassified)
	parts = appendRouteCoveragePart(parts, "missing_from_catalog", d.missing)
	parts = appendRouteCoveragePart(parts, "stale_in_catalog", d.stale)
	return fmt.Errorf("%w: %s", errManagedAPIRouteCoverage, strings.Join(parts, " "))
}

func appendRouteCoveragePart(parts []string, name string, routes []string) []string {
	if len(routes) == 0 {
		return parts
	}
	return append(parts, name+"=["+strings.Join(routes, ", ")+"]")
}
