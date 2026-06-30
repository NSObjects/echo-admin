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
		if seed.description == "" {
			t.Fatalf("api seed %s %s description is empty", seed.method, seed.path)
		}
		if seed.group == "" {
			t.Fatalf("api seed %s %s group is empty", seed.method, seed.path)
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

func TestAPISeedsCoverRegisteredRoutes(t *testing.T) {
	seen := map[string]struct{}{}
	for _, seed := range apiSeeds {
		seen[seed.method+" "+seed.path] = struct{}{}
	}
	for _, key := range registeredAPISeedRoutes {
		if _, ok := seen[key]; !ok {
			t.Fatalf("api seed missing required route %s", key)
		}
	}
}

var registeredAPISeedRoutes = []string{
	"GET /api/health",
	"GET /api/info",
	"GET /api/ready",
	"GET /api/capabilities",
	"GET /api/setup/state",
	"POST /api/setup",
	"POST /api/auth/login",
	"POST /api/auth/logout",
	"POST /api/auth/logout-others",
	"POST /api/auth/password",
	"POST /api/auth/role",
	"GET /api/auth/me",
	"PATCH /api/auth/me",
	"GET /api/admins",
	"POST /api/admins",
	"PATCH /api/admins/:id",
	"DELETE /api/admins/:id",
	"GET /api/roles",
	"POST /api/roles",
	"PATCH /api/roles/:id",
	"DELETE /api/roles/:id",
	"POST /api/roles/:id/copy",
	"GET /api/roles/:id/admins",
	"PUT /api/roles/:id/admins",
	"GET /api/permissions",
	"GET /api/apis",
	"GET /api/apis/groups",
	"POST /api/apis",
	"POST /api/apis/batch-delete",
	"GET /api/apis/:id",
	"PATCH /api/apis/:id",
	"DELETE /api/apis/:id",
	"GET /api/apis/:id/roles",
	"PUT /api/apis/:id/roles",
	"GET /api/api-tokens",
	"POST /api/api-tokens",
	"PATCH /api/api-tokens/:id",
	"DELETE /api/api-tokens/:id",
	"GET /api/menus",
	"POST /api/menus",
	"GET /api/menus/:id",
	"PATCH /api/menus/:id",
	"DELETE /api/menus/:id",
	"GET /api/menus/:id/roles",
	"PUT /api/menus/:id/roles",
	"GET /api/system/configs",
	"PUT /api/system/configs/:key",
	"DELETE /api/system/configs/:key",
	"GET /api/system/params",
	"POST /api/system/params",
	"POST /api/system/params/batch-delete",
	"GET /api/system/params/key/:key",
	"GET /api/system/params/:id",
	"PATCH /api/system/params/:id",
	"DELETE /api/system/params/:id",
	"GET /api/system/versions",
	"POST /api/system/versions",
	"POST /api/system/versions/export",
	"POST /api/system/versions/import",
	"POST /api/system/versions/batch-delete",
	"GET /api/system/versions/:id",
	"GET /api/system/versions/:id/download",
	"PATCH /api/system/versions/:id",
	"DELETE /api/system/versions/:id",
	"GET /api/dictionaries",
	"POST /api/dictionaries",
	"GET /api/dictionaries/export",
	"POST /api/dictionaries/import",
	"PATCH /api/dictionaries/:code",
	"DELETE /api/dictionaries/:code",
	"POST /api/dictionaries/:code/items",
	"PATCH /api/dictionaries/:code/items/:item_id",
	"DELETE /api/dictionaries/:code/items/:item_id",
	"GET /api/file-categories",
	"POST /api/file-categories",
	"PATCH /api/file-categories/:id",
	"DELETE /api/file-categories/:id",
	"GET /api/files",
	"POST /api/files",
	"POST /api/files/import-url",
	"PATCH /api/files/:id/name",
	"DELETE /api/files/:id",
	"GET /api/uploads/*",
	"GET /api/logs/operations",
	"GET /api/logs/operations/:id",
	"DELETE /api/logs/operations/:id",
	"POST /api/logs/operations/batch-delete",
	"GET /api/logs/logins",
	"GET /api/logs/logins/:id",
	"DELETE /api/logs/logins/:id",
	"POST /api/logs/logins/batch-delete",
	"GET /api/logs/errors",
	"GET /api/logs/errors/:id",
	"POST /api/logs/errors/:id/resolve",
	"DELETE /api/logs/errors/:id/resolve",
	"DELETE /api/logs/errors/:id",
	"POST /api/logs/errors/batch-delete",
}

func TestMenuSeedsHaveComponentsAndUniqueButtons(t *testing.T) {
	seenPaths := map[string]struct{}{}
	for _, seed := range defaultMenuSeeds {
		if seed.path == "" {
			t.Fatal("menu seed path is empty")
		}
		if seed.component == "" {
			t.Fatalf("menu seed %s component is empty", seed.path)
		}
		if _, ok := seenPaths[seed.path]; ok {
			t.Fatalf("menu seed path %s is duplicated", seed.path)
		}
		if seed.parentPath != "" {
			if _, ok := seenPaths[seed.parentPath]; !ok {
				t.Fatalf("menu seed %s parent %s must be declared before child", seed.path, seed.parentPath)
			}
		}
		seenPaths[seed.path] = struct{}{}
		seenButtons := map[string]struct{}{}
		for _, button := range seed.buttons {
			if button.name == "" {
				t.Fatalf("menu seed %s has empty button name", seed.path)
			}
			if button.description == "" {
				t.Fatalf("menu seed %s button %s description is empty", seed.path, button.name)
			}
			if _, ok := seenButtons[button.name]; ok {
				t.Fatalf("menu seed %s button %s is duplicated", seed.path, button.name)
			}
			seenButtons[button.name] = struct{}{}
		}
	}
}

func TestMenuSeedsUseBackOfficeGroups(t *testing.T) {
	expectedParents := map[string]string{
		"/admins":       "/access",
		"/roles":        "/access",
		"/menus":        "/access",
		"/apis":         "/access",
		"/api-tokens":   "/access",
		"/configs":      "/system",
		"/params":       "/system",
		"/versions":     "/system",
		"/dictionaries": "/system",
		"/files":        "/resources",
		"/logs":         "/audit",
	}
	expectedGroups := map[string]struct{}{
		"/access":    {},
		"/system":    {},
		"/resources": {},
		"/audit":     {},
	}
	seenGroups := map[string]struct{}{}

	for _, seed := range defaultMenuSeeds {
		if wantParent, ok := expectedParents[seed.path]; ok && seed.parentPath != wantParent {
			t.Fatalf("menu seed %s parentPath = %q, want %q", seed.path, seed.parentPath, wantParent)
		}
		if _, ok := expectedGroups[seed.path]; !ok {
			continue
		}
		if seed.parentPath != "" {
			t.Fatalf("menu group %s parentPath = %q, want empty", seed.path, seed.parentPath)
		}
		if seed.permission != "" {
			t.Fatalf("menu group %s permission = %q, want empty", seed.path, seed.permission)
		}
		if len(seed.buttons) != 0 {
			t.Fatalf("menu group %s buttons = %d, want 0", seed.path, len(seed.buttons))
		}
		seenGroups[seed.path] = struct{}{}
	}

	for group := range expectedGroups {
		if _, ok := seenGroups[group]; !ok {
			t.Fatalf("menu group %s is missing", group)
		}
	}
}
