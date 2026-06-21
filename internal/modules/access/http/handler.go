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
	RequireRoutePermission(context.Context, string, string, string) error
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
	group.GET("/permissions", handler.ListPermissions)
	group.GET("/roles", handler.ListRoles)
	group.POST("/roles", handler.CreateRole)
	group.PATCH("/roles/:id", handler.UpdateRole)
	group.DELETE("/roles/:id", handler.DeleteRole)
	group.POST("/roles/:id/copy", handler.CopyRole)
	group.GET("/apis", handler.ListAPIs)
	group.GET("/apis/groups", handler.ListAPIGroups)
	group.POST("/apis", handler.CreateAPI)
	group.POST("/apis/batch-delete", handler.BatchDeleteAPIs)
	group.GET("/apis/:id", handler.ReadAPI)
	group.PATCH("/apis/:id", handler.UpdateAPI)
	group.DELETE("/apis/:id", handler.DeleteAPI)
	group.GET("/apis/:id/roles", handler.ReadAPIRoles)
	group.PUT("/apis/:id/roles", handler.SetAPIRoles)
	group.GET("/menus", handler.ListMenus)
	group.POST("/menus", handler.CreateMenu)
	group.GET("/menus/:id", handler.ReadMenu)
	group.PATCH("/menus/:id", handler.UpdateMenu)
	group.DELETE("/menus/:id", handler.DeleteMenu)
	group.GET("/menus/:id/roles", handler.ReadMenuRoles)
	group.PUT("/menus/:id/roles", handler.SetMenuRoles)
}

// ListPermissions returns grant metadata.
func (h *Handler) ListPermissions(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionRoleRead); err != nil {
		return err
	}
	// The catalog feeds role editors. Route authorization keeps disabled
	// identities, unassigned API tokens, and non-role-management users from
	// enumerating the grant surface.
	permissions, err := h.usecase.ListPermissions(c.Request().Context())
	if err != nil {
		return err
	}
	return httpresp.OK(c, permissions)
}

// ListRoles returns roles.
func (h *Handler) ListRoles(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionRoleRead); err != nil {
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
	if err := h.authorize(c, accessdomain.PermissionRoleCreate); err != nil {
		return err
	}
	var req createRoleRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	role, err := h.usecase.CreateRole(c.Request().Context(), roleInputFromRequest(req))
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
	if err := h.authorize(c, accessdomain.PermissionRoleUpdate); err != nil {
		return err
	}
	input, err := updateRoleInput(c)
	if err != nil {
		return err
	}
	role, err := h.usecase.UpdateRole(c.Request().Context(), input)
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "role", strconv.FormatInt(role.ID, 10), "updated role"); err != nil {
		return err
	}
	return httpresp.OK(c, role)
}

// DeleteRole deletes a role.
func (h *Handler) DeleteRole(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionRoleDelete); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "role")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteRole(c.Request().Context(), id); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "role", strconv.FormatInt(id, 10), "deleted role"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{ID: id})
}

// CopyRole copies a role with its grants.
func (h *Handler) CopyRole(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionRoleCreate); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "role")
	if err != nil {
		return err
	}
	var req copyRoleRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	role, err := h.usecase.CopyRole(c.Request().Context(), usecase.CopyRoleInput{
		SourceID:    id,
		ParentID:    req.ParentID,
		Code:        req.Code,
		Name:        req.Name,
		DefaultPath: req.DefaultPath,
		Active:      req.Active,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "copy", "role", strconv.FormatInt(role.ID, 10), "copied role"); err != nil {
		return err
	}
	return httpresp.Created(c, role)
}

// ListAPIs returns API routes.
func (h *Handler) ListAPIs(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionAPIRead); err != nil {
		return err
	}
	input, err := listInput(c)
	if err != nil {
		return err
	}
	output, err := h.usecase.ListAPIs(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return paginated(c, output.Items, output.Page, output.PageSize, output.Total)
}

// ListAPIGroups returns API group names.
func (h *Handler) ListAPIGroups(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionAPIRead); err != nil {
		return err
	}
	groups, err := h.usecase.APIGroups(c.Request().Context())
	if err != nil {
		return err
	}
	return httpresp.OK(c, apiGroupsResponse{Groups: groups})
}

// ReadAPI returns one API route.
func (h *Handler) ReadAPI(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionAPIRead); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "api")
	if err != nil {
		return err
	}
	api, err := h.usecase.FindAPI(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return httpresp.OK(c, api)
}

// CreateAPI creates an API route.
func (h *Handler) CreateAPI(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionAPICreate); err != nil {
		return err
	}
	var req apiRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	api, err := h.usecase.CreateAPI(c.Request().Context(), apiInputFromRequest(req))
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "create", "api", strconv.FormatInt(api.ID, 10), "created api"); err != nil {
		return err
	}
	return httpresp.Created(c, api)
}

// UpdateAPI updates an API route.
func (h *Handler) UpdateAPI(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionAPIUpdate); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "api")
	if err != nil {
		return err
	}
	var req apiRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	api, err := h.usecase.UpdateAPI(c.Request().Context(), usecase.UpdateAPIInput{
		ID:          id,
		Method:      req.Method,
		Path:        req.Path,
		Description: req.Description,
		Group:       req.Group,
		Permission:  req.Permission,
		Public:      req.Public,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "api", strconv.FormatInt(api.ID, 10), "updated api"); err != nil {
		return err
	}
	return httpresp.OK(c, api)
}

// BatchDeleteAPIs removes API routes by id.
func (h *Handler) BatchDeleteAPIs(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionAPIDelete); err != nil {
		return err
	}
	var req idsRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.usecase.DeleteAPIs(c.Request().Context(), req.IDs); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "api", "batch", "deleted apis"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedIDsResponse(req))
}

// DeleteAPI deletes an API route.
func (h *Handler) DeleteAPI(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionAPIDelete); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "api")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteAPI(c.Request().Context(), id); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "api", strconv.FormatInt(id, 10), "deleted api"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{ID: id})
}

// ReadAPIRoles returns role ids assigned to an API route.
func (h *Handler) ReadAPIRoles(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionAPIRead); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "api")
	if err != nil {
		return err
	}
	roleIDs, err := h.usecase.APIRoleIDs(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return httpresp.OK(c, roleIDsResponse{RoleIDs: roleIDs})
}

// SetAPIRoles replaces role assignments for an API route.
func (h *Handler) SetAPIRoles(c *echo.Context) error {
	return h.setRoleAssignments(c, accessdomain.PermissionAPIUpdate, "api", "updated api roles", func(ctx context.Context, id int64, roleIDs []int64) ([]int64, error) {
		return h.usecase.SetAPIRoles(ctx, usecase.APIRolesInput{APIID: id, RoleIDs: roleIDs})
	})
}

// ListMenus returns menus.
func (h *Handler) ListMenus(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionMenuRead); err != nil {
		return err
	}
	menus, err := h.usecase.ListMenus(c.Request().Context())
	if err != nil {
		return err
	}
	return httpresp.OK(c, menus)
}

// ReadMenu returns one menu.
func (h *Handler) ReadMenu(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionMenuRead); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "menu")
	if err != nil {
		return err
	}
	menu, err := h.usecase.FindMenu(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return httpresp.OK(c, menu)
}

// CreateMenu creates a menu.
func (h *Handler) CreateMenu(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionMenuCreate); err != nil {
		return err
	}
	var req menuRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	menu, err := h.usecase.CreateMenu(c.Request().Context(), menuInputFromRequest(req))
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
	if err := h.authorize(c, accessdomain.PermissionMenuUpdate); err != nil {
		return err
	}
	input, err := updateMenuInput(c)
	if err != nil {
		return err
	}
	menu, err := h.usecase.UpdateMenu(c.Request().Context(), input)
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "menu", strconv.FormatInt(menu.ID, 10), "updated menu"); err != nil {
		return err
	}
	return httpresp.OK(c, menu)
}

// DeleteMenu deletes a menu.
func (h *Handler) DeleteMenu(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionMenuDelete); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "menu")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteMenu(c.Request().Context(), id); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "menu", strconv.FormatInt(id, 10), "deleted menu"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{ID: id})
}

// ReadMenuRoles returns role ids assigned to a menu.
func (h *Handler) ReadMenuRoles(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionMenuRead); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "menu")
	if err != nil {
		return err
	}
	roleIDs, err := h.usecase.MenuRoleIDs(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return httpresp.OK(c, roleIDsResponse{RoleIDs: roleIDs})
}

// SetMenuRoles replaces role assignments for a menu.
func (h *Handler) SetMenuRoles(c *echo.Context) error {
	return h.setRoleAssignments(c, accessdomain.PermissionMenuUpdate, "menu", "updated menu roles", func(ctx context.Context, id int64, roleIDs []int64) ([]int64, error) {
		return h.usecase.SetMenuRoles(ctx, usecase.MenuRolesInput{MenuID: id, RoleIDs: roleIDs})
	})
}

func (h *Handler) setRoleAssignments(
	c *echo.Context,
	permission string,
	resource string,
	message string,
	assign func(context.Context, int64, []int64) ([]int64, error),
) error {
	if err := h.authorize(c, permission); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", resource)
	if err != nil {
		return err
	}
	var req roleIDsRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	roleIDs, err := assign(c.Request().Context(), id, req.RoleIDs)
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "set_roles", resource, strconv.FormatInt(id, 10), message); err != nil {
		return err
	}
	return httpresp.OK(c, roleIDsResponse{RoleIDs: roleIDs})
}

func (h *Handler) authorize(c *echo.Context, permission string) error {
	if err := h.ready(); err != nil {
		return err
	}
	return h.auth.RequireRoutePermission(c.Request().Context(), permission, c.Request().Method, c.Path())
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

func roleInputFromRequest(req createRoleRequest) usecase.RoleInput {
	return usecase.RoleInput{
		ParentID:    req.ParentID,
		Code:        req.Code,
		Name:        req.Name,
		Permissions: req.Permissions,
		MenuIDs:     req.MenuIDs,
		APIIDs:      req.APIIDs,
		ButtonIDs:   req.ButtonIDs,
		DataRoleIDs: req.DataRoleIDs,
		DefaultPath: req.DefaultPath,
		Active:      req.Active,
	}
}

func updateRoleInput(c *echo.Context) (usecase.UpdateRoleInput, error) {
	id, err := httpreq.PathID(c, "id", "role")
	if err != nil {
		return usecase.UpdateRoleInput{}, err
	}
	var req updateRoleRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return usecase.UpdateRoleInput{}, bindErr
	}
	return usecase.UpdateRoleInput{
		ID:          id,
		ParentID:    req.ParentID,
		Name:        req.Name,
		Permissions: req.Permissions,
		MenuIDs:     req.MenuIDs,
		APIIDs:      req.APIIDs,
		ButtonIDs:   req.ButtonIDs,
		DataRoleIDs: req.DataRoleIDs,
		DefaultPath: req.DefaultPath,
		Active:      req.Active,
	}, nil
}

func menuInputFromRequest(req menuRequest) usecase.MenuInput {
	return usecase.MenuInput{
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Icon:      req.Icon,
		Hidden:    req.Hidden,
		Component: req.Component,
		Meta: usecase.MenuMetaInput{
			ActiveName:     req.Meta.ActiveName,
			KeepAlive:      req.Meta.KeepAlive,
			DefaultMenu:    req.Meta.DefaultMenu,
			CloseTab:       req.Meta.CloseTab,
			TransitionType: req.Meta.TransitionType,
		},
		Permission: req.Permission,
		Sort:       req.Sort,
		Active:     req.Active,
		Buttons:    buttonInputsFromRequest(req.Buttons),
	}
}

func updateMenuInput(c *echo.Context) (usecase.UpdateMenuInput, error) {
	id, err := httpreq.PathID(c, "id", "menu")
	if err != nil {
		return usecase.UpdateMenuInput{}, err
	}
	var req menuRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return usecase.UpdateMenuInput{}, bindErr
	}
	return usecase.UpdateMenuInput{
		ID:        id,
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Icon:      req.Icon,
		Hidden:    req.Hidden,
		Component: req.Component,
		Meta: usecase.MenuMetaInput{
			ActiveName:     req.Meta.ActiveName,
			KeepAlive:      req.Meta.KeepAlive,
			DefaultMenu:    req.Meta.DefaultMenu,
			CloseTab:       req.Meta.CloseTab,
			TransitionType: req.Meta.TransitionType,
		},
		Permission: req.Permission,
		Sort:       req.Sort,
		Active:     req.Active,
		Buttons:    buttonInputsFromRequest(req.Buttons),
	}, nil
}

func buttonInputsFromRequest(buttons []menuButtonRequest) []usecase.MenuButtonInput {
	out := make([]usecase.MenuButtonInput, 0, len(buttons))
	for _, button := range buttons {
		out = append(out, usecase.MenuButtonInput{
			ID:          button.ID,
			Name:        button.Name,
			Description: button.Description,
		})
	}
	return out
}

func apiInputFromRequest(req apiRequest) usecase.APIInput {
	return usecase.APIInput{
		Method:      req.Method,
		Path:        req.Path,
		Description: req.Description,
		Group:       req.Group,
		Permission:  req.Permission,
		Public:      req.Public,
	}
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

type deletedIDsResponse struct {
	IDs []int64 `json:"ids"`
}

type apiGroupsResponse struct {
	Groups []string `json:"groups"`
}

type roleIDsResponse struct {
	RoleIDs []int64 `json:"role_ids"`
}
