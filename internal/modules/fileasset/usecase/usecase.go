package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/NSObjects/echo-admin/internal/modules/fileasset/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/pagination"
)

// CreateFile stores uploaded file metadata.
func (u *Usecase) CreateFile(ctx context.Context, input FileInput) (FileObject, error) {
	if err := u.ready(); err != nil {
		return FileObject{}, err
	}
	file, err := domain.RestoreFileObject(0, input.Name, input.URL, input.Size, input.ContentType, time.Time{})
	if err != nil {
		return FileObject{}, mapDomainError(err)
	}
	created, err := u.store.CreateFile(ctx, file)
	if err != nil {
		return FileObject{}, err
	}
	return fromFile(created), nil
}

// ListFiles returns paginated uploaded file records.
func (u *Usecase) ListFiles(ctx context.Context, input ListInput) (ListOutput, error) {
	if err := u.ready(); err != nil {
		return ListOutput{}, err
	}
	filter, err := normalizeListInput(input)
	if err != nil {
		return ListOutput{}, err
	}
	files, total, err := u.store.ListFiles(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}
	return ListOutput{
		Items:    mapFiles(files),
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Total:    total,
	}, nil
}

func (u *Usecase) ready() error {
	if u == nil || u.store == nil {
		return apperr.New(apperr.ErrInternalServer, "file store is not configured")
	}
	return nil
}

func normalizeListInput(input ListInput) (ListFilter, error) {
	window, err := pagination.Normalize(input.Page, input.PageSize, pagination.Options{
		DefaultPageSize: defaultPageSize,
		MaxPageSize:     maxPageSize,
	})
	if err != nil {
		return ListFilter{}, apperr.NewBadRequest("invalid pagination")
	}
	return ListFilter{Offset: window.Offset, Limit: window.Limit, Page: window.Page, PageSize: window.PageSize}, nil
}

func mapFiles(files []domain.FileObject) []FileObject {
	out := make([]FileObject, 0, len(files))
	for _, file := range files {
		out = append(out, fromFile(file))
	}
	return out
}

func fromFile(file domain.FileObject) FileObject {
	return FileObject{
		ID:          file.ID(),
		Name:        file.Name(),
		URL:         file.URL(),
		Size:        file.Size(),
		ContentType: file.ContentType(),
		CreatedAt:   file.CreatedAt(),
	}
}

func mapDomainError(err error) error {
	for _, entry := range domainErrorMessages {
		if errors.Is(err, entry.err) {
			return apperr.NewBadRequest(entry.message)
		}
	}
	return err
}

var domainErrorMessages = []struct {
	err     error
	message string
}{
	{domain.ErrInvalidFileID, "invalid file id"},
	{domain.ErrInvalidFileName, "invalid file name"},
	{domain.ErrInvalidFileURL, "invalid file url"},
	{domain.ErrInvalidFileSize, "invalid file size"},
	{domain.ErrInvalidContentType, "invalid content type"},
}
