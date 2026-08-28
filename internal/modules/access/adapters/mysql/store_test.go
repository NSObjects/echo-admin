package mysql

import (
	"reflect"
	"strings"
	"testing"

	"github.com/NSObjects/echo-admin/internal/modules/access/domain"
)

func planUpgradeCatalog() authorizationCatalog {
	return authorizationCatalog{
		permissionTokens: permissionCatalogTokens(),
		apiIDs:           []int64{1, 2, 3},
		menuIDs:          []int64{10, 20},
		buttonIDs:        []int64{100, 200},
	}
}

func assertRoleUpgradePlan(t *testing.T, got, want roleUpgradePlan) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("plan =\n got  %+v\n want %+v", got, want)
	}
}

func TestPlanRoleUpgradeRootReceivesCompleteCatalog(t *testing.T) {
	catalog := planUpgradeCatalog()
	want := roleUpgradePlan{
		permissions: catalog.permissionTokens,
		apiIDs:      []int64{1, 2, 3},
		menuIDs:     []int64{10, 20},
		buttonIDs:   []int64{100, 200},
	}

	withGrants := roleAuthorization{
		code:        domain.RoleCodeSuperAdmin,
		permissions: []string{domain.PermissionAdminRead},
		apiIDs:      []int64{1},
		menuIDs:     []int64{10},
		buttonIDs:   []int64{100},
	}
	got, err := planRoleUpgrade(withGrants, catalog)
	if err != nil {
		t.Fatalf("planRoleUpgrade(root with grants) error = %v", err)
	}
	assertRoleUpgradePlan(t, got, want)

	empty := roleAuthorization{code: domain.RoleCodeSuperAdmin}
	got, err = planRoleUpgrade(empty, catalog)
	if err != nil {
		t.Fatalf("planRoleUpgrade(root with empty grants) error = %v", err)
	}
	assertRoleUpgradePlan(t, got, want)
}

func TestPlanRoleUpgradeRetainsOrdinaryGrants(t *testing.T) {
	catalog := planUpgradeCatalog()

	got, err := planRoleUpgrade(roleAuthorization{
		code:        "operator",
		permissions: []string{domain.PermissionAdminRead, "garbage:token"},
		apiIDs:      []int64{3, 9, 9, 1},
		menuIDs:     []int64{10, 30},
		buttonIDs:   []int64{200, 300},
	}, catalog)
	if err != nil {
		t.Fatalf("planRoleUpgrade(ordinary) error = %v", err)
	}
	assertRoleUpgradePlan(t, got, roleUpgradePlan{
		permissions: []string{domain.PermissionAdminRead},
		apiIDs:      []int64{3, 1},
		menuIDs:     []int64{10},
		buttonIDs:   []int64{200},
	})

	_, err = planRoleUpgrade(roleAuthorization{code: "operator", permissions: []string{"garbage:token"}}, catalog)
	if err == nil || !strings.Contains(err.Error(), "operator") {
		t.Fatalf("planRoleUpgrade(no valid permissions) error = %v, want error naming role", err)
	}
}

func TestPlanRoleUpgradeMapsRetiredTokens(t *testing.T) {
	catalog := planUpgradeCatalog()

	got, err := planRoleUpgrade(roleAuthorization{
		code:        "operator",
		permissions: []string{retiredPermissionAPICreate, retiredPermissionAPIDelete},
	}, catalog)
	if err != nil {
		t.Fatalf("planRoleUpgrade(retired management) error = %v", err)
	}
	assertRoleUpgradePlan(t, got, roleUpgradePlan{
		permissions: []string{domain.PermissionAPIRead},
		apiIDs:      []int64{},
		menuIDs:     []int64{},
		buttonIDs:   []int64{},
	})

	got, err = planRoleUpgrade(roleAuthorization{
		code:        "operator",
		permissions: []string{retiredPermissionAPIUpdate},
	}, catalog)
	if err != nil {
		t.Fatalf("planRoleUpgrade(retired update) error = %v", err)
	}
	assertRoleUpgradePlan(t, got, roleUpgradePlan{
		permissions: []string{domain.PermissionAPIRead, domain.PermissionAPIGrant},
		apiIDs:      []int64{},
		menuIDs:     []int64{},
		buttonIDs:   []int64{},
	})
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
