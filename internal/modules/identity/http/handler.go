// Package identityhttp adapts administrator HTTP requests to the identity usecase.
package identityhttp

import (
	"context"
	"strconv"

	"github.com/labstack/echo/v5"

	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	"github.com/NSObjects/echo-admin/internal/modules/identity/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpreq"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpresp"
)

const defaultPageSize = 20

// OperationRecorder records administrator mutations for audit.
type OperationRecorder interface {
	RecordOperation(context.Context, auditusecase.OperationInput) (auditusecase.OperationLog, error)
}

// Handler adapts administrator HTTP requests to the identity usecase.
type Handler struct {
	usecase   *usecase.Usecase
	operation OperationRecorder
}

// New creates an identity HTTP handler.
func New(uc *usecase.Usecase, operation OperationRecorder) *Handler {
	return &Handler{usecase: uc, operation: operation}
}

// Register mounts administrator routes on group.
func Register(group *echo.Group, handler *Handler) {
	group.GET("/admins", handler.ListAdmins)
	group.POST("/admins", handler.CreateAdmin)
	group.PATCH("/admins/:id", handler.UpdateAdmin)
	group.DELETE("/admins/:id", handler.DeleteAdmin)
	group.GET("/roles/:id/admins", handler.ListRoleAdmins)
	group.PUT("/roles/:id/admins", handler.SetRoleAdmins)
}

// ListAdmins returns administrators.
func (h *Handler) ListAdmins(c *echo.Context) error {
	if err := h.ready(); err != nil {
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
	if err := h.ready(); err != nil {
		return err
	}
	var req createAdminRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	admin, err := h.usecase.Create(c.Request().Context(), usecase.AdminInput{
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		Email:        req.Email,
		Password:     req.Password,
		RoleIDs:      req.RoleIDs,
		ActiveRoleID: req.ActiveRoleID,
		Active:       req.Active,
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
	if err := h.ready(); err != nil {
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
		ID:           id,
		DisplayName:  req.DisplayName,
		Email:        req.Email,
		Password:     req.Password,
		RoleIDs:      req.RoleIDs,
		ActiveRoleID: req.ActiveRoleID,
		Active:       req.Active,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "admin", strconv.FormatInt(admin.ID, 10), "updated admin"); err != nil {
		return err
	}
	return httpresp.OK(c, admin)
}

// DeleteAdmin deletes an administrator.
func (h *Handler) DeleteAdmin(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "admin")
	if err != nil {
		return err
	}
	if err := h.usecase.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "admin", strconv.FormatInt(id, 10), "deleted admin"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{ID: id})
}

// ListRoleAdmins returns administrator ids assigned to a role.
func (h *Handler) ListRoleAdmins(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "role")
	if err != nil {
		return err
	}
	adminIDs, err := h.usecase.RoleAdminIDs(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return httpresp.OK(c, adminIDsResponse{AdminIDs: adminIDs})
}

// SetRoleAdmins replaces administrator assignments for a role.
func (h *Handler) SetRoleAdmins(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "role")
	if err != nil {
		return err
	}
	var req adminIDsRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	adminIDs, err := h.usecase.SetRoleAdmins(c.Request().Context(), usecase.RoleAdminsInput{
		RoleID:   id,
		AdminIDs: req.AdminIDs,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "set_admins", "role", strconv.FormatInt(id, 10), "updated role admins"); err != nil {
		return err
	}
	return httpresp.OK(c, adminIDsResponse{AdminIDs: adminIDs})
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
	if h == nil || h.usecase == nil || h.operation == nil {
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

type deletedResponse struct {
	ID int64 `json:"id"`
}

type adminIDsResponse struct {
	AdminIDs []int64 `json:"admin_ids"`
}
