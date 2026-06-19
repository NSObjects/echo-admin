package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/NSObjects/echo-admin/internal/modules/audit/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/pagination"
)

// RecordOperation stores one operation log.
func (u *Usecase) RecordOperation(ctx context.Context, input OperationInput) (OperationLog, error) {
	if err := u.ready(); err != nil {
		return OperationLog{}, err
	}
	log, err := domain.RestoreOperationLog(0, input.ActorID, input.Action, input.Resource, input.ResourceID, input.Method, input.Path, input.IP, input.UserAgent, input.Success, input.Message, time.Time{})
	if err != nil {
		return OperationLog{}, mapDomainError(err)
	}
	created, err := u.store.RecordOperation(ctx, log)
	if err != nil {
		return OperationLog{}, err
	}
	return fromOperationLog(created), nil
}

// RecordLogin stores one login attempt log.
func (u *Usecase) RecordLogin(ctx context.Context, input LoginInput) (LoginLog, error) {
	if err := u.ready(); err != nil {
		return LoginLog{}, err
	}
	username := input.Username
	if username == "" {
		username = "unknown"
	}
	log, err := domain.RestoreLoginLog(0, input.AdminID, username, input.IP, input.UserAgent, input.Success, input.Reason, time.Time{})
	if err != nil {
		return LoginLog{}, mapDomainError(err)
	}
	created, err := u.store.RecordLogin(ctx, log)
	if err != nil {
		return LoginLog{}, err
	}
	return fromLoginLog(created), nil
}

// ListOperationLogs returns paginated operation logs.
func (u *Usecase) ListOperationLogs(ctx context.Context, input ListInput) (OperationListOutput, error) {
	if err := u.ready(); err != nil {
		return OperationListOutput{}, err
	}
	filter, err := normalizeListInput(input)
	if err != nil {
		return OperationListOutput{}, err
	}
	logs, total, err := u.store.ListOperationLogs(ctx, filter)
	if err != nil {
		return OperationListOutput{}, err
	}
	return OperationListOutput{
		Items:    mapOperationLogs(logs),
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Total:    total,
	}, nil
}

// ListLoginLogs returns paginated login logs.
func (u *Usecase) ListLoginLogs(ctx context.Context, input ListInput) (LoginListOutput, error) {
	if err := u.ready(); err != nil {
		return LoginListOutput{}, err
	}
	filter, err := normalizeListInput(input)
	if err != nil {
		return LoginListOutput{}, err
	}
	logs, total, err := u.store.ListLoginLogs(ctx, filter)
	if err != nil {
		return LoginListOutput{}, err
	}
	return LoginListOutput{
		Items:    mapLoginLogs(logs),
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Total:    total,
	}, nil
}

func (u *Usecase) ready() error {
	if u == nil || u.store == nil {
		return apperr.New(apperr.ErrInternalServer, "audit store is not configured")
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

func mapOperationLogs(logs []domain.OperationLog) []OperationLog {
	out := make([]OperationLog, 0, len(logs))
	for _, log := range logs {
		out = append(out, fromOperationLog(log))
	}
	return out
}

func mapLoginLogs(logs []domain.LoginLog) []LoginLog {
	out := make([]LoginLog, 0, len(logs))
	for _, log := range logs {
		out = append(out, fromLoginLog(log))
	}
	return out
}

func fromOperationLog(log domain.OperationLog) OperationLog {
	return OperationLog{
		ID:         log.ID,
		ActorID:    log.ActorID,
		Action:     log.Action,
		Resource:   log.Resource,
		ResourceID: log.ResourceID,
		Method:     log.Method,
		Path:       log.Path,
		IP:         log.IP,
		UserAgent:  log.UserAgent,
		Success:    log.Success,
		Message:    log.Message,
		CreatedAt:  log.CreatedAt,
	}
}

func fromLoginLog(log domain.LoginLog) LoginLog {
	return LoginLog{
		ID:        log.ID,
		AdminID:   log.AdminID,
		Username:  log.Username,
		IP:        log.IP,
		UserAgent: log.UserAgent,
		Success:   log.Success,
		Reason:    log.Reason,
		CreatedAt: log.CreatedAt,
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
	{domain.ErrInvalidLogID, "invalid log id"},
	{domain.ErrInvalidAction, "invalid action"},
	{domain.ErrInvalidResource, "invalid resource"},
	{domain.ErrInvalidLoginUsername, "invalid login username"},
}
