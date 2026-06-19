package domain

import (
	"errors"
	"testing"
	"time"
)

func TestRestoreRoleRejectsInvalidPermissionToken(t *testing.T) {
	_, err := RestoreRole(1, 0, "operator", "Operator", []string{"admin"}, []int64{1}, DefaultRolePath, true, time.Now(), time.Now())
	if !errors.Is(err, ErrInvalidPermission) {
		t.Fatalf("RestoreRole() error = %v, want %v", err, ErrInvalidPermission)
	}
}

func TestRestoreMenuRejectsInvalidPermissionToken(t *testing.T) {
	_, err := RestoreMenu(1, 0, "Admins", "/admins", "user", "admin", 10, true, time.Now(), time.Now())
	if !errors.Is(err, ErrInvalidPermission) {
		t.Fatalf("RestoreMenu() error = %v, want %v", err, ErrInvalidPermission)
	}
}
