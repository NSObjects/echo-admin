package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/NSObjects/echo-admin/internal/platform/apperr"
)

// InstallationStateReader reports whether first initialization has completed.
type InstallationStateReader interface {
	Initialized(context.Context) (bool, error)
}

// InstallationGateConfig controls the uninitialized-system route gate.
type InstallationGateConfig struct {
	Reader     InstallationStateReader
	Exemptions []RouteExemption
	Enabled    bool
}

// InstallationGate blocks normal administration routes until setup completes.
// Exemptions are injected by the composition root and matched by exact method
// and registered pattern, so this layer carries no route policy of its own.
func InstallationGate(config *InstallationGateConfig) (echo.MiddlewareFunc, error) {
	if config == nil || !config.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}, nil
	}
	if config.Reader == nil {
		return nil, errors.New("installation state reader is required when installation gate is enabled")
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.Request().Method == http.MethodOptions || routeExempt(c, config.Exemptions) {
				return next(c)
			}
			initialized, err := config.Reader.Initialized(c.Request().Context())
			if err != nil {
				return err
			}
			if !initialized {
				return apperr.New(apperr.ErrSystemUninitialized, "")
			}
			return next(c)
		}
	}, nil
}
