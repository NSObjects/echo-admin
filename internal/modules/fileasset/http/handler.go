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
	RequireRoutePermission(context.Context, string, string, string) error
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
	group.GET("/file-categories", handler.ListCategories)
	group.POST("/file-categories", handler.CreateCategory)
	group.PATCH("/file-categories/:id", handler.UpdateCategory)
	group.DELETE("/file-categories/:id", handler.DeleteCategory)
	group.GET("/files", handler.ListFiles)
	group.POST("/files", handler.UploadFile)
	group.POST("/files/import-url", handler.ImportURL)
	group.PATCH("/files/:id/name", handler.RenameFile)
	group.DELETE("/files/:id", handler.DeleteFile)
	group.GET("/uploads/*", handler.ServeUpload)
}

// ListCategories returns the category tree used by file management.
func (h *Handler) ListCategories(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionFileRead); err != nil {
		return err
	}
	categories, err := h.usecase.ListCategories(c.Request().Context())
	if err != nil {
		return err
	}
	return httpresp.OK(c, categories)
}

// CreateCategory adds one file category.
func (h *Handler) CreateCategory(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionFileCategoryCreate); err != nil {
		return err
	}
	var req categoryRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	category, err := h.usecase.CreateCategory(c.Request().Context(), usecase.CategoryInput{
		Name:     req.Name,
		ParentID: req.ParentID,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(
		c,
		"create",
		"file_category",
		strconv.FormatInt(category.ID, 10),
		"created file category",
	); err != nil {
		return err
	}
	return httpresp.Created(c, category)
}

// UpdateCategory changes one file category.
func (h *Handler) UpdateCategory(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionFileCategoryUpdate); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "file category")
	if err != nil {
		return err
	}
	var req categoryRequest
	if bindErr := httpreq.BindAndValidate(c, &req); bindErr != nil {
		return bindErr
	}
	category, err := h.usecase.UpdateCategory(c.Request().Context(), usecase.UpdateCategoryInput{
		ID:       id,
		Name:     req.Name,
		ParentID: req.ParentID,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(
		c,
		"update",
		"file_category",
		strconv.FormatInt(category.ID, 10),
		"updated file category",
	); err != nil {
		return err
	}
	return httpresp.OK(c, category)
}

// DeleteCategory removes one file category without deleting files.
func (h *Handler) DeleteCategory(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionFileCategoryDelete); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "file category")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteCategory(c.Request().Context(), id); err != nil {
		return err
	}
	if err := h.recordOperation(
		c,
		"delete",
		"file_category",
		strconv.FormatInt(id, 10),
		"deleted file category",
	); err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{ID: id})
}

// ListFiles returns uploaded file records.
func (h *Handler) ListFiles(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionFileRead); err != nil {
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
	categoryID, err := formInt64(c, "category_id")
	if err != nil {
		return err
	}
	saved, err := h.saveUploadedFile(c, header)
	if err != nil {
		return err
	}
	saved.CategoryID = categoryID
	file, err := h.usecase.CreateFile(c.Request().Context(), saved)
	if err != nil {
		if cleanupErr := h.removeLocalUpload(usecase.FileObject{URL: saved.URL}); cleanupErr != nil {
			return errors.Join(err, cleanupErr)
		}
		return err
	}
	if err := h.recordOperation(c, "upload", "file", file.URL, "uploaded file"); err != nil {
		return err
	}
	return httpresp.Created(c, file)
}

// ImportURL registers an external HTTP(S) URL as a file asset.
func (h *Handler) ImportURL(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionFileUpload); err != nil {
		return err
	}
	var req importURLRequest
	if err := httpreq.BindAndValidate(c, &req); err != nil {
		return err
	}
	file, err := h.usecase.ImportURL(c.Request().Context(), usecase.URLImportInput{
		Name:       req.Name,
		URL:        req.URL,
		CategoryID: req.CategoryID,
	})
	if err != nil {
		return err
	}
	if err := h.recordOperation(c, "import_url", "file", file.URL, "imported file url"); err != nil {
		return err
	}
	return httpresp.Created(c, file)
}

// RenameFile updates one file display name.
func (h *Handler) RenameFile(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionFileUpdate); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "file")
	if err != nil {
		return err
	}
	var req renameFileRequest
	err = httpreq.BindAndValidate(c, &req)
	if err != nil {
		return err
	}
	file, err := h.usecase.RenameFile(c.Request().Context(), usecase.RenameInput{
		ID:   id,
		Name: req.Name,
	})
	if err != nil {
		return err
	}
	err = h.recordOperation(c, "rename", "file", strconv.FormatInt(file.ID, 10), "renamed file")
	if err != nil {
		return err
	}
	return httpresp.OK(c, file)
}

// DeleteFile removes one file metadata record and its local upload when present.
func (h *Handler) DeleteFile(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionFileDelete); err != nil {
		return err
	}
	id, err := httpreq.PathID(c, "id", "file")
	if err != nil {
		return err
	}
	file, err := h.usecase.DeleteFile(c.Request().Context(), id)
	if err != nil {
		return err
	}
	err = h.removeLocalUpload(file)
	if err != nil {
		return err
	}
	err = h.recordOperation(c, "delete", "file", strconv.FormatInt(file.ID, 10), "deleted file")
	if err != nil {
		return err
	}
	return httpresp.OK(c, deletedResponse{ID: file.ID})
}

// ServeUpload returns one local uploaded file after the same route/API grant
// check used by file metadata reads.
func (h *Handler) ServeUpload(c *echo.Context) error {
	if err := h.authorize(c, accessdomain.PermissionFileRead); err != nil {
		return err
	}
	storedName, err := cleanStoredUploadName(c.Param("*"))
	if err != nil {
		return err
	}
	return c.FileFS(storedName, os.DirFS(h.uploadDir))
}

func (h *Handler) authorize(c *echo.Context, permission string) error {
	if err := h.ready(); err != nil {
		return err
	}
	return h.auth.RequireRoutePermission(c.Request().Context(), permission, c.Request().Method, c.Path())
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

func (h *Handler) removeLocalUpload(file usecase.FileObject) error {
	const uploadURLPrefix = "/api/uploads/"
	if !strings.HasPrefix(file.URL, uploadURLPrefix) {
		return nil
	}
	storedName := strings.TrimPrefix(file.URL, uploadURLPrefix)
	if storedName == "" || filepath.Base(storedName) != storedName {
		return fmt.Errorf("remove upload file: invalid stored name")
	}
	targetPath := filepath.Join(h.uploadDir, storedName)
	if err := os.Remove(targetPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove upload file: %w", err)
	}
	return nil
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
	categoryID, err := httpreq.QueryInt64(c, "category_id", 0)
	if err != nil {
		return usecase.ListInput{}, err
	}
	return usecase.ListInput{Page: page, PageSize: pageSize, CategoryID: categoryID}, nil
}

func formInt64(c *echo.Context, name string) (int64, error) {
	raw := strings.TrimSpace(c.FormValue(name))
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 {
		return 0, apperr.NewBadRequest("invalid " + name)
	}
	return value, nil
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

func cleanStoredUploadName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || filepath.Base(name) != name || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return "", apperr.NewBadRequest("invalid upload file")
	}
	return name, nil
}

func contentType(header *multipart.FileHeader) string {
	if value := header.Header.Get(echo.HeaderContentType); value != "" {
		return value
	}
	return http.DetectContentType(nil)
}

type deletedResponse struct {
	ID int64 `json:"id"`
}
