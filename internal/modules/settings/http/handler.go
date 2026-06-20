// Package settingshttp adapts setting and dictionary HTTP requests to the settings usecase.
package settingshttp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	"github.com/NSObjects/echo-admin/internal/modules/settings/usecase"
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

// OperationRecorder records setting mutations for audit.
type OperationRecorder interface {
	RecordOperation(context.Context, auditusecase.OperationInput) (auditusecase.OperationLog, error)
}

// Handler adapts setting and dictionary HTTP requests to the settings usecase.
type Handler struct {
	usecase   *usecase.Usecase
	auth      Authorizer
	operation OperationRecorder
}

// New creates a settings HTTP handler.
func New(uc *usecase.Usecase, auth Authorizer, operation OperationRecorder) *Handler {
	return &Handler{usecase: uc, auth: auth, operation: operation}
}

// Register mounts setting and dictionary routes on group.
func Register(group *echo.Group, handler *Handler) {
	group.GET("/system/configs", handler.ListConfigs)
	group.PUT("/system/configs/:key", handler.UpsertConfig)
	group.DELETE("/system/configs/:key", handler.DeleteConfig)
	group.GET("/system/params", handler.ListParams)
	group.POST("/system/params", handler.CreateParam)
	group.POST("/system/params/batch-delete", handler.BatchDeleteParams)
	group.GET("/system/params/key/:key", handler.FindParamByKey)
	group.GET("/system/params/:id", handler.FindParam)
	group.PATCH("/system/params/:id", handler.UpdateParam)
	group.DELETE("/system/params/:id", handler.DeleteParam)
	group.GET("/dictionaries", handler.ListDictionaries)
	group.POST("/dictionaries", handler.CreateDictionary)
	group.GET("/dictionaries/export", handler.ExportDictionaries)
	group.POST("/dictionaries/import", handler.ImportDictionaries)
	group.PATCH("/dictionaries/:code", handler.UpdateDictionary)
	group.DELETE("/dictionaries/:code", handler.DeleteDictionary)
	group.POST("/dictionaries/:code/items", handler.AddDictionaryItem)
	group.PATCH("/dictionaries/:code/items/:item_id", handler.UpdateDictionaryItem)
	group.DELETE("/dictionaries/:code/items/:item_id", handler.DeleteDictionaryItem)
	group.GET("/system/versions", handler.ListVersions)
	group.POST("/system/versions", handler.CreateVersion)
	group.POST("/system/versions/export", handler.ExportVersion)
	group.POST("/system/versions/import", handler.ImportVersion)
	group.POST("/system/versions/batch-delete", handler.BatchDeleteVersions)
	group.GET("/system/versions/:id", handler.FindVersion)
	group.GET("/system/versions/:id/download", handler.DownloadVersion)
	group.PATCH("/system/versions/:id", handler.UpdateVersion)
	group.DELETE("/system/versions/:id", handler.DeleteVersion)
}

// ListConfigs returns system configs.
func (h *Handler) ListConfigs(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionConfigRead); err != nil {
		return err
	}
	configs, err := h.usecase.ListConfigs(c.Request().Context())
	if err != nil {
		return err
	}
	return httpresp.OK(c, configs)
}

// UpsertConfig creates or updates a system config.
func (h *Handler) UpsertConfig(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionConfigUpdate); err != nil {
		return err
	}
	var req configRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	key := c.Param("key")
	config, err := h.usecase.UpsertConfig(c.Request().Context(), usecase.ConfigInput{
		Key:    key,
		Name:   req.Name,
		Value:  req.Value,
		Public: req.Public,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "upsert", "config", config.Key, "updated config"); err != nil {
		return err
	}
	return httpresp.OK(c, config)
}

// DeleteConfig deletes a system config.
func (h *Handler) DeleteConfig(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionConfigDelete); err != nil {
		return err
	}
	key := c.Param("key")
	if err := h.usecase.DeleteConfig(c.Request().Context(), key); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "config", key, "deleted config"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{Code: key})
}

// ListParams returns system parameters.
func (h *Handler) ListParams(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionParamRead); err != nil {
		return err
	}
	input, err := paramListInput(c)
	if err != nil {
		return err
	}
	output, err := h.usecase.ListParams(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return paginated(c, output.Items, output.Page, output.PageSize, output.Total)
}

// FindParam returns one system parameter by id.
func (h *Handler) FindParam(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionParamRead); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "system param")
	if err != nil {
		return err
	}
	param, err := h.usecase.FindParam(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return httpresp.OK(c, param)
}

// FindParamByKey returns one system parameter by key.
func (h *Handler) FindParamByKey(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionParamRead); err != nil {
		return err
	}
	param, err := h.usecase.FindParamByKey(c.Request().Context(), c.Param("key"))
	if err != nil {
		return err
	}
	return httpresp.OK(c, param)
}

// CreateParam creates one system parameter.
func (h *Handler) CreateParam(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionParamCreate); err != nil {
		return err
	}
	var req paramRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	param, err := h.usecase.CreateParam(c.Request().Context(), paramInput(req))
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "create", "system_param", strconv.FormatInt(param.ID, 10), "created system param"); err != nil {
		return err
	}
	return httpresp.Created(c, param)
}

// UpdateParam updates one system parameter.
func (h *Handler) UpdateParam(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionParamUpdate); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "system param")
	if err != nil {
		return err
	}
	var req paramRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	param, err := h.usecase.UpdateParam(c.Request().Context(), updateParamInput(id, req))
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "system_param", strconv.FormatInt(param.ID, 10), "updated system param"); err != nil {
		return err
	}
	return httpresp.OK(c, param)
}

// DeleteParam deletes one system parameter.
func (h *Handler) DeleteParam(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionParamDelete); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "system param")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteParam(c.Request().Context(), id); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "system_param", strconv.FormatInt(id, 10), "deleted system param"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedIDResponse{ID: id})
}

// BatchDeleteParams deletes system parameters by id.
func (h *Handler) BatchDeleteParams(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionParamDelete); err != nil {
		return err
	}
	var req idsRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.usecase.DeleteParams(c.Request().Context(), req.IDs); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "system_param", "batch", "deleted system params"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedIDsResponse(req))
}

// ListDictionaries returns dictionaries.
func (h *Handler) ListDictionaries(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionDictRead); err != nil {
		return err
	}
	dictionaries, err := h.usecase.ListDictionaries(c.Request().Context())
	if err != nil {
		return err
	}
	return httpresp.OK(c, dictionaries)
}

// ExportDictionaries downloads all dictionaries as a JSON bundle.
func (h *Handler) ExportDictionaries(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionDictRead); err != nil {
		return err
	}
	data, err := h.usecase.DictionaryBundleJSON(c.Request().Context())
	if err != nil {
		return err
	}
	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="dictionaries.json"`)
	return c.Blob(http.StatusOK, "application/json; charset=utf-8", data)
}

// ImportDictionaries imports dictionaries from a JSON bundle.
func (h *Handler) ImportDictionaries(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionDictCreate); err != nil {
		return err
	}
	var bundle usecase.DictionaryBundle
	if err := httpreq.BindAndValidate(c, &bundle); err != nil {
		return err
	}
	dictionaries, err := h.usecase.ImportDictionaries(c.Request().Context(), bundle)
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "import", "dictionary", "bundle", "imported dictionaries"); err != nil {
		return err
	}
	return httpresp.OK(c, dictionaries)
}

// CreateDictionary creates a dictionary.
func (h *Handler) CreateDictionary(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionDictCreate); err != nil {
		return err
	}
	var req dictionaryRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	dictionary, err := h.usecase.CreateDictionary(c.Request().Context(), usecase.DictionaryInput{
		Code: req.Code,
		Name: req.Name,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "create", "dictionary", dictionary.Code, "created dictionary"); err != nil {
		return err
	}
	return httpresp.Created(c, dictionary)
}

// UpdateDictionary updates a dictionary.
func (h *Handler) UpdateDictionary(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionDictUpdate); err != nil {
		return err
	}
	var req updateDictionaryRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	dictionary, err := h.usecase.UpdateDictionary(c.Request().Context(), usecase.UpdateDictionaryInput{
		Code: c.Param("code"),
		Name: req.Name,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "dictionary", dictionary.Code, "updated dictionary"); err != nil {
		return err
	}
	return httpresp.OK(c, dictionary)
}

// DeleteDictionary deletes a dictionary.
func (h *Handler) DeleteDictionary(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionDictDelete); err != nil {
		return err
	}
	code := c.Param("code")
	if err := h.usecase.DeleteDictionary(c.Request().Context(), code); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "dictionary", code, "deleted dictionary"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{Code: code})
}

// AddDictionaryItem appends one dictionary item.
func (h *Handler) AddDictionaryItem(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionDictCreate); err != nil {
		return err
	}
	var req dictionaryItemRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	dictionary, err := h.usecase.AddDictionaryItem(c.Request().Context(), c.Param("code"), usecase.DictionaryItemInput{
		ParentID: req.ParentID,
		Label:    req.Label,
		Value:    req.Value,
		Extend:   req.Extend,
		Sort:     req.Sort,
		Active:   req.Active,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "create", "dictionary_item", c.Param("code"), "created dictionary item"); err != nil {
		return err
	}
	return httpresp.Created(c, dictionary)
}

// UpdateDictionaryItem updates one dictionary item.
func (h *Handler) UpdateDictionaryItem(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionDictUpdate); err != nil {
		return err
	}
	itemID, err := httpreq.PathID(c, "item_id", "dictionary item")
	if err != nil {
		return err
	}
	var req dictionaryItemRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	dictionary, err := h.usecase.UpdateDictionaryItem(c.Request().Context(), c.Param("code"), usecase.DictionaryItemInput{
		ID:       itemID,
		ParentID: req.ParentID,
		Label:    req.Label,
		Value:    req.Value,
		Extend:   req.Extend,
		Sort:     req.Sort,
		Active:   req.Active,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "dictionary_item", strconv.FormatInt(itemID, 10), "updated dictionary item"); err != nil {
		return err
	}
	return httpresp.OK(c, dictionary)
}

// DeleteDictionaryItem deletes one dictionary item.
func (h *Handler) DeleteDictionaryItem(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionDictDelete); err != nil {
		return err
	}
	itemID, err := httpreq.PathID(c, "item_id", "dictionary item")
	if err != nil {
		return err
	}
	dictionary, err := h.usecase.DeleteDictionaryItem(c.Request().Context(), c.Param("code"), itemID)
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "dictionary_item", strconv.FormatInt(itemID, 10), "deleted dictionary item"); err != nil {
		return err
	}
	return httpresp.OK(c, dictionary)
}

// ListVersions returns release records.
func (h *Handler) ListVersions(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionVersionRead); err != nil {
		return err
	}
	versions, err := h.usecase.ListVersions(c.Request().Context())
	if err != nil {
		return err
	}
	return httpresp.OK(c, versions)
}

// FindVersion returns one release record.
func (h *Handler) FindVersion(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionVersionRead); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "system version")
	if err != nil {
		return err
	}
	version, err := h.usecase.FindVersion(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return httpresp.OK(c, version)
}

// DownloadVersion returns one release record as a JSON attachment.
func (h *Handler) DownloadVersion(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionVersionRead); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "system version")
	if err != nil {
		return err
	}
	data, version, err := h.usecase.VersionJSON(c.Request().Context(), id)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("version_%s.json", version.Version)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.Blob(http.StatusOK, "application/json; charset=utf-8", data)
}

// CreateVersion creates a release record.
func (h *Handler) CreateVersion(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionVersionCreate); err != nil {
		return err
	}
	var req versionRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	version, err := h.usecase.CreateVersion(c.Request().Context(), usecase.VersionInput{
		Version:     req.Version,
		Name:        req.Name,
		Description: req.Description,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "create", "system_version", version.Version, "created system version"); err != nil {
		return err
	}
	return httpresp.Created(c, version)
}

// ExportVersion creates a portable version bundle from selected resources.
func (h *Handler) ExportVersion(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionVersionCreate); err != nil {
		return err
	}
	var req exportVersionRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	version, err := h.usecase.ExportVersion(c.Request().Context(), usecase.ExportVersionInput{
		Version:       req.Version,
		Name:          req.Name,
		Description:   req.Description,
		MenuIDs:       req.MenuIDs,
		APIIDs:        req.APIIDs,
		DictionaryIDs: req.DictionaryIDs,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "export", "system_version", version.Version, "exported system version"); err != nil {
		return err
	}
	return httpresp.Created(c, version)
}

// ImportVersion imports a portable version bundle.
func (h *Handler) ImportVersion(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionVersionCreate); err != nil {
		return err
	}
	var bundle usecase.VersionBundle
	if err := httpreq.BindAndValidate(c, &bundle); err != nil {
		return err
	}
	version, err := h.usecase.ImportVersion(c.Request().Context(), bundle)
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "import", "system_version", version.Version, "imported system version"); err != nil {
		return err
	}
	return httpresp.Created(c, version)
}

// UpdateVersion updates a release record.
func (h *Handler) UpdateVersion(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionVersionUpdate); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "system version")
	if err != nil {
		return err
	}
	var req versionRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	version, err := h.usecase.UpdateVersion(c.Request().Context(), usecase.UpdateVersionInput{
		ID:          id,
		Version:     req.Version,
		Name:        req.Name,
		Description: req.Description,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "system_version", strconv.FormatInt(version.ID, 10), "updated system version"); err != nil {
		return err
	}
	return httpresp.OK(c, version)
}

// DeleteVersion deletes a release record.
func (h *Handler) DeleteVersion(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionVersionDelete); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "system version")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteVersion(c.Request().Context(), id); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "system_version", strconv.FormatInt(id, 10), "deleted system version"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedIDResponse{ID: id})
}

// BatchDeleteVersions deletes release records by id.
func (h *Handler) BatchDeleteVersions(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionVersionDelete); err != nil {
		return err
	}
	var req idsRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.usecase.DeleteVersions(c.Request().Context(), req.IDs); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "system_version", "batch", "deleted system versions"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedIDsResponse(req))
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
		return apperr.New(apperr.ErrInternalServer, "settings handler is not configured")
	}
	return nil
}

func paramListInput(c *echo.Context) (usecase.ParamListInput, error) {
	page, pageSize, err := httpreq.Pagination(c, defaultPageSize)
	if err != nil {
		return usecase.ParamListInput{}, err
	}
	return usecase.ParamListInput{
		Page:     page,
		PageSize: pageSize,
		Name:     c.QueryParam("name"),
		Key:      c.QueryParam("key"),
	}, nil
}

func paramInput(req paramRequest) usecase.ParamInput {
	return usecase.ParamInput{
		Name:  req.Name,
		Key:   req.Key,
		Value: req.Value,
		Desc:  req.Desc,
	}
}

func updateParamInput(id int64, req paramRequest) usecase.UpdateParamInput {
	return usecase.UpdateParamInput{
		ID:    id,
		Name:  req.Name,
		Key:   req.Key,
		Value: req.Value,
		Desc:  req.Desc,
	}
}

func paginated(c *echo.Context, data interface{}, page, pageSize, total int) error {
	meta, err := httpresp.NewPageMeta(page, pageSize, total)
	if err != nil {
		return err
	}
	return httpresp.List(c, data, meta)
}

type deletedResponse struct {
	Code string `json:"code"`
}

type deletedIDResponse struct {
	ID int64 `json:"id"`
}

type idsRequest struct {
	IDs []int64 `json:"ids" validate:"required,min=1,dive,gt=0"`
}

type deletedIDsResponse struct {
	IDs []int64 `json:"ids"`
}
