// Package authhttp adapts authentication HTTP requests to the auth usecase.
package authhttp

import (
	"github.com/labstack/echo/v5"

	"github.com/NSObjects/echo-admin/internal/modules/auth/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpreq"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpresp"
)

// Handler adapts auth HTTP requests to the auth usecase.
type Handler struct {
	usecase *usecase.Usecase
}

// New creates an auth HTTP handler.
func New(uc *usecase.Usecase) *Handler {
	return &Handler{usecase: uc}
}

// Register mounts authentication routes on group.
func Register(group *echo.Group, handler *Handler) {
	group.POST("/auth/login", handler.Login)
	group.POST("/auth/logout", handler.Logout)
	group.GET("/auth/me", handler.Me)
}

// Login handles administrator login.
func (h *Handler) Login(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	var req loginRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	output, err := h.usecase.Login(c.Request().Context(), usecase.LoginInput{
		Username:  req.Username,
		Password:  req.Password,
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})
	if err != nil {
		return err
	}
	return httpresp.OK(c, output)
}

// Logout handles client-side token disposal.
func (h *Handler) Logout(c *echo.Context) error {
	return httpresp.OK(c, map[string]bool{"ok": true})
}

// Me returns the authenticated administrator profile.
func (h *Handler) Me(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	user, err := h.usecase.CurrentUser(c.Request().Context())
	if err != nil {
		return err
	}
	return httpresp.OK(c, user)
}

func (h *Handler) ready() error {
	if h == nil || h.usecase == nil {
		return apperr.New(apperr.ErrInternalServer, "auth usecase is not configured")
	}
	return nil
}
