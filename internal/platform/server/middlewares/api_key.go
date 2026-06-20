package middlewares

import (
	"context"
	"errors"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/infrastructure/logging"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
)

// APIKeyHeader is the header used for back-office API token authentication.
const APIKeyHeader = "X-API-Token"

// APIKeyIdentity is the request identity produced by a verified API token.
type APIKeyIdentity struct {
	UserID string
	RoleID string
}

// APIKeyVerifier validates a raw API token without exposing token storage to the server.
type APIKeyVerifier interface {
	VerifyAPIKey(context.Context, string) (APIKeyIdentity, error)
}

// APIKeyConfig controls API token authentication middleware.
type APIKeyConfig struct {
	Header   string
	Verifier APIKeyVerifier
	Enabled  bool
}

// DefaultAPIKeyConfig returns disabled API token authentication defaults.
func DefaultAPIKeyConfig() *APIKeyConfig {
	return &APIKeyConfig{Header: APIKeyHeader}
}

// APIKey creates middleware that authenticates requests carrying an API token.
func APIKey(config *APIKeyConfig) (echo.MiddlewareFunc, error) {
	if config == nil || !config.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}, nil
	}
	if config.Verifier == nil {
		return nil, errors.New("api key verifier is required when api key authentication is enabled")
	}
	header := strings.TrimSpace(config.Header)
	if header == "" {
		header = APIKeyHeader
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			raw := strings.TrimSpace(c.Request().Header.Get(header))
			if raw == "" {
				return next(c)
			}
			identity, err := config.Verifier.VerifyAPIKey(c.Request().Context(), raw)
			if err != nil {
				return err
			}
			if identity.UserID == "" || identity.RoleID == "" {
				return apperr.NewUnauthorized()
			}
			request := c.Request()
			ctx := requestctx.WithRoleID(requestctx.WithUserID(request.Context(), identity.UserID), identity.RoleID)
			logger := logging.FromContext(ctx).With().
				Str("user_id", identity.UserID).
				Str("role_id", identity.RoleID).
				Str("auth", "api_token").
				Logger()
			c.SetRequest(request.WithContext(logger.WithContext(ctx)))
			return next(c)
		}
	}, nil
}
