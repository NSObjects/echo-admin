package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
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
// login session.
type LoginSessionIdentity struct {
	SessionID string
	UserID    string
	RoleID    string
}

// LoginSessionAuthenticator validates browser login session credentials.
type LoginSessionAuthenticator interface {
	AuthenticateLoginSession(context.Context, string) (LoginSessionIdentity, error)
}

// LoginSessionConfig controls browser login-session authentication.
type LoginSessionConfig struct {
	CookieName    string
	SkipPaths     []string
	Authenticator LoginSessionAuthenticator
	Enabled       bool
}

// DefaultLoginSessionConfig returns disabled login-session authentication
// defaults.
func DefaultLoginSessionConfig() *LoginSessionConfig {
	return &LoginSessionConfig{
		CookieName: LoginSessionCookieName,
		SkipPaths: []string{
			"/api/health",
			"/api/info",
			"/api/ready",
			"/api/capabilities",
			"/api/auth/login",
			"/api/setup/state",
			"/api/setup",
		},
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
			if requestctx.GetUserID(c.Request().Context()) != "" || pathSkipped(c, config.SkipPaths) {
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
			if identity.UserID == "" || identity.RoleID == "" || identity.SessionID == "" {
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
				Str("user_id", identity.UserID).
				Str("role_id", identity.RoleID).
				Str("login_session_id", identity.SessionID).
				Str("auth", "login_session").
				Logger()
			c.SetRequest(request.WithContext(logger.WithContext(ctx)))
			return next(c)
		}
	}, nil
}

// CSRFConfig returns the Echo CSRF middleware configuration used for browser
// login-session requests.
func CSRFConfig(skipPaths []string, secureCookies bool) middleware.CSRFConfig {
	return middleware.CSRFConfig{
		Skipper: func(c *echo.Context) bool {
			return requestctx.GetLoginSessionID(c.Request().Context()) == "" || pathSkipped(c, skipPaths)
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

func pathSkipped(c *echo.Context, skipPaths []string) bool {
	path := requestPath(c)
	for _, skipPath := range skipPaths {
		if path == skipPath ||
			(len(skipPath) > 0 && skipPath[len(skipPath)-1] == '*' &&
				len(path) >= len(skipPath)-1 &&
				strings.HasPrefix(path, skipPath[:len(skipPath)-1])) {
			return true
		}
	}
	return false
}

// FormatLoginSessionID converts positive numeric IDs to request metadata.
func FormatLoginSessionID(id int64) string {
	if id <= 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}
