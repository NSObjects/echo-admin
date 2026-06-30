package boot

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
)

type routeAuthorizer interface {
	AuthorizeRoute(context.Context, string, string) error
}

func (a *App) installAPIAuthorization() error {
	authorizer, err := do.InvokeAs[routeAuthorizer](a.injector)
	if err != nil {
		if optionalServiceMissing(err) {
			return nil
		}
		return err
	}
	a.server.API().Use(routeAuthorizationMiddleware(authorizer))
	return nil
}

func routeAuthorizationMiddleware(authorizer routeAuthorizer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if shouldSkipRouteAuthorization(c) {
				return next(c)
			}
			if err := authorizer.AuthorizeRoute(c.Request().Context(), c.Request().Method, c.Path()); err != nil {
				return err
			}
			return next(c)
		}
	}
}

func shouldSkipRouteAuthorization(c *echo.Context) bool {
	if c == nil || c.Request() == nil {
		return false
	}
	method := c.Request().Method
	if method == http.MethodOptions {
		return true
	}
	path := c.Path()
	switch {
	case method == http.MethodPost && path == "/api/auth/login":
		return true
	case method == http.MethodGet && path == "/api/setup/state":
		return true
	case method == http.MethodPost && path == "/api/setup":
		return true
	default:
		return false
	}
}
