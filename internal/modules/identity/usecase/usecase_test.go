package usecase_test

import (
	"context"
	"testing"
	"time"

	identitydomain "github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	"github.com/NSObjects/echo-admin/internal/modules/identity/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
)

func TestDeleteRejectsCurrentAdmin(t *testing.T) {
	store := &storeSpy{admins: map[int64]identitydomain.Admin{
		42: newAdmin(t, 42, "admin"),
	}}
	uc := usecase.New(store, roleScopeSpy{})

	err := uc.Delete(requestctx.WithUserID(context.Background(), "42"), 42)
	if err == nil {
		t.Fatal("Delete(current admin) error = nil, want bad request")
	}
	if store.deletedID != 0 {
		t.Fatalf("deletedID = %d, want 0", store.deletedID)
	}
}

func TestDeleteRemovesScopedAdmin(t *testing.T) {
	store := &storeSpy{admins: map[int64]identitydomain.Admin{
		7: newAdmin(t, 7, "operator"),
	}}
	uc := usecase.New(store, roleScopeSpy{})

	if err := uc.Delete(requestctx.WithUserID(context.Background(), "42"), 7); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if store.deletedID != 7 {
		t.Fatalf("deletedID = %d, want 7", store.deletedID)
	}
}

func newAdmin(t *testing.T, id int64, username string) identitydomain.Admin {
	t.Helper()
	now := time.Unix(1_800_000_000, 0).UTC()
	admin, err := identitydomain.RestoreAdmin(id, username, username, username+"@example.com", []byte("hash"), []int64{1}, 1, true, now, now)
	if err != nil {
		t.Fatalf("RestoreAdmin() error = %v", err)
	}
	return admin
}

type storeSpy struct {
	admins    map[int64]identitydomain.Admin
	deletedID int64
}

func (s *storeSpy) FindByUsername(context.Context, string) (identitydomain.Admin, error) {
	return identitydomain.Admin{}, apperr.NewNotFound("admin")
}

func (s *storeSpy) FindByID(_ context.Context, id int64) (identitydomain.Admin, error) {
	admin, ok := s.admins[id]
	if !ok {
		return identitydomain.Admin{}, apperr.NewNotFound("admin")
	}
	return admin, nil
}

func (s *storeSpy) ListAll(context.Context) ([]identitydomain.Admin, error) {
	admins := make([]identitydomain.Admin, 0, len(s.admins))
	for _, admin := range s.admins {
		admins = append(admins, admin)
	}
	return admins, nil
}

func (s *storeSpy) Create(_ context.Context, admin identitydomain.Admin) (identitydomain.Admin, error) {
	return admin, nil
}

func (s *storeSpy) Update(_ context.Context, admin identitydomain.Admin) (identitydomain.Admin, error) {
	return admin, nil
}

func (s *storeSpy) Delete(_ context.Context, id int64) error {
	s.deletedID = id
	return nil
}

type roleScopeSpy struct{}

func (roleScopeSpy) AssignableRoleIDs(context.Context) ([]int64, error) {
	return []int64{1}, nil
}

func (roleScopeSpy) EnsureAssignableRoles(context.Context, []int64) error {
	return nil
}
