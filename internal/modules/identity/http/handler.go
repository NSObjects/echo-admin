// Package identityhttp adapts administrator HTTP requests to the identity usecase.
package identityhttp

import (
	"context"
	"strconv"

	"github.com/labstack/echo/v5"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	"github.com/NSObjects/echo-admin/internal/modules/identity/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpreq"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpresp"
)

const defaultPageSize = 20

// Authorizer checks whether the current request can perform an action.
type Authorizer interface {
	RequirePermission(context.Context, string) error
}

// OperationRecorder records administrator mutations for audit.
type OperationRecorder interface {
	RecordOperation(context.Context, auditusecase.OperationInput) (auditusecase.OperationLog, error)
}

// Handler adapts administrator HTTP requests to the identity usecase.
type Handler struct {
	usecase   *usecase.Usecase
	auth      Authorizer
	operation OperationRecorder
}

// New creates an identity HTTP handler.
func New(uc *usecase.Usecase, auth Authorizer, operation OperationRecorder) *Handler {
	return &Handler{usecase: uc, auth: auth, operation: operation}
}

// Register mounts administrator routes on group.
func Register(group *echo.Group, handler *Handler) {
	group.GET("/admins", handler.ListAdmins)
	group.POST("/admins", handler.CreateAdmin)
	group.PATCH("/admins/:id", handler.UpdateAdmin)
}

// ListAdmins returns administrators.
func (h *Handler) ListAdmins(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionAdminManage); err != nil {
		return err
	}
	input, err := listInput(c)
	if err != nil {
		return err
	}
	output, err := h.usecase.List(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return paginated(c, output.Items, output.Page, output.PageSize, output.Total)
}

// CreateAdmin creates an administrator.
func (h *Handler) CreateAdmin(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionAdminManage); err != nil {
		return err
	}
	var req createAdminRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	admin, err := h.usecase.Create(c.Request().Context(), usecase.AdminInput{
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Password:    req.Password,
		RoleIDs:     req.RoleIDs,
		Active:      req.Active,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "create", "admin", strconv.FormatInt(admin.ID, 10), "created admin"); err != nil {
		return err
	}
	return httpresp.Created(c, admin)
}

// UpdateAdmin updates an administrator.
func (h *Handler) UpdateAdmin(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionAdminManage); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "admin")
	if err != nil {
		return err
	}
	var req updateAdminRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	admin, err := h.usecase.Update(c.Request().Context(), usecase.UpdateAdminInput{
		ID:          id,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Password:    req.Password,
		RoleIDs:     req.RoleIDs,
		Active:      req.Active,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "admin", strconv.FormatInt(admin.ID, 10), "updated admin"); err != nil {
		return err
	}
	return httpresp.OK(c, admin)
}

func (h *Handler) authorize(c *echo.Context, permission string) error {
	if err := h.ready(); err != nil {
		return err
	}
	return h.auth.RequirePermission(c.Request().Context(), permission)
}

func (h *Handler) recordOperation(c *echo.Context, action, resource, resourceID, message string) error {
	actorID, err := strconv.ParseInt(requestctx.GetUserID(c.Request().Context()), 10, 64)
	if err != nil {
		return apperr.NewUnauthorized()
	}
	_, err = h.operation.RecordOperation(c.Request().Context(), auditusecase.OperationInput{
		ActorID:    actorID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Method:     c.Request().Method,
		Path:       c.Path(),
		IP:         c.RealIP(),
		UserAgent:  c.Request().UserAgent(),
		Success:    true,
		Message:    message,
	})
	return err
}

func (h *Handler) ready() error {
	if h == nil || h.usecase == nil || h.auth == nil || h.operation == nil {
		return apperr.New(apperr.ErrInternalServer, "identity handler is not configured")
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
