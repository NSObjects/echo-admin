package usecase

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/NSObjects/echo-admin/internal/modules/setup/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/infrastructure/logging"
)

// State returns whether setup has completed.
func (u *Usecase) State(ctx context.Context) (State, error) {
	if err := u.ready(); err != nil {
		return State{}, err
	}
	initialized, err := u.state.Initialized(ctx)
	if err != nil {
		return State{}, err
	}
	return stateFromDomain(domain.NewInstallationState(initialized)), nil
}

// Submit performs the one-time system first initialization.
func (u *Usecase) Submit(ctx context.Context, input SubmitInput) (State, error) {
	if err := u.ready(); err != nil {
		return State{}, err
	}
	normalized, err := normalizeSubmitInput(input)
	if err != nil {
		return State{}, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(normalized.Password), bcrypt.DefaultCost)
	if err != nil {
		return State{}, apperr.WrapInternal(err, "hash setup administrator password")
	}

	err = u.runner.RunInitialization(ctx, func(txCtx context.Context, tx Transaction) error {
		err = tx.RequireOpenInstallation(txCtx)
		if err != nil {
			return err
		}
		var rootRole RootRole
		rootRole, err = tx.InstallRootAuthorization(txCtx)
		if err != nil {
			return err
		}
		if rootRole.ID <= 0 {
			return apperr.WrapInternal(errors.New("root role id is empty"), "install root authorization")
		}
		admin := FirstAdministrator{
			Username:     normalized.Username,
			DisplayName:  normalized.DisplayName,
			Email:        normalized.Email,
			PasswordHash: passwordHash,
			RootRoleID:   rootRole.ID,
		}
		err = tx.CreateFirstAdministrator(txCtx, admin)
		if err != nil {
			return err
		}
		err = tx.InstallInitialSettings(txCtx, InitialSettings{SiteName: normalized.SiteName})
		if err != nil {
			return err
		}
		return tx.CompleteInstallation(txCtx)
	})
	if err != nil {
		return State{}, err
	}

	logging.FromContext(ctx).Info().Str("username", normalized.Username).Msg("system first initialization completed")
	return State{Initialized: true}, nil
}

func (u *Usecase) ready() error {
	if u == nil || u.state == nil || u.runner == nil {
		return apperr.WrapInternal(errors.New("setup usecase is not configured"), "setup usecase is not configured")
	}
	return nil
}

func normalizeSubmitInput(input SubmitInput) (SubmitInput, error) {
	input.Username = strings.ToLower(strings.TrimSpace(input.Username))
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Email = strings.TrimSpace(input.Email)
	input.SiteName = strings.TrimSpace(input.SiteName)
	if input.SiteName == "" {
		input.SiteName = DefaultSiteName
	}
	if err := validateSubmitUsername(input.Username); err != nil {
		return SubmitInput{}, err
	}
	if err := validateSubmitDisplayName(input.DisplayName); err != nil {
		return SubmitInput{}, err
	}
	if err := validateSubmitEmail(input.Email); err != nil {
		return SubmitInput{}, err
	}
	if err := validateSubmitPassword(input.Password); err != nil {
		return SubmitInput{}, err
	}
	if len(input.SiteName) > 120 {
		return SubmitInput{}, apperr.NewBadRequest("site_name is too long")
	}
	return input, nil
}

func validateSubmitUsername(username string) error {
	if username == "" {
		return apperr.NewBadRequest("username is required")
	}
	if len(username) > 64 {
		return apperr.NewBadRequest("username is too long")
	}
	return nil
}

func validateSubmitDisplayName(displayName string) error {
	if displayName == "" {
		return apperr.NewBadRequest("display_name is required")
	}
	if len(displayName) > 80 {
		return apperr.NewBadRequest("display_name is too long")
	}
	return nil
}

func validateSubmitEmail(email string) error {
	if len(email) > 160 {
		return apperr.NewBadRequest("email is too long")
	}
	if email == "" {
		return nil
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return apperr.NewBadRequest("email is invalid")
	}
	return nil
}

func validateSubmitPassword(password string) error {
	if len(password) < minPasswordLength {
		return apperr.NewBadRequest("password is too short")
	}
	if len(password) > maxPasswordLength {
		return apperr.NewBadRequest("password is too long")
	}
	return nil
}
