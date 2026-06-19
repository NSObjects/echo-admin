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

// OperationLog records one state-changing back-office operation.
type OperationLog struct {
	id         int64
	actorID    int64
	action     string
	resource   string
	resourceID string
	method     string
	path       string
	ip         string
	userAgent  string
	success    bool
	message    string
	createdAt  time.Time
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
		id:         id,
		actorID:    actorID,
		action:     action,
		resource:   resource,
		resourceID: strings.TrimSpace(resourceID),
		method:     strings.TrimSpace(method),
		path:       strings.TrimSpace(path),
		ip:         strings.TrimSpace(ip),
		userAgent:  strings.TrimSpace(userAgent),
		success:    success,
		message:    strings.TrimSpace(message),
		createdAt:  createdAt,
	}, nil
}

// ID returns the persisted operation log id.
func (l OperationLog) ID() int64 { return l.id }

// ActorID returns the admin id that initiated the operation.
func (l OperationLog) ActorID() int64 { return l.actorID }

// Action returns the operation action token.
func (l OperationLog) Action() string { return l.action }

// Resource returns the operated resource token.
func (l OperationLog) Resource() string { return l.resource }

// ResourceID returns the optional resource identifier.
func (l OperationLog) ResourceID() string { return l.resourceID }

// Method returns the HTTP method observed at the boundary.
func (l OperationLog) Method() string { return l.method }

// Path returns the HTTP path observed at the boundary.
func (l OperationLog) Path() string { return l.path }

// IP returns the client IP captured at the boundary.
func (l OperationLog) IP() string { return l.ip }

// UserAgent returns the user agent captured at the boundary.
func (l OperationLog) UserAgent() string { return l.userAgent }

// Success reports whether the operation completed successfully.
func (l OperationLog) Success() bool { return l.success }

// Message returns a safe operator-facing operation message.
func (l OperationLog) Message() string { return l.message }

// CreatedAt returns the creation timestamp.
func (l OperationLog) CreatedAt() time.Time { return l.createdAt }

// LoginLog records one sign-in attempt.
type LoginLog struct {
	id        int64
	adminID   int64
	username  string
	ip        string
	userAgent string
	success   bool
	reason    string
	createdAt time.Time
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
		id:        id,
		adminID:   adminID,
		username:  username,
		ip:        strings.TrimSpace(ip),
		userAgent: strings.TrimSpace(userAgent),
		success:   success,
		reason:    strings.TrimSpace(reason),
		createdAt: createdAt,
	}, nil
}

// ID returns the persisted login log id.
func (l LoginLog) ID() int64 { return l.id }

// AdminID returns the admin id when the username resolved to an admin.
func (l LoginLog) AdminID() int64 { return l.adminID }

// Username returns the attempted username.
func (l LoginLog) Username() string { return l.username }

// IP returns the client IP captured at the boundary.
func (l LoginLog) IP() string { return l.ip }

// UserAgent returns the user agent captured at the boundary.
func (l LoginLog) UserAgent() string { return l.userAgent }

// Success reports whether sign-in succeeded.
func (l LoginLog) Success() bool { return l.success }

// Reason returns a safe failure or success reason.
func (l LoginLog) Reason() string { return l.reason }

// CreatedAt returns the creation timestamp.
func (l LoginLog) CreatedAt() time.Time { return l.createdAt }

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
