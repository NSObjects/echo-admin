// Package usecase coordinates operation and login audit workflows.
package usecase

import (
	"context"
	"time"

	"github.com/NSObjects/echo-admin/internal/modules/audit/domain"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// Store persists operation and login logs.
type Store interface {
	RecordOperation(context.Context, domain.OperationLog) (domain.OperationLog, error)
	ListOperationLogs(context.Context, ListFilter) ([]domain.OperationLog, int, error)
	RecordLogin(context.Context, domain.LoginLog) (domain.LoginLog, error)
	ListLoginLogs(context.Context, ListFilter) ([]domain.LoginLog, int, error)
}

// Usecase coordinates audit log rules.
type Usecase struct {
	store Store
}

// New creates an audit usecase.
func New(store Store) *Usecase {
	return &Usecase{store: store}
}

// OperationInput carries an operation log request from adapters.
type OperationInput struct {
	ActorID    int64
	Action     string
	Resource   string
	ResourceID string
	Method     string
	Path       string
	IP         string
	UserAgent  string
	Success    bool
	Message    string
}

// LoginInput carries a sign-in attempt audit request.
type LoginInput struct {
	AdminID   int64
	Username  string
	IP        string
	UserAgent string
	Success   bool
	Reason    string
}

// ListInput carries pagination for audit lists.
type ListInput struct {
	Page     int
	PageSize int
}

// ListFilter is the validated store-facing pagination window.
type ListFilter struct {
	Offset   int
	Limit    int
	Page     int
	PageSize int
}

// OperationListOutput is a paginated operation log result.
type OperationListOutput struct {
	Items    []OperationLog
	Page     int
	PageSize int
	Total    int
}

// LoginListOutput is a paginated login log result.
type LoginListOutput struct {
	Items    []LoginLog
	Page     int
	PageSize int
	Total    int
}

// OperationLog is the adapter-facing operation log DTO.
type OperationLog struct {
	ID         int64     `json:"id"`
	ActorID    int64     `json:"actor_id"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	Success    bool      `json:"success"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

// LoginLog is the adapter-facing login log DTO.
type LoginLog struct {
	ID        int64     `json:"id"`
	AdminID   int64     `json:"admin_id"`
	Username  string    `json:"username"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}
