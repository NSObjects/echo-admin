package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/pagination"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
)

// List returns paginated administrators visible to the active role.
func (u *Usecase) List(ctx context.Context, input ListInput) (ListOutput, error) {
	if err := u.ready(); err != nil {
		return ListOutput{}, err
	}
	filter, err := normalizeListInput(input)
	if err != nil {
		return ListOutput{}, err
	}
	admins, err := u.store.ListAll(ctx)
	if err != nil {
		return ListOutput{}, err
	}
	assignable, err := u.assignableRoleSet(ctx)
	if err != nil {
		return ListOutput{}, err
	}
	scoped := filterAdminsByRoles(admins, assignable)
	pageAdmins := paginateAdmins(scoped, filter)
	return ListOutput{
		Items:    mapAdmins(pageAdmins),
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Total:    len(scoped),
	}, nil
}

// Create validates and stores a new administrator.
func (u *Usecase) Create(ctx context.Context, input AdminInput) (Admin, error) {
	if err := u.ready(); err != nil {
		return Admin{}, err
	}
	if err := u.roles.EnsureAssignableRoles(ctx, input.RoleIDs); err != nil {
		return Admin{}, err
	}
	hash, err := hashPassword(input.Password)
	if err != nil {
		return Admin{}, err
	}
	admin, err := domain.RestoreAdmin(0, input.Username, input.DisplayName, input.Email, hash, input.RoleIDs, input.ActiveRoleID, input.Active, zeroTime(), zeroTime())
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
	if checkErr := u.roles.EnsureAssignableRoles(ctx, existing.RoleIDs); checkErr != nil {
		return Admin{}, checkErr
	}
	if checkErr := rejectSelfDisable(ctx, existing.ID, input.Active); checkErr != nil {
		return Admin{}, checkErr
	}
	updated, err := u.updateAdmin(ctx, existing, input)
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
	if u == nil || u.store == nil || u.roles == nil {
		return apperr.New(apperr.ErrInternalServer, "identity dependencies are not configured")
	}
	return nil
}

func (u *Usecase) updateAdmin(ctx context.Context, existing domain.Admin, input UpdateAdminInput) (domain.Admin, error) {
	displayName := existing.DisplayName
	if input.DisplayName != nil {
		displayName = *input.DisplayName
	}
	email := existing.Email
	if input.Email != nil {
		email = *input.Email
	}
	active := existing.Active
	if input.Active != nil {
		active = *input.Active
	}
	roleIDs := coalesceIDs(input.RoleIDs, existing.RoleIDs)
	if err := u.roles.EnsureAssignableRoles(ctx, roleIDs); err != nil {
		return domain.Admin{}, err
	}
	activeRoleID := existing.ActiveRoleID
	if input.ActiveRoleID != nil {
		activeRoleID = *input.ActiveRoleID
	} else if !containsID(roleIDs, activeRoleID) {
		activeRoleID = 0
	}
	updated, err := existing.UpdateProfile(displayName, email, roleIDs, activeRoleID, active)
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
	updated, err = updated.ReplacePassword(hash)
	if err != nil {
		return domain.Admin{}, mapDomainError(err)
	}
	return updated, nil
}

func (u *Usecase) assignableRoleSet(ctx context.Context) (map[int64]struct{}, error) {
	ids, err := u.roles.AssignableRoleIDs(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		out[id] = struct{}{}
	}
	return out, nil
}

func rejectSelfDisable(ctx context.Context, adminID int64, active *bool) error {
	if active == nil || *active {
		return nil
	}
	currentID, err := currentAdminID(ctx)
	if err != nil {
		return err
	}
	if currentID == adminID {
		return apperr.NewBadRequest("cannot disable current admin")
	}
	return nil
}

func currentAdminID(ctx context.Context) (int64, error) {
	raw := requestctx.GetUserID(ctx)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperr.NewUnauthorized()
	}
	return id, nil
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

func paginateAdmins(admins []domain.Admin, filter ListFilter) []domain.Admin {
	if filter.Offset >= len(admins) {
		return []domain.Admin{}
	}
	end := filter.Offset + filter.Limit
	if end > len(admins) {
		end = len(admins)
	}
	return admins[filter.Offset:end]
}

func filterAdminsByRoles(admins []domain.Admin, allowed map[int64]struct{}) []domain.Admin {
	out := make([]domain.Admin, 0, len(admins))
	for _, admin := range admins {
		if allRolesAllowed(admin.RoleIDs, allowed) {
			out = append(out, admin)
		}
	}
	return out
}

func allRolesAllowed(roleIDs []int64, allowed map[int64]struct{}) bool {
	if len(roleIDs) == 0 {
		return false
	}
	for _, roleID := range roleIDs {
		if _, ok := allowed[roleID]; !ok {
			return false
		}
	}
	return true
}

func coalesceIDs(next, fallback []int64) []int64 {
	if next == nil {
		return fallback
	}
	return next
}

func containsID(ids []int64, want int64) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
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
		ID:           admin.ID,
		Username:     admin.Username,
		DisplayName:  admin.DisplayName,
		Email:        admin.Email,
		RoleIDs:      admin.RoleIDs,
		ActiveRoleID: admin.ActiveRoleID,
		Active:       admin.Active,
		CreatedAt:    admin.CreatedAt,
		UpdatedAt:    admin.UpdatedAt,
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
	{domain.ErrInvalidActiveRole, "active role is not assigned"},
}
