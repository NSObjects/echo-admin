// Package audithttp adapts audit HTTP requests to the audit usecase.
package audithttp

import (
	"context"
	"strconv"

	"github.com/labstack/echo/v5"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	"github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpreq"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpresp"
)

const defaultPageSize = 20

// Authorizer checks whether the current request can perform an action.
type Authorizer interface {
	RequireRoutePermission(context.Context, string, string, string) error
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
	group.GET("/logs/operations/:id", handler.ReadOperationLog)
	group.DELETE("/logs/operations/:id", handler.DeleteOperationLog)
	group.POST("/logs/operations/batch-delete", handler.BatchDeleteOperationLogs)
	group.GET("/logs/logins", handler.ListLoginLogs)
	group.GET("/logs/logins/:id", handler.ReadLoginLog)
	group.DELETE("/logs/logins/:id", handler.DeleteLoginLog)
	group.POST("/logs/logins/batch-delete", handler.BatchDeleteLoginLogs)
	group.GET("/logs/errors", handler.ListSystemErrorLogs)
	group.GET("/logs/errors/:id", handler.ReadSystemErrorLog)
	group.POST("/logs/errors/:id/resolve", handler.ResolveSystemErrorLog)
	group.DELETE("/logs/errors/:id/resolve", handler.ReopenSystemErrorLog)
	group.DELETE("/logs/errors/:id", handler.DeleteSystemErrorLog)
	group.POST("/logs/errors/batch-delete", handler.BatchDeleteSystemErrorLogs)
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

// ReadOperationLog returns one operation log.
func (h *Handler) ReadOperationLog(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionLogRead); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "operation log")
	if err != nil {
		return err
	}
	log, err := h.usecase.FindOperationLog(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return httpresp.OK(c, log)
}

// DeleteOperationLog removes one operation log.
func (h *Handler) DeleteOperationLog(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionLogDelete); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "operation log")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteOperationLogs(c.Request().Context(), []int64{id}); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{IDs: []int64{id}})
}

// BatchDeleteOperationLogs removes multiple operation logs.
func (h *Handler) BatchDeleteOperationLogs(c *echo.Context) error {
	ids, err := h.deleteIDs(c)
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteOperationLogs(c.Request().Context(), ids); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{IDs: ids})
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

// ReadLoginLog returns one login log.
func (h *Handler) ReadLoginLog(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionLogRead); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "login log")
	if err != nil {
		return err
	}
	log, err := h.usecase.FindLoginLog(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return httpresp.OK(c, log)
}

// DeleteLoginLog removes one login log.
func (h *Handler) DeleteLoginLog(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionLogDelete); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "login log")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteLoginLogs(c.Request().Context(), []int64{id}); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{IDs: []int64{id}})
}

// BatchDeleteLoginLogs removes multiple login logs.
func (h *Handler) BatchDeleteLoginLogs(c *echo.Context) error {
	ids, err := h.deleteIDs(c)
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteLoginLogs(c.Request().Context(), ids); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{IDs: ids})
}

// ListSystemErrorLogs returns internal API failure logs.
func (h *Handler) ListSystemErrorLogs(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionLogRead); err != nil {
		return err
	}
	input, err := listInput(c)
	if err != nil {
		return err
	}
	output, err := h.usecase.ListSystemErrorLogs(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return paginated(c, output.Items, output.Page, output.PageSize, output.Total)
}

// ReadSystemErrorLog returns one system error log.
func (h *Handler) ReadSystemErrorLog(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionLogRead); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "system error log")
	if err != nil {
		return err
	}
	log, err := h.usecase.FindSystemErrorLog(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return httpresp.OK(c, log)
}

// ResolveSystemErrorLog marks one system error as handled.
func (h *Handler) ResolveSystemErrorLog(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionLogResolve); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "system error log")
	if err != nil {
		return err
	}
	var req resolveErrorRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	actorID, err := currentActorID(c.Request().Context())
	if err != nil {
		return err
	}
	log, err := h.usecase.ResolveSystemErrorLog(c.Request().Context(), usecase.ResolveSystemErrorInput{
		ID:         id,
		ResolverID: actorID,
		Note:       req.Note,
	})
	if err != nil {
		return err
	}
	return httpresp.OK(c, log)
}

// ReopenSystemErrorLog clears the handled state for one system error.
func (h *Handler) ReopenSystemErrorLog(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionLogResolve); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "system error log")
	if err != nil {
		return err
	}
	log, err := h.usecase.ReopenSystemErrorLog(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return httpresp.OK(c, log)
}

// DeleteSystemErrorLog removes one system error log.
func (h *Handler) DeleteSystemErrorLog(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionLogDelete); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "system error log")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteSystemErrorLogs(c.Request().Context(), []int64{id}); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{IDs: []int64{id}})
}

// BatchDeleteSystemErrorLogs removes multiple system error logs.
func (h *Handler) BatchDeleteSystemErrorLogs(c *echo.Context) error {
	ids, err := h.deleteIDs(c)
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteSystemErrorLogs(c.Request().Context(), ids); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{IDs: ids})
}

func (h *Handler) authorize(c *echo.Context, permission string) error {
	if err := h.ready(); err != nil {
		return err
	}
	return h.auth.RequireRoutePermission(c.Request().Context(), permission, c.Request().Method, c.Path())
}

func (h *Handler) deleteIDs(c *echo.Context) ([]int64, error) {
	if err := h.authorize(c, accessdomain.PermissionLogDelete); err != nil {
		return nil, err
	}
	var req idsRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return nil, err
	}
	return req.IDs, nil
}

func (h *Handler) ready() error {
	if h == nil || h.usecase == nil || h.auth == nil {
		return apperr.New(apperr.ErrInternalServer, "audit handler is not configured")
	}
	return nil
}

func currentActorID(ctx context.Context) (int64, error) {
	id, err := strconv.ParseInt(requestctx.GetUserID(ctx), 10, 64)
	if err != nil || id <= 0 {
		return 0, apperr.NewUnauthorized()
	}
	return id, nil
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

type idsRequest struct {
	IDs []int64 `json:"ids" validate:"required,min=1,dive,gt=0"`
}

type resolveErrorRequest struct {
	Note string `json:"note" validate:"omitempty,max=1000"`
}

type deletedResponse struct {
	IDs []int64 `json:"ids"`
}
