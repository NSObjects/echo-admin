package mysql

import (
	"reflect"
	"testing"

	"github.com/NSObjects/echo-admin/internal/modules/access/domain"
)

func TestUpgradeRolePermissionsMapsRetiredAPIPermissions(t *testing.T) {
	tests := []struct {
		name    string
		current []string
		root    bool
		want    []string
	}{
		{
			name:    "ordinary role keeps current grants and maps api update",
			current: []string{domain.PermissionAdminRead, "api:create", "api:update", "api:delete"},
			want:    []string{domain.PermissionAdminRead, domain.PermissionAPIRead, domain.PermissionAPIGrant},
		},
		{
			name:    "ordinary role maps retired api management to read",
			current: []string{"api:create"},
			want:    []string{domain.PermissionAPIRead},
		},
		{
			name:    "root receives exact current catalog",
			current: []string{"api:create"},
			root:    true,
			want:    permissionCatalogTokens(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := upgradeRolePermissions(tt.current, tt.root)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("upgradeRolePermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetainCatalogIDs(t *testing.T) {
	got := retainCatalogIDs([]int64{9, 2, 2, 7}, []int64{2, 4, 9})
	want := []int64{9, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("retainCatalogIDs() = %v, want %v", got, want)
	}
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
