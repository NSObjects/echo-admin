// Package apitokenhttp adapts API token HTTP requests to the token usecase.
package apitokenhttp

import (
	"context"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/NSObjects/echo-admin/internal/modules/apitoken/usecase"
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpreq"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpresp"
)

const defaultPageSize = 20

// OperationRecorder records API token mutations for audit.
type OperationRecorder interface {
	RecordOperation(context.Context, auditusecase.OperationInput) (auditusecase.OperationLog, error)
}

// Handler adapts API token HTTP requests to the token usecase.
type Handler struct {
	usecase   *usecase.Usecase
	operation OperationRecorder
}

// New creates an API token HTTP handler.
func New(uc *usecase.Usecase, operation OperationRecorder) *Handler {
	return &Handler{usecase: uc, operation: operation}
}

// Register mounts API token routes on group.
func Register(group *echo.Group, handler *Handler) {
	group.GET("/api-tokens", handler.ListTokens)
	group.POST("/api-tokens", handler.CreateToken)
	group.PATCH("/api-tokens/:id", handler.UpdateToken)
	group.DELETE("/api-tokens/:id", handler.DeleteToken)
}

// ListTokens returns token metadata without raw secrets or hashes.
func (h *Handler) ListTokens(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	input, err := listInput(c)
	if err != nil {
		return err
	}
	output, err := h.usecase.ListTokens(c.Request().Context(), input)
	if err != nil {
		return err
	}
	page, err := httpresp.NewPageMeta(output.Page, output.PageSize, output.Total)
	if err != nil {
		return err
	}
	return httpresp.List(c, output.Items, page)
}

// CreateToken creates an API token and returns the raw secret once.
func (h *Handler) CreateToken(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	var req tokenRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	created, err := h.usecase.CreateToken(c.Request().Context(), tokenInputFromRequest(req))
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "create", "api_token", strconv.FormatInt(created.Token.ID, 10), "created api token"); err != nil {
		return err
	}
	return httpresp.Created(c, created)
}

// UpdateToken updates token metadata without changing the secret.
func (h *Handler) UpdateToken(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "api token")
	if err != nil {
		return err
	}
	var req tokenRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	token, err := h.usecase.UpdateToken(c.Request().Context(), usecase.UpdateTokenInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Active:      req.Active,
		ExpiresAt:   req.ExpiresAt,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "update", "api_token", strconv.FormatInt(token.ID, 10), "updated api token"); err != nil {
		return err
	}
	return httpresp.OK(c, token)
}

// DeleteToken revokes an API token.
func (h *Handler) DeleteToken(c *echo.Context) error {
	if err := h.ready(); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "api token")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteToken(c.Request().Context(), id); err != nil {
		return err
	}
	if err := h.recordOperation(c, "delete", "api_token", strconv.FormatInt(id, 10), "revoked api token"); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{ID: id})
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
		return apperr.New(apperr.ErrInternalServer, "api token handler is not configured")
	}
	return nil
}

func listInput(c *echo.Context) (usecase.ListInput, error) {
	page, pageSize, err := httpreq.Pagination(c, defaultPageSize)
	if err != nil {
		return usecase.ListInput{}, err
	}
	adminID, err := httpreq.QueryInt64(c, "admin_id", 0)
	if err != nil {
		return usecase.ListInput{}, err
	}
	var active *bool
	if c.QueryParam("active") != "" {
		value, err := httpreq.QueryBool(c, "active", false)
		if err != nil {
			return usecase.ListInput{}, err
		}
		active = &value
	}
	return usecase.ListInput{Page: page, PageSize: pageSize, AdminID: adminID, Active: active}, nil
}

func tokenInputFromRequest(req tokenRequest) usecase.TokenInput {
	return usecase.TokenInput{
		AdminID:     req.AdminID,
		RoleID:      req.RoleID,
		Name:        req.Name,
		Description: req.Description,
		Active:      req.Active,
		Days:        req.Days,
		ExpiresAt:   req.ExpiresAt,
	}
}

type deletedResponse struct {
	ID int64 `json:"id"`
}
