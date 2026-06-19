// Package domain contains audit log business rules.
package domain

import (
	"errors"
	"strings"
	"time"
)

// Audit validation errors.
var (
	ErrInvalidLogID         = errors.New("invalid log id")
	ErrInvalidAction        = errors.New("invalid action")
	ErrInvalidResource      = errors.New("invalid resource")
	ErrInvalidLoginUsername = errors.New("invalid login username")
)

// OperationLog is a validated record of one state-changing back-office operation.
type OperationLog struct {
	ID         int64
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
	CreatedAt  time.Time
}

// RestoreOperationLog rebuilds an operation log from a trusted store representation.
func RestoreOperationLog(id, actorID int64, action, resource, resourceID, method, path, ip, userAgent string, success bool, message string, createdAt time.Time) (OperationLog, error) {
	action = normalizeToken(action)
	resource = normalizeToken(resource)
	if id < 0 {
		return OperationLog{}, ErrInvalidLogID
	}
	if action == "" || len(action) > 80 {
		return OperationLog{}, ErrInvalidAction
	}
	if resource == "" || len(resource) > 80 {
		return OperationLog{}, ErrInvalidResource
	}
	return OperationLog{
		ID:         id,
		ActorID:    actorID,
		Action:     action,
		Resource:   resource,
		ResourceID: strings.TrimSpace(resourceID),
		Method:     strings.TrimSpace(method),
		Path:       strings.TrimSpace(path),
		IP:         strings.TrimSpace(ip),
		UserAgent:  strings.TrimSpace(userAgent),
		Success:    success,
		Message:    strings.TrimSpace(message),
		CreatedAt:  createdAt,
	}, nil
}

// LoginLog is a validated record of one sign-in attempt.
type LoginLog struct {
	ID        int64
	AdminID   int64
	Username  string
	IP        string
	UserAgent string
	Success   bool
	Reason    string
	CreatedAt time.Time
}

// RestoreLoginLog rebuilds a login log from a trusted store representation.
func RestoreLoginLog(id, adminID int64, username, ip, userAgent string, success bool, reason string, createdAt time.Time) (LoginLog, error) {
	username = normalizeToken(username)
	if id < 0 {
		return LoginLog{}, ErrInvalidLogID
	}
	if username == "" || len(username) > 64 {
		return LoginLog{}, ErrInvalidLoginUsername
	}
	return LoginLog{
		ID:        id,
		AdminID:   adminID,
		Username:  username,
		IP:        strings.TrimSpace(ip),
		UserAgent: strings.TrimSpace(userAgent),
		Success:   success,
		Reason:    strings.TrimSpace(reason),
		CreatedAt: createdAt,
	}, nil
}

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
