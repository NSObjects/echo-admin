package mysql

import (
	"testing"

	"github.com/NSObjects/echo-admin/internal/modules/access/domain"
)

func TestAPISeedsHaveRequiredFieldsAndUniqueRoutes(t *testing.T) {
	seen := map[string]struct{}{}
	for _, seed := range apiSeeds {
		if seed.method == "" {
			t.Fatal("api seed method is empty")
		}
		if seed.path == "" {
			t.Fatal("api seed path is empty")
		}
		if seed.name == "" {
			t.Fatalf("api seed %s %s name is empty", seed.method, seed.path)
		}
		key := seed.method + " " + seed.path
		if _, ok := seen[key]; ok {
			t.Fatalf("api seed %s is duplicated", key)
		}
		seen[key] = struct{}{}
	}
}

func TestAPISeedPermissionsExistInCatalog(t *testing.T) {
	permissions := map[string]struct{}{}
	for _, permission := range domain.PermissionCatalog() {
		permissions[permission.Token] = struct{}{}
	}

	for _, seed := range apiSeeds {
		key := seed.method + " " + seed.path
		if seed.public && seed.permission != "" {
			t.Fatalf("public api seed %s has permission %q", key, seed.permission)
		}
		if seed.permission == "" {
			continue
		}
		if _, ok := permissions[seed.permission]; !ok {
			t.Fatalf("api seed %s permission = %q, want a catalog permission", key, seed.permission)
		}
	}
}

func TestAPISeedsCoverCriticalMutationRoutes(t *testing.T) {
	seen := map[string]struct{}{}
	for _, seed := range apiSeeds {
		seen[seed.method+" "+seed.path] = struct{}{}
	}
	requiredRoutes := []string{
		"DELETE /api/admins/:id",
		"DELETE /api/roles/:id",
		"POST /api/roles/:id/copy",
		"DELETE /api/menus/:id",
		"DELETE /api/dictionaries/:code",
		"DELETE /api/dictionaries/:code/items/:item_id",
	}
	for _, key := range requiredRoutes {
		if _, ok := seen[key]; !ok {
			t.Fatalf("api seed missing required route %s", key)
		}
	}
}
