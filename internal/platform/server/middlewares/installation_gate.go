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
	Reader    InstallationStateReader
	SkipPaths []string
	Enabled   bool
}

// DefaultInstallationGateConfig returns the public setup and system routes that
// remain reachable before first initialization completes.
func DefaultInstallationGateConfig() *InstallationGateConfig {
	return &InstallationGateConfig{
		SkipPaths: []string{
			"/api/health",
			"/api/info",
			"/api/setup/state",
			"/api/setup",
		},
	}
}

// InstallationGate blocks normal administration routes until setup completes.
func InstallationGate(config *InstallationGateConfig) (echo.MiddlewareFunc, error) {
	if config == nil || !config.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}, nil
	}
	if config.Reader == nil {
		return nil, errors.New("installation state reader is required when installation gate is enabled")
	}
	skipPaths := config.SkipPaths
	if len(skipPaths) == 0 {
		skipPaths = DefaultInstallationGateConfig().SkipPaths
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.Request().Method == http.MethodOptions || pathSkipped(c, skipPaths) {
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
