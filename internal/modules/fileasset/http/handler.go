// Package fileassethttp adapts file metadata HTTP requests to the file usecase.
package fileassethttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	"github.com/NSObjects/echo-admin/internal/modules/fileasset/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpreq"
	"github.com/NSObjects/echo-admin/internal/platform/server/httpresp"
)

const (
	defaultPageSize = 20
	maxUploadBytes  = 10 << 20
)

// Authorizer checks whether the current request can perform an action.
type Authorizer interface {
	RequirePermission(context.Context, string) error
}

// OperationRecorder records file mutations for audit.
type OperationRecorder interface {
	RecordOperation(context.Context, auditusecase.OperationInput) (auditusecase.OperationLog, error)
}

// Handler adapts file HTTP requests to the file usecase.
type Handler struct {
	usecase   *usecase.Usecase
	auth      Authorizer
	operation OperationRecorder
	uploadDir string
}

// New creates a file HTTP handler.
func New(uc *usecase.Usecase, auth Authorizer, operation OperationRecorder, uploadDir string) *Handler {
	return &Handler{usecase: uc, auth: auth, operation: operation, uploadDir: uploadDir}
}

// Register mounts file routes on group.
func Register(group *echo.Group, handler *Handler) {
	group.GET("/files", handler.ListFiles)
	group.POST("/files", handler.UploadFile)
	group.Static("/uploads", handler.uploadDir)
}

// ListFiles returns uploaded file records.
func (h *Handler) ListFiles(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionFileUpload); err != nil {
		return err
	}
	input, err := listInput(c)
	if err != nil {
		return err
	}
	output, err := h.usecase.ListFiles(c.Request().Context(), input)
	if err != nil {
		return err
	}
	return paginated(c, output.Items, output.Page, output.PageSize, output.Total)
}

// UploadFile stores one uploaded file and records its metadata.
func (h *Handler) UploadFile(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionFileUpload); err != nil {
		return err
	}
	header, err := c.FormFile("file")
	if err != nil {
		return apperr.WrapBadRequest(err, "file is required")
	}
	saved, err := h.saveUploadedFile(c, header)
	if err != nil {
		return err
	}
	file, err := h.usecase.CreateFile(c.Request().Context(), saved)
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "upload", "file", file.URL, "uploaded file"); err != nil {
		return err
	}
	return httpresp.Created(c, file)
}

func (h *Handler) authorize(c *echo.Context, permission string) error {
	if err := h.ready(); err != nil {
		return err
	}
	return h.auth.RequirePermission(c.Request().Context(), permission)
}

func (h *Handler) saveUploadedFile(c *echo.Context, header *multipart.FileHeader) (usecase.FileInput, error) {
	if err := c.Request().Context().Err(); err != nil {
		return usecase.FileInput{}, err
	}
	cleanName, err := validateUploadHeader(header)
	if err != nil {
		return usecase.FileInput{}, err
	}
	if mkdirErr := os.MkdirAll(h.uploadDir, 0o755); mkdirErr != nil {
		return usecase.FileInput{}, fmt.Errorf("create upload dir: %w", mkdirErr)
	}

	source, err := header.Open()
	if err != nil {
		return usecase.FileInput{}, fmt.Errorf("open upload: %w", err)
	}

	storedName := uuid.NewString() + "-" + cleanName
	targetPath := filepath.Join(h.uploadDir, storedName)
	written, writeErr := writeUploadTarget(source, targetPath)
	closeErr := source.Close()
	if writeErr != nil {
		if closeErr != nil {
			return usecase.FileInput{}, errors.Join(writeErr, fmt.Errorf("close upload source: %w", closeErr))
		}
		return usecase.FileInput{}, writeErr
	}
	if closeErr != nil {
		return usecase.FileInput{}, errors.Join(fmt.Errorf("close upload source: %w", closeErr), os.Remove(targetPath))
	}
	if err := c.Request().Context().Err(); err != nil {
		return usecase.FileInput{}, errors.Join(err, os.Remove(targetPath))
	}
	return usecase.FileInput{
		Name:        cleanName,
		URL:         "/api/uploads/" + storedName,
		Size:        written,
		ContentType: contentType(header),
	}, nil
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
	if h == nil || h.usecase == nil || h.auth == nil || h.operation == nil || strings.TrimSpace(h.uploadDir) == "" {
		return apperr.New(apperr.ErrInternalServer, "file handler is not configured")
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

func validateUploadHeader(header *multipart.FileHeader) (string, error) {
	if header == nil {
		return "", apperr.NewBadRequest("file is required")
	}
	if header.Size <= 0 || header.Size > maxUploadBytes {
		return "", apperr.NewBadRequest("invalid file size")
	}
	return cleanUploadName(header.Filename)
}

func writeUploadTarget(source io.Reader, targetPath string) (int64, error) {
	target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return 0, fmt.Errorf("create upload file: %w", err)
	}

	written, err := io.Copy(target, io.LimitReader(source, maxUploadBytes+1))
	if err != nil {
		return 0, errors.Join(fmt.Errorf("write upload: %w", err), cleanupUploadTarget(target, targetPath))
	}
	if written > maxUploadBytes {
		if err := cleanupUploadTarget(target, targetPath); err != nil {
			return 0, fmt.Errorf("cleanup oversized upload: %w", err)
		}
		return 0, apperr.NewBadRequest("invalid file size")
	}
	if err := target.Close(); err != nil {
		return 0, fmt.Errorf("close upload file: %w", err)
	}
	return written, nil
}

func cleanupUploadTarget(target *os.File, path string) error {
	return errors.Join(target.Close(), os.Remove(path))
}

func cleanUploadName(name string) (string, error) {
	cleaned := filepath.Base(strings.TrimSpace(name))
	if cleaned == "." || cleaned == string(filepath.Separator) || cleaned == "" {
		return "", apperr.NewBadRequest("invalid file name")
	}
	if strings.Contains(cleaned, "/") || strings.Contains(cleaned, "\\") {
		return "", apperr.NewBadRequest("invalid file name")
	}
	return cleaned, nil
}

func contentType(header *multipart.FileHeader) string {
	if value := header.Header.Get(echo.HeaderContentType); value != "" {
		return value
	}
	return http.DetectContentType(nil)
}
