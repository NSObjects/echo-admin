package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/infrastructure/logging"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
)

const (
	// LoginSessionCookieName is the HttpOnly browser credential for login
	// sessions.
	LoginSessionCookieName = "login_session"
	// CSRFCookieName is the browser-readable double-submit CSRF cookie.
	CSRFCookieName = "csrf_token"

	csrfTokenBytes = 32
)

// LoginSessionIdentity is the request identity produced by a verified browser
// login session. Fields are int64 database keys.
type LoginSessionIdentity struct {
	SessionID int64
	UserID    int64
	RoleID    int64
}

// LoginSessionAuthenticator validates browser login session credentials.
type LoginSessionAuthenticator interface {
	AuthenticateLoginSession(context.Context, string) (LoginSessionIdentity, error)
}

// LoginSessionConfig controls browser login-session authentication.
type LoginSessionConfig struct {
	CookieName    string
	Exemptions    []RouteExemption
	Authenticator LoginSessionAuthenticator
	Enabled       bool
}

// DefaultLoginSessionConfig returns disabled login-session authentication
// defaults. Route exemptions are composition-root policy and must be injected
// by the caller; this layer defaults to none.
func DefaultLoginSessionConfig() *LoginSessionConfig {
	return &LoginSessionConfig{
		CookieName: LoginSessionCookieName,
	}
}

// LoginSession creates browser login-session authentication middleware.
func LoginSession(config *LoginSessionConfig) (echo.MiddlewareFunc, error) {
	if config == nil || !config.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}, nil
	}
	if config.Authenticator == nil {
		return nil, errors.New("login session authenticator is required when login sessions are enabled")
	}
	cookieName := strings.TrimSpace(config.CookieName)
	if cookieName == "" {
		cookieName = LoginSessionCookieName
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if requestctx.GetUserID(c.Request().Context()) != 0 || routeExempt(c, config.Exemptions) {
				return next(c)
			}
			cookie, err := c.Cookie(cookieName)
			if err != nil || strings.TrimSpace(cookie.Value) == "" {
				return apperr.NewUnauthorized()
			}
			identity, err := config.Authenticator.AuthenticateLoginSession(c.Request().Context(), cookie.Value)
			if err != nil {
				return err
			}
			if identity.UserID <= 0 || identity.RoleID <= 0 || identity.SessionID <= 0 {
				return apperr.NewUnauthorized()
			}
			request := c.Request()
			ctx := requestctx.WithLoginSessionID(
				requestctx.WithRoleID(
					requestctx.WithUserID(request.Context(), identity.UserID),
					identity.RoleID,
				),
				identity.SessionID,
			)
			logger := logging.FromContext(ctx).With().
				Int64("user_id", identity.UserID).
				Int64("role_id", identity.RoleID).
				Int64("login_session_id", identity.SessionID).
				Str("auth", "login_session").
				Logger()
			c.SetRequest(request.WithContext(logger.WithContext(ctx)))
			return next(c)
		}
	}, nil
}

// CSRFConfig returns the Echo CSRF middleware configuration used for browser
// login-session requests. It shares the login-session exemptions because a
// route without a login session has no CSRF obligation.
func CSRFConfig(exemptions []RouteExemption, secureCookies bool) middleware.CSRFConfig {
	return middleware.CSRFConfig{
		Skipper: func(c *echo.Context) bool {
			return requestctx.GetLoginSessionID(c.Request().Context()) == 0 || routeExempt(c, exemptions)
		},
		TokenLookup:    "header:" + echo.HeaderXCSRFToken,
		CookieName:     CSRFCookieName,
		CookiePath:     "/",
		CookieMaxAge:   int((12 * time.Hour).Seconds()),
		CookieSecure:   secureCookies,
		CookieHTTPOnly: false,
		CookieSameSite: http.SameSiteLaxMode,
	}
}

// SetLoginSessionCookie stores the opaque browser session credential.
func SetLoginSessionCookie(c *echo.Context, token string, expiresAt time.Time, secure bool) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	c.SetCookie(&http.Cookie{
		Name:     LoginSessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expiresAt,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearLoginSessionCookie removes the browser session credential.
func ClearLoginSessionCookie(c *echo.Context, secure bool) {
	clearCookie(c, LoginSessionCookieName, true, secure)
}

// NewCSRFToken creates a browser-readable CSRF token for login responses.
func NewCSRFToken() (string, error) {
	token := make([]byte, csrfTokenBytes)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token), nil
}

// SetCSRFCookie stores the browser-readable CSRF token used by Echo's CSRF
// middleware.
func SetCSRFCookie(c *echo.Context, token string, expiresAt time.Time, secure bool) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	c.SetCookie(&http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expiresAt,
		Secure:   secure,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearCSRFCookie removes the browser-readable CSRF token.
func ClearCSRFCookie(c *echo.Context, secure bool) {
	clearCookie(c, CSRFCookieName, false, secure)
}

func clearCookie(c *echo.Context, name string, httpOnly, secure bool) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
	})
}
