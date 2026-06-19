// Package audithttp adapts audit HTTP requests to the audit usecase.
package audithttp

import (
	"context"

	"github.com/labstack/echo/v5"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	"github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpreq"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpresp"
)

const defaultPageSize = 20

// Authorizer checks whether the current request can perform an action.
type Authorizer interface {
	RequirePermission(context.Context, string) error
}

// Handler adapts audit HTTP requests to the audit usecase.
type Handler struct {
	usecase *usecase.Usecase
	auth    Authorizer
}

// New creates an audit HTTP handler.
func New(uc *usecase.Usecase, auth Authorizer) *Handler {
	return &Handler{usecase: uc, auth: auth}
}

// Register mounts audit routes on group.
func Register(group *echo.Group, handler *Handler) {
	group.GET("/logs/operations", handler.ListOperationLogs)
	group.GET("/logs/logins", handler.ListLoginLogs)
}

// ListOperationLogs returns operation logs.
func (h *Handler) ListOperationLogs(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionLogRead); err != nil {
		return err
	}
	input, err := listInput(c)
	if err != nil {
		return err
	}
	output, err := h.usecase.ListOperationLogs(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return paginated(c, output.Items, output.Page, output.PageSize, output.Total)
}

// ListLoginLogs returns login logs.
func (h *Handler) ListLoginLogs(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionLogRead); err != nil {
		return err
	}
	input, err := listInput(c)
	if err != nil {
		return err
	}
	output, err := h.usecase.ListLoginLogs(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return paginated(c, output.Items, output.Page, output.PageSize, output.Total)
}

func (h *Handler) authorize(c *echo.Context, permission string) error {
	if err := h.ready(); err != nil {
		return err
	}
	return h.auth.RequirePermission(c.Request().Context(), permission)
}

func (h *Handler) ready() error {
	if h == nil || h.usecase == nil || h.auth == nil {
		return apperr.New(apperr.ErrInternalServer, "audit handler is not configured")
	}
	return nil
}

func listInput(c *echo.Context) (usecase.ListInput, error) {
	page, pageSize, err := httpreq.Pagination(c, defaultPageSize)
	if err != nil {
		return usecase.ListInput{}, err
	}
	return usecase.ListInput{Page: page, PageSize: pageSize}, nil
}

func paginated(c *echo.Context, items interface{}, page, pageSize, total int) error {
	meta, err := httpresp.NewPageMeta(page, pageSize, total)
	if err != nil {
		return err
	}
	return httpresp.List(c, items, meta)
}
