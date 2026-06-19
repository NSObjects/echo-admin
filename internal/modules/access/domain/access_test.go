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

func TestPermissionCatalogCoversFoundationPermissions(t *testing.T) {
	catalog := map[string]struct{}{}
	for _, permission := range PermissionCatalog() {
		if permission.Token == "" || permission.Resource == "" || permission.Action == "" || permission.Name == "" {
			t.Fatalf("permission catalog contains incomplete definition: %#v", permission)
		}
		if _, ok := catalog[permission.Token]; ok {
			t.Fatalf("permission catalog contains duplicate token %q", permission.Token)
		}
		catalog[permission.Token] = struct{}{}
	}

	required := []string{
		PermissionAdminRead,
		PermissionAdminCreate,
		PermissionAdminUpdate,
		PermissionAdminDelete,
		PermissionRoleRead,
		PermissionRoleCreate,
		PermissionRoleUpdate,
		PermissionRoleDelete,
		PermissionMenuRead,
		PermissionMenuCreate,
		PermissionMenuUpdate,
		PermissionMenuDelete,
		PermissionConfigRead,
		PermissionConfigUpdate,
		PermissionDictRead,
		PermissionDictCreate,
		PermissionDictUpdate,
		PermissionDictDelete,
		PermissionFileRead,
		PermissionFileUpload,
		PermissionLogRead,
	}
	for _, token := range required {
		if _, ok := catalog[token]; !ok {
			t.Fatalf("permission catalog missing %q", token)
		}
	}
}
