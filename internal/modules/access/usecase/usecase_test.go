package usecase_test

import (
	"context"
	"testing"
	"time"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	"github.com/NSObjects/echo-admin/internal/modules/access/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
)

func TestUpdateMenuPreservesCreatedAt(t *testing.T) {
	createdAt := time.Unix(1_700_000_000, 0).UTC()
	store := &storeSpy{menuCreatedAt: createdAt}
	uc := usecase.New(store)

	_, err := uc.UpdateMenu(context.Background(), usecase.UpdateMenuInput{
		ID:         1,
		ParentID:   0,
		Name:       "Updated Menu",
		Path:       "/updated",
		Icon:       "menu",
		Permission: accessdomain.PermissionLogRead,
		Sort:       20,
		Active:     true,
	})
	if err != nil {
		t.Fatalf("UpdateMenu() error = %v", err)
	}
	if got := store.updatedMenu.CreatedAt(); !got.Equal(createdAt) {
		t.Fatalf("updated CreatedAt = %s, want %s", got, createdAt)
	}
}

type storeSpy struct {
	menuCreatedAt time.Time
	updatedMenu   accessdomain.Menu
}

func (s *storeSpy) FindRoleByID(context.Context, int64) (accessdomain.Role, error) {
	return accessdomain.Role{}, apperr.NewNotFound("role")
}

func (s *storeSpy) ListRoles(context.Context, usecase.ListFilter) ([]accessdomain.Role, int, error) {
	return nil, 0, nil
}

func (s *storeSpy) CreateRole(context.Context, accessdomain.Role) (accessdomain.Role, error) {
	return accessdomain.Role{}, nil
}

func (s *storeSpy) UpdateRole(context.Context, accessdomain.Role) (accessdomain.Role, error) {
	return accessdomain.Role{}, nil
}

func (s *storeSpy) FindMenuByID(context.Context, int64) (accessdomain.Menu, error) {
	return accessdomain.RestoreMenu(1, 0, "Existing", "/existing", "menu", accessdomain.PermissionLogRead, 10, true, s.menuCreatedAt, s.menuCreatedAt)
}

func (s *storeSpy) ListMenus(context.Context) ([]accessdomain.Menu, error) {
	return nil, nil
}

func (s *storeSpy) CreateMenu(context.Context, accessdomain.Menu) (accessdomain.Menu, error) {
	return accessdomain.Menu{}, nil
}

func (s *storeSpy) UpdateMenu(_ context.Context, menu accessdomain.Menu) (accessdomain.Menu, error) {
	s.updatedMenu = menu
	return menu, nil
}
