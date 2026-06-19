package boot

import (
	"context"

	"github.com/samber/do/v2"
	"gorm.io/gorm"

	accessmysql "github.com/NSObjects/echo-admin/internal/modules/access/adapters/mysql"
	accesshttp "github.com/NSObjects/echo-admin/internal/modules/access/http"
	accessusecase "github.com/NSObjects/echo-admin/internal/modules/access/usecase"
	auditmysql "github.com/NSObjects/echo-admin/internal/modules/audit/adapters/mysql"
	audithttp "github.com/NSObjects/echo-admin/internal/modules/audit/http"
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	authhttp "github.com/NSObjects/echo-admin/internal/modules/auth/http"
	authusecase "github.com/NSObjects/echo-admin/internal/modules/auth/usecase"
	filemysql "github.com/NSObjects/echo-admin/internal/modules/fileasset/adapters/mysql"
	filehttp "github.com/NSObjects/echo-admin/internal/modules/fileasset/http"
	fileusecase "github.com/NSObjects/echo-admin/internal/modules/fileasset/usecase"
	identitymysql "github.com/NSObjects/echo-admin/internal/modules/identity/adapters/mysql"
	identityhttp "github.com/NSObjects/echo-admin/internal/modules/identity/http"
	identityusecase "github.com/NSObjects/echo-admin/internal/modules/identity/usecase"
	settingsmysql "github.com/NSObjects/echo-admin/internal/modules/settings/adapters/mysql"
	settingshttp "github.com/NSObjects/echo-admin/internal/modules/settings/http"
	settingsusecase "github.com/NSObjects/echo-admin/internal/modules/settings/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/configs"
)

// BusinessModules returns the business modules installed by the default runtime.
func BusinessModules() []Module {
	return []Module{
		accessModule(),
		identityModule(),
		auditModule(),
		authModule(),
		settingsModule(),
		fileModule(),
	}
}

func accessModule() Module {
	return NewModule("access",
		Provide(newAccessStore),
		Provide(newAccessUsecase),
		Provide(newAccessHandler),
		Route(accesshttp.Register),
	)
}

func identityModule() Module {
	return NewModule("identity",
		Provide(newIdentityStore),
		Provide(newIdentityUsecase),
		Provide(newIdentityHandler),
		Route(identityhttp.Register),
	)
}

func auditModule() Module {
	return NewModule("audit",
		Provide(newAuditStore),
		Provide(newAuditUsecase),
		Provide(newAuditHandler),
		Route(audithttp.Register),
	)
}

func authModule() Module {
	return NewModule("auth",
		Provide(newAuthUsecase),
		Provide(newAuthHandler),
		Route(authhttp.Register),
	)
}

func settingsModule() Module {
	return NewModule("settings",
		Provide(newSettingsStore),
		Provide(newSettingsUsecase),
		Provide(newSettingsHandler),
		Route(settingshttp.Register),
	)
}

func fileModule() Module {
	return NewModule("fileasset",
		Provide(newFileStore),
		Provide(newFileUsecase),
		Provide(newFileHandler),
		Route(filehttp.Register),
	)
}

func newAccessStore(i do.Injector) (*accessmysql.Store, error) {
	ctx, db, err := startupMySQL(i)
	if err != nil {
		return nil, err
	}
	return accessmysql.NewStore(ctx, db)
}

func newAccessUsecase(i do.Injector) (*accessusecase.Usecase, error) {
	store, err := do.Invoke[*accessmysql.Store](i)
	if err != nil {
		return nil, err
	}
	return accessusecase.New(store), nil
}

func newAccessHandler(i do.Injector) (*accesshttp.Handler, error) {
	uc, auth, audit, err := accessHandlerDeps(i)
	if err != nil {
		return nil, err
	}
	return accesshttp.New(uc, auth, audit), nil
}

func newIdentityStore(i do.Injector) (*identitymysql.Store, error) {
	ctx, db, err := startupMySQL(i)
	if err != nil {
		return nil, err
	}
	return identitymysql.NewStore(ctx, db)
}

func newIdentityUsecase(i do.Injector) (*identityusecase.Usecase, error) {
	store, err := do.Invoke[*identitymysql.Store](i)
	if err != nil {
		return nil, err
	}
	return identityusecase.New(store), nil
}

func newIdentityHandler(i do.Injector) (*identityhttp.Handler, error) {
	uc, err := do.Invoke[*identityusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	auth, err := do.Invoke[*authusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return identityhttp.New(uc, auth, audit), nil
}

func newAuditStore(i do.Injector) (*auditmysql.Store, error) {
	ctx, db, err := startupMySQL(i)
	if err != nil {
		return nil, err
	}
	return auditmysql.NewStore(ctx, db)
}

func newAuditUsecase(i do.Injector) (*auditusecase.Usecase, error) {
	store, err := do.Invoke[*auditmysql.Store](i)
	if err != nil {
		return nil, err
	}
	return auditusecase.New(store), nil
}

func newAuditHandler(i do.Injector) (*audithttp.Handler, error) {
	uc, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	auth, err := do.Invoke[*authusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return audithttp.New(uc, auth), nil
}

func newAuthUsecase(i do.Injector) (*authusecase.Usecase, error) {
	ctx, err := do.Invoke[context.Context](i)
	if err != nil {
		return nil, err
	}
	identityStore, err := do.Invoke[*identitymysql.Store](i)
	if err != nil {
		return nil, err
	}
	accessStore, err := do.Invoke[*accessmysql.Store](i)
	if err != nil {
		return nil, err
	}
	role, err := accessStore.FindRoleByCode(ctx, "super_admin")
	if err != nil {
		return nil, err
	}
	if seedErr := identityStore.SeedDefaultAdmin(ctx, role.ID()); seedErr != nil {
		return nil, seedErr
	}
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	cfg, err := do.Invoke[configs.Config](i)
	if err != nil {
		return nil, err
	}
	return authusecase.New(identityStore, accessStore, accessStore, authLoginRecorder{audit: audit}, cfg.JWT.Secret), nil
}

func newAuthHandler(i do.Injector) (*authhttp.Handler, error) {
	uc, err := do.Invoke[*authusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return authhttp.New(uc), nil
}

func newSettingsStore(i do.Injector) (*settingsmysql.Store, error) {
	ctx, db, err := startupMySQL(i)
	if err != nil {
		return nil, err
	}
	return settingsmysql.NewStore(ctx, db)
}

func newSettingsUsecase(i do.Injector) (*settingsusecase.Usecase, error) {
	store, err := do.Invoke[*settingsmysql.Store](i)
	if err != nil {
		return nil, err
	}
	return settingsusecase.New(store), nil
}

func newSettingsHandler(i do.Injector) (*settingshttp.Handler, error) {
	uc, err := do.Invoke[*settingsusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	auth, err := do.Invoke[*authusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return settingshttp.New(uc, auth, audit), nil
}

func newFileStore(i do.Injector) (*filemysql.Store, error) {
	ctx, db, err := startupMySQL(i)
	if err != nil {
		return nil, err
	}
	return filemysql.NewStore(ctx, db)
}

func newFileUsecase(i do.Injector) (*fileusecase.Usecase, error) {
	store, err := do.Invoke[*filemysql.Store](i)
	if err != nil {
		return nil, err
	}
	return fileusecase.New(store), nil
}

func newFileHandler(i do.Injector) (*filehttp.Handler, error) {
	uc, err := do.Invoke[*fileusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	auth, err := do.Invoke[*authusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	cfg, err := do.Invoke[configs.Config](i)
	if err != nil {
		return nil, err
	}
	return filehttp.New(uc, auth, audit, cfg.Admin.UploadDir), nil
}

func startupMySQL(i do.Injector) (context.Context, *gorm.DB, error) {
	ctx, err := do.Invoke[context.Context](i)
	if err != nil {
		return nil, nil, err
	}
	db, err := do.Invoke[*gorm.DB](i)
	if err != nil {
		return nil, nil, err
	}
	return ctx, db, nil
}

func accessHandlerDeps(i do.Injector) (*accessusecase.Usecase, *authusecase.Usecase, *auditusecase.Usecase, error) {
	uc, err := do.Invoke[*accessusecase.Usecase](i)
	if err != nil {
		return nil, nil, nil, err
	}
	auth, err := do.Invoke[*authusecase.Usecase](i)
	if err != nil {
		return nil, nil, nil, err
	}
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, nil, nil, err
	}
	return uc, auth, audit, nil
}

type authLoginRecorder struct {
	audit *auditusecase.Usecase
}

func (r authLoginRecorder) RecordLogin(ctx context.Context, record authusecase.LoginRecord) error {
	_, err := r.audit.RecordLogin(ctx, auditusecase.LoginInput{
		AdminID:   record.AdminID,
		Username:  record.Username,
		IP:        record.IP,
		UserAgent: record.UserAgent,
		Success:   record.Success,
		Reason:    record.Reason,
	})
	return err
}
