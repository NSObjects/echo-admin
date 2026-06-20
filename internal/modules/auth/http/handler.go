// Package authhttp adapts authentication HTTP requests to the auth usecase.
package authhttp

import (
	"strings"

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
	group.POST("/auth/password", handler.ChangePassword)
	group.POST("/auth/role", handler.SwitchRole)
	group.GET("/auth/me", handler.Me)
	group.PATCH("/auth/me", handler.UpdateProfile)
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

// Logout revokes the current bearer token on the server.
func (h *Handler) Logout(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	token, err := bearerToken(c)
	if err != nil {
		return err
	}
	if err := h.usecase.Logout(c.Request().Context(), token); err != nil {
		return err
	}
	return httpresp.OK(c, map[string]bool{"ok": true})
}

// ChangePassword updates the current administrator password and revokes the
// bearer token used by the request.
func (h *Handler) ChangePassword(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	var req changePasswordRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	token, err := bearerToken(c)
	if err != nil {
		return err
	}
	if err := h.usecase.ChangePassword(c.Request().Context(), usecase.ChangePasswordInput{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
		RawToken:        token,
	}); err != nil {
		return err
	}
	return httpresp.OK(c, map[string]bool{"ok": true})
}

// SwitchRole changes the active role for the current administrator.
func (h *Handler) SwitchRole(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	var req switchRoleRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	output, err := h.usecase.SwitchRole(c.Request().Context(), usecase.RoleSwitchInput{RoleID: req.RoleID})
	if err != nil {
		return err
	}
	return httpresp.OK(c, output)
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

// UpdateProfile updates the current administrator's profile fields.
func (h *Handler) UpdateProfile(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	var req updateProfileRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	user, err := h.usecase.UpdateProfile(c.Request().Context(), usecase.UpdateProfileInput{
		DisplayName: req.DisplayName,
		Email:       req.Email,
	})
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

func bearerToken(c *echo.Context) (string, error) {
	authHeader := strings.TrimSpace(c.Request().Header.Get(echo.HeaderAuthorization))
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", apperr.NewUnauthorized()
	}
	return parts[1], nil
}
