// Package settingshttp adapts setting and dictionary HTTP requests to the settings usecase.
package settingshttp

import (
	"context"
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

// Authorizer checks whether the current request can perform an action.
type Authorizer interface {
	RequirePermission(context.Context, string) error
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
	group.GET("/dictionaries", handler.ListDictionaries)
	group.POST("/dictionaries", handler.CreateDictionary)
	group.POST("/dictionaries/:code/items", handler.AddDictionaryItem)
	group.PATCH("/dictionaries/:code/items/:item_id", handler.UpdateDictionaryItem)
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
		Label:  req.Label,
		Value:  req.Value,
		Sort:   req.Sort,
		Active: req.Active,
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
		ID:     itemID,
		Label:  req.Label,
		Value:  req.Value,
		Sort:   req.Sort,
		Active: req.Active,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "dictionary_item", strconv.FormatInt(itemID, 10), "updated dictionary item"); err != nil {
		return err
	}
	return httpresp.OK(c, dictionary)
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
		return apperr.New(apperr.ErrInternalServer, "settings handler is not configured")
	}
	return nil
}
