// Package accesshttp adapts role and menu HTTP requests to the access usecase.
package accesshttp

import (
	"context"
	"strconv"

	"github.com/labstack/echo/v5"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	"github.com/NSObjects/echo-admin/internal/modules/access/usecase"
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
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

// OperationRecorder records access mutations for audit.
type OperationRecorder interface {
	RecordOperation(context.Context, auditusecase.OperationInput) (auditusecase.OperationLog, error)
}

// Handler adapts role and menu HTTP requests to the access usecase.
type Handler struct {
	usecase   *usecase.Usecase
	auth      Authorizer
	operation OperationRecorder
}

// New creates an access HTTP handler.
func New(uc *usecase.Usecase, auth Authorizer, operation OperationRecorder) *Handler {
	return &Handler{usecase: uc, auth: auth, operation: operation}
}

// Register mounts role and menu routes on group.
func Register(group *echo.Group, handler *Handler) {
	group.GET("/roles", handler.ListRoles)
	group.POST("/roles", handler.CreateRole)
	group.PATCH("/roles/:id", handler.UpdateRole)
	group.GET("/menus", handler.ListMenus)
	group.POST("/menus", handler.CreateMenu)
	group.PATCH("/menus/:id", handler.UpdateMenu)
}

// ListRoles returns roles.
func (h *Handler) ListRoles(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionRoleManage); err != nil {
		return err
	}
	input, err := listInput(c)
	if err != nil {
		return err
	}
	output, err := h.usecase.ListRoles(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return paginated(c, output.Items, output.Page, output.PageSize, output.Total)
}

// CreateRole creates a role.
func (h *Handler) CreateRole(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionRoleManage); err != nil {
		return err
	}
	var req createRoleRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	role, err := h.usecase.CreateRole(c.Request().Context(), usecase.RoleInput{
		Code:        req.Code,
		Name:        req.Name,
		Permissions: req.Permissions,
		MenuIDs:     req.MenuIDs,
		Active:      req.Active,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "create", "role", strconv.FormatInt(role.ID, 10), "created role"); err != nil {
		return err
	}
	return httpresp.Created(c, role)
}

// UpdateRole updates a role.
func (h *Handler) UpdateRole(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionRoleManage); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "role")
	if err != nil {
		return err
	}
	var req updateRoleRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	role, err := h.usecase.UpdateRole(c.Request().Context(), usecase.UpdateRoleInput{
		ID:          id,
		Name:        req.Name,
		Permissions: req.Permissions,
		MenuIDs:     req.MenuIDs,
		Active:      req.Active,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "role", strconv.FormatInt(role.ID, 10), "updated role"); err != nil {
		return err
	}
	return httpresp.OK(c, role)
}

// ListMenus returns menus.
func (h *Handler) ListMenus(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionMenuManage); err != nil {
		return err
	}
	menus, err := h.usecase.ListMenus(c.Request().Context())
	if err != nil {
		return err
	}
	return httpresp.OK(c, menus)
}

// CreateMenu creates a menu.
func (h *Handler) CreateMenu(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionMenuManage); err != nil {
		return err
	}
	var req menuRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	menu, err := h.usecase.CreateMenu(c.Request().Context(), usecase.MenuInput{
		ParentID:   req.ParentID,
		Name:       req.Name,
		Path:       req.Path,
		Icon:       req.Icon,
		Permission: req.Permission,
		Sort:       req.Sort,
		Active:     req.Active,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "create", "menu", strconv.FormatInt(menu.ID, 10), "created menu"); err != nil {
		return err
	}
	return httpresp.Created(c, menu)
}

// UpdateMenu updates a menu.
func (h *Handler) UpdateMenu(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionMenuManage); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "menu")
	if err != nil {
		return err
	}
	var req menuRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	menu, err := h.usecase.UpdateMenu(c.Request().Context(), usecase.UpdateMenuInput{
		ID:         id,
		ParentID:   req.ParentID,
		Name:       req.Name,
		Path:       req.Path,
		Icon:       req.Icon,
		Permission: req.Permission,
		Sort:       req.Sort,
		Active:     req.Active,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "menu", strconv.FormatInt(menu.ID, 10), "updated menu"); err != nil {
		return err
	}
	return httpresp.OK(c, menu)
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
		return apperr.New(apperr.ErrInternalServer, "access handler is not configured")
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
