package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/pagination"
)

// List returns paginated administrators.
func (u *Usecase) List(ctx context.Context, input ListInput) (ListOutput, error) {
	if err := u.ready(); err != nil {
		return ListOutput{}, err
	}
	filter, err := normalizeListInput(input)
	if err != nil {
		return ListOutput{}, err
	}
	admins, total, err := u.store.List(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}
	return ListOutput{
		Items:    mapAdmins(admins),
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Total:    total,
	}, nil
}

// Create validates and stores a new administrator.
func (u *Usecase) Create(ctx context.Context, input AdminInput) (Admin, error) {
	if err := u.ready(); err != nil {
		return Admin{}, err
	}
	hash, err := hashPassword(input.Password)
	if err != nil {
		return Admin{}, err
	}
	admin, err := domain.RestoreAdmin(0, input.Username, input.DisplayName, input.Email, hash, input.RoleIDs, input.Active, zeroTime(), zeroTime())
	if err != nil {
		return Admin{}, mapDomainError(err)
	}
	created, err := u.store.Create(ctx, admin)
	if err != nil {
		return Admin{}, err
	}
	return fromAdmin(created), nil
}

// Update applies mutable administrator changes.
func (u *Usecase) Update(ctx context.Context, input UpdateAdminInput) (Admin, error) {
	if err := u.ready(); err != nil {
		return Admin{}, err
	}
	existing, err := u.store.FindByID(ctx, input.ID)
	if err != nil {
		return Admin{}, err
	}
	updated, err := updateAdmin(existing, input)
	if err != nil {
		return Admin{}, err
	}
	saved, err := u.store.Update(ctx, updated)
	if err != nil {
		return Admin{}, err
	}
	return fromAdmin(saved), nil
}

func (u *Usecase) ready() error {
	if u == nil || u.store == nil {
		return apperr.New(apperr.ErrInternalServer, "identity store is not configured")
	}
	return nil
}

func updateAdmin(existing domain.Admin, input UpdateAdminInput) (domain.Admin, error) {
	displayName := existing.DisplayName()
	if input.DisplayName != nil {
		displayName = *input.DisplayName
	}
	email := existing.Email()
	if input.Email != nil {
		email = *input.Email
	}
	active := existing.Active()
	if input.Active != nil {
		active = *input.Active
	}
	roleIDs := coalesceIDs(input.RoleIDs, existing.RoleIDs())
	updated, err := existing.UpdateProfile(displayName, email, roleIDs, active)
	if err != nil {
		return domain.Admin{}, mapDomainError(err)
	}
	if input.Password == nil {
		return updated, nil
	}
	hash, err := hashPassword(*input.Password)
	if err != nil {
		return domain.Admin{}, err
	}
	return updated.ReplacePassword(hash)
}

func hashPassword(password string) ([]byte, error) {
	if len(password) < 8 || len(password) > 72 {
		return nil, apperr.NewBadRequest("invalid password")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	return hash, nil
}

func normalizeListInput(input ListInput) (ListFilter, error) {
	window, err := pagination.Normalize(input.Page, input.PageSize, pagination.Options{
		DefaultPageSize: defaultPageSize,
		MaxPageSize:     maxPageSize,
	})
	if err != nil {
		return ListFilter{}, apperr.NewBadRequest("invalid pagination")
	}
	return ListFilter{Offset: window.Offset, Limit: window.Limit, Page: window.Page, PageSize: window.PageSize}, nil
}

func coalesceIDs(next, fallback []int64) []int64 {
	if next == nil {
		return fallback
	}
	return next
}

func mapAdmins(admins []domain.Admin) []Admin {
	out := make([]Admin, 0, len(admins))
	for _, admin := range admins {
		out = append(out, fromAdmin(admin))
	}
	return out
}

func fromAdmin(admin domain.Admin) Admin {
	return Admin{
		ID:          admin.ID(),
		Username:    admin.Username(),
		DisplayName: admin.DisplayName(),
		Email:       admin.Email(),
		RoleIDs:     admin.RoleIDs(),
		Active:      admin.Active(),
		CreatedAt:   admin.CreatedAt(),
		UpdatedAt:   admin.UpdatedAt(),
	}
}

func mapDomainError(err error) error {
	for _, entry := range domainErrorMessages {
		if errors.Is(err, entry.err) {
			return apperr.NewBadRequest(entry.message)
		}
	}
	return err
}

func zeroTime() time.Time {
	return time.Time{}
}

var domainErrorMessages = []struct {
	err     error
	message string
}{
	{domain.ErrInvalidAdminID, "invalid admin id"},
	{domain.ErrInvalidUsername, "invalid username"},
	{domain.ErrInvalidDisplayName, "invalid display name"},
	{domain.ErrInvalidPasswordHash, "invalid password"},
	{domain.ErrAdminNeedsRole, "admin needs a role"},
}
