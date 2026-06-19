// Package mysql persists audit logs in MySQL.
package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/NSObjects/echo-admin/internal/modules/audit/domain"
	"github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
)

// Store persists operation and login logs in MySQL.
type Store struct {
	db *gorm.DB
}

// NewStore migrates the MySQL audit tables.
func NewStore(ctx context.Context, db *gorm.DB) (*Store, error) {
	if ctx == nil {
		return nil, errors.New("create audit store: nil context")
	}
	if db == nil {
		return nil, errors.New("create audit store: nil db")
	}
	if err := db.WithContext(ctx).AutoMigrate(&operationLogModel{}, &loginLogModel{}); err != nil {
		return nil, apperr.WrapDatabase(err, "migrate audit tables")
	}
	return &Store{db: db}, nil
}

// RecordOperation inserts one operation log.
func (s *Store) RecordOperation(ctx context.Context, log domain.OperationLog) (domain.OperationLog, error) {
	if err := ctx.Err(); err != nil {
		return domain.OperationLog{}, err
	}
	model := operationLogModelFromDomain(log, time.Now().UTC())
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.OperationLog{}, apperr.WrapDatabase(err, "record operation log")
	}
	return model.toDomain()
}

// ListOperationLogs returns operation logs ordered by id descending.
func (s *Store) ListOperationLogs(ctx context.Context, filter usecase.ListFilter) ([]domain.OperationLog, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	return listAuditRows(ctx, s.db, &operationLogModel{}, filter, "count operation logs", "list operation logs", operationLogModel.toDomain)
}

// RecordLogin inserts one login log.
func (s *Store) RecordLogin(ctx context.Context, log domain.LoginLog) (domain.LoginLog, error) {
	if err := ctx.Err(); err != nil {
		return domain.LoginLog{}, err
	}
	model := loginLogModelFromDomain(log, time.Now().UTC())
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.LoginLog{}, apperr.WrapDatabase(err, "record login log")
	}
	return model.toDomain()
}

// ListLoginLogs returns login logs ordered by id descending.
func (s *Store) ListLoginLogs(ctx context.Context, filter usecase.ListFilter) ([]domain.LoginLog, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	return listAuditRows(ctx, s.db, &loginLogModel{}, filter, "count login logs", "list login logs", loginLogModel.toDomain)
}

type operationLogModel struct {
	ID         int64     `gorm:"primaryKey"`
	ActorID    int64     `gorm:"not null;index"`
	Action     string    `gorm:"type:varchar(80);not null"`
	Resource   string    `gorm:"type:varchar(80);not null"`
	ResourceID string    `gorm:"type:varchar(120);not null"`
	Method     string    `gorm:"type:varchar(16);not null"`
	Path       string    `gorm:"type:varchar(255);not null"`
	IP         string    `gorm:"type:varchar(64);not null"`
	UserAgent  string    `gorm:"type:varchar(255);not null"`
	Success    bool      `gorm:"not null"`
	Message    string    `gorm:"type:varchar(255);not null"`
	CreatedAt  time.Time `gorm:"not null;index"`
}

func (operationLogModel) TableName() string {
	return "audit_operation_logs"
}

func operationLogModelFromDomain(log domain.OperationLog, createdAt time.Time) operationLogModel {
	return operationLogModel{
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
		CreatedAt:  coalesceTime(log.CreatedAt, createdAt),
	}
}

func (m operationLogModel) toDomain() (domain.OperationLog, error) {
	return domain.RestoreOperationLog(m.ID, m.ActorID, m.Action, m.Resource, m.ResourceID, m.Method, m.Path, m.IP, m.UserAgent, m.Success, m.Message, m.CreatedAt)
}

type loginLogModel struct {
	ID        int64     `gorm:"primaryKey"`
	AdminID   int64     `gorm:"not null;index"`
	Username  string    `gorm:"type:varchar(64);not null;index"`
	IP        string    `gorm:"type:varchar(64);not null"`
	UserAgent string    `gorm:"type:varchar(255);not null"`
	Success   bool      `gorm:"not null;index"`
	Reason    string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"not null;index"`
}

func (loginLogModel) TableName() string {
	return "audit_login_logs"
}

func loginLogModelFromDomain(log domain.LoginLog, createdAt time.Time) loginLogModel {
	return loginLogModel{
		ID:        log.ID,
		AdminID:   log.AdminID,
		Username:  log.Username,
		IP:        log.IP,
		UserAgent: log.UserAgent,
		Success:   log.Success,
		Reason:    log.Reason,
		CreatedAt: coalesceTime(log.CreatedAt, createdAt),
	}
}

func (m loginLogModel) toDomain() (domain.LoginLog, error) {
	return domain.RestoreLoginLog(m.ID, m.AdminID, m.Username, m.IP, m.UserAgent, m.Success, m.Reason, m.CreatedAt)
}

func coalesceTime(value, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}

func listAuditRows[M, D any](
	ctx context.Context,
	db *gorm.DB,
	model any,
	filter usecase.ListFilter,
	countOperation string,
	listOperation string,
	convert func(M) (D, error),
) ([]D, int, error) {
	var total int64
	query := db.WithContext(ctx).Model(model)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperr.WrapDatabase(err, countOperation)
	}
	var models []M
	err := query.Order("id DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&models).Error
	if err != nil {
		return nil, 0, apperr.WrapDatabase(err, listOperation)
	}
	items := make([]D, 0, len(models))
	for _, row := range models {
		item, err := convert(row)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, int(total), nil
}
