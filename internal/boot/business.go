package boot

import (
	"context"
	"strconv"
	"time"

	"github.com/samber/do/v2"
	"gorm.io/gorm"

	accessmysql "github.com/NSObjects/echo-admin/internal/modules/access/adapters/mysql"
	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	accesshttp "github.com/NSObjects/echo-admin/internal/modules/access/http"
	accessusecase "github.com/NSObjects/echo-admin/internal/modules/access/usecase"
	apitokenmysql "github.com/NSObjects/echo-admin/internal/modules/apitoken/adapters/mysql"
	apitokenhttp "github.com/NSObjects/echo-admin/internal/modules/apitoken/http"
	apitokenusecase "github.com/NSObjects/echo-admin/internal/modules/apitoken/usecase"
	auditmysql "github.com/NSObjects/echo-admin/internal/modules/audit/adapters/mysql"
	audithttp "github.com/NSObjects/echo-admin/internal/modules/audit/http"
	"github.com/NSObjects/echo-admin/internal/modules/audit/oprec"
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	authmysql "github.com/NSObjects/echo-admin/internal/modules/auth/adapters/mysql"
	authhttp "github.com/NSObjects/echo-admin/internal/modules/auth/http"
	authusecase "github.com/NSObjects/echo-admin/internal/modules/auth/usecase"
	filemysql "github.com/NSObjects/echo-admin/internal/modules/fileasset/adapters/mysql"
	filehttp "github.com/NSObjects/echo-admin/internal/modules/fileasset/http"
	fileusecase "github.com/NSObjects/echo-admin/internal/modules/fileasset/usecase"
	identitymysql "github.com/NSObjects/echo-admin/internal/modules/identity/adapters/mysql"
	identitydomain "github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	identityhttp "github.com/NSObjects/echo-admin/internal/modules/identity/http"
	identityusecase "github.com/NSObjects/echo-admin/internal/modules/identity/usecase"
	settingsmysql "github.com/NSObjects/echo-admin/internal/modules/settings/adapters/mysql"
	settingsdomain "github.com/NSObjects/echo-admin/internal/modules/settings/domain"
	settingshttp "github.com/NSObjects/echo-admin/internal/modules/settings/http"
	settingsusecase "github.com/NSObjects/echo-admin/internal/modules/settings/usecase"
	setupmysql "github.com/NSObjects/echo-admin/internal/modules/setup/adapters/mysql"
	setuphttp "github.com/NSObjects/echo-admin/internal/modules/setup/http"
	setupusecase "github.com/NSObjects/echo-admin/internal/modules/setup/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/configs"
	"github.com/NSObjects/echo-admin/internal/platform/server"
)

// BusinessModules returns the business modules installed by the default runtime.
func BusinessModules() []Module {
	return []Module{
		setupModule(),
		accessModule(),
		identityModule(),
		auditModule(),
		apiTokenModule(),
		authModule(),
		settingsModule(),
		fileModule(),
	}
}

func setupModule() Module {
	return NewModule("setup",
		Provide(newSetupStore),
		Provide(newSetupTransactionRunner),
		Provide(newSetupUsecase),
		Provide(newInstallationStateReader),
		Provide(newSetupHandler),
		Route(setuphttp.Register),
	)
}

func accessModule() Module {
	return NewModule("access",
		Provide(newAccessStore),
		Provide(newAccessUsecase),
		Provide(newAccessHandler),
		Route(accesshttp.Register),
	)
}

func apiTokenModule() Module {
	return NewModule("apitoken",
		Provide(newAPITokenStore),
		Provide(newAPITokenUsecase),
		Provide(newAPIKeyVerifier),
		Provide(newAPITokenHandler),
		Route(apitokenhttp.Register),
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
		Provide(newSystemErrorRecorder),
		Provide(newAuditHandler),
		Route(audithttp.Register),
	)
}

func authModule() Module {
	return NewModule("auth",
		Provide(newAuthStore),
		Provide(newAuthUsecase),
		Provide(newLoginSessionAuthenticator),
		Provide(newAuthHandler),
		Route(authhttp.Register),
	)
}

func settingsModule() Module {
	return NewModule("settings",
		Provide(newSettingsStore),
		Provide(newSettingsImportRunner),
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

func newSetupStore(i do.Injector) (*setupmysql.Store, error) {
	ctx, db, err := startupMySQL(i)
	if err != nil {
		return nil, err
	}
	return setupmysql.NewStore(ctx, db)
}

func newSetupTransactionRunner(i do.Injector) (setupusecase.TransactionRunner, error) {
	_, db, err := startupMySQL(i)
	if err != nil {
		return nil, err
	}
	setupStore, err := do.Invoke[*setupmysql.Store](i)
	if err != nil {
		return nil, err
	}
	accessStore, err := do.Invoke[*accessmysql.Store](i)
	if err != nil {
		return nil, err
	}
	identityStore, err := do.Invoke[*identitymysql.Store](i)
	if err != nil {
		return nil, err
	}
	settingsStore, err := do.Invoke[*settingsmysql.Store](i)
	if err != nil {
		return nil, err
	}
	return setupTransactionRunner{
		db:       db,
		setup:    setupStore,
		access:   accessStore,
		identity: identityStore,
		settings: settingsStore,
	}, nil
}

func newSetupUsecase(i do.Injector) (*setupusecase.Usecase, error) {
	store, err := do.Invoke[*setupmysql.Store](i)
	if err != nil {
		return nil, err
	}
	runner, err := do.Invoke[setupusecase.TransactionRunner](i)
	if err != nil {
		return nil, err
	}
	return setupusecase.New(store, runner), nil
}

func newInstallationStateReader(i do.Injector) (server.InstallationStateReader, error) {
	uc, err := do.Invoke[*setupusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return installationStateReader{setup: uc}, nil
}

func newSetupHandler(i do.Injector) (*setuphttp.Handler, error) {
	uc, err := do.Invoke[*setupusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return setuphttp.New(uc), nil
}

func newAccessUsecase(i do.Injector) (*accessusecase.Usecase, error) {
	store, err := do.Invoke[*accessmysql.Store](i)
	if err != nil {
		return nil, err
	}
	identityStore, err := do.Invoke[*identitymysql.Store](i)
	if err != nil {
		return nil, err
	}
	return accessusecase.New(store, accessAdminRoleReader{store: identityStore}), nil
}

func newAccessHandler(i do.Injector) (*accesshttp.Handler, error) {
	uc, audit, err := accessHandlerDeps(i)
	if err != nil {
		return nil, err
	}
	return accesshttp.New(uc, oprec.New(audit)), nil
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
	roles, err := do.Invoke[*accessusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	auth, err := do.Invoke[*authusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return identityusecase.New(store, roles, identitySessionRevoker{auth: auth}), nil
}

func newIdentityHandler(i do.Injector) (*identityhttp.Handler, error) {
	uc, err := do.Invoke[*identityusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return identityhttp.New(uc, oprec.New(audit)), nil
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

func newSystemErrorRecorder(i do.Injector) (server.SystemErrorRecorder, error) {
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return systemErrorRecorder{audit: audit}, nil
}

func newAuditHandler(i do.Injector) (*audithttp.Handler, error) {
	uc, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return audithttp.New(uc), nil
}

func newAPITokenStore(i do.Injector) (*apitokenmysql.Store, error) {
	ctx, db, err := startupMySQL(i)
	if err != nil {
		return nil, err
	}
	return apitokenmysql.NewStore(ctx, db)
}

func newAPITokenUsecase(i do.Injector) (*apitokenusecase.Usecase, error) {
	store, err := do.Invoke[*apitokenmysql.Store](i)
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
	return apitokenusecase.New(
		store,
		apiTokenAdminReader{store: identityStore},
		apiTokenRolePolicy{store: accessStore},
	), nil
}

func newAPIKeyVerifier(i do.Injector) (server.APIKeyVerifier, error) {
	uc, err := do.Invoke[*apitokenusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return apiKeyVerifier{tokens: uc}, nil
}

func newAPITokenHandler(i do.Injector) (*apitokenhttp.Handler, error) {
	uc, err := do.Invoke[*apitokenusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return apitokenhttp.New(uc, oprec.New(audit)), nil
}

func newAuthUsecase(i do.Injector) (*authusecase.Usecase, error) {
	authStore, err := do.Invoke[*authmysql.Store](i)
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
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return authusecase.New(identityStore, accessStore, accessStore, accessStore, authStore, authStore, authLoginRecorder{audit: audit}), nil
}

func newAuthStore(i do.Injector) (*authmysql.Store, error) {
	ctx, db, err := startupMySQL(i)
	if err != nil {
		return nil, err
	}
	return authmysql.NewStore(ctx, db)
}

func newLoginSessionAuthenticator(i do.Injector) (server.LoginSessionAuthenticator, error) {
	uc, err := do.Invoke[*authusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return loginSessionAuthenticator{auth: uc}, nil
}

type loginSessionAuthenticator struct {
	auth *authusecase.Usecase
}

func (a loginSessionAuthenticator) AuthenticateLoginSession(ctx context.Context, token string) (server.LoginSessionIdentity, error) {
	identity, err := a.auth.AuthenticateLoginSession(ctx, token)
	if err != nil {
		return server.LoginSessionIdentity{}, err
	}
	return server.LoginSessionIdentity{
		SessionID: formatID(identity.SessionID),
		UserID:    formatID(identity.AdminID),
		RoleID:    formatID(identity.RoleID),
	}, nil
}

type identitySessionRevoker struct {
	auth *authusecase.Usecase
}

func (r identitySessionRevoker) RevokeLoginSessions(ctx context.Context, adminID int64) error {
	return r.auth.RevokeLoginSessions(ctx, adminID)
}

type apiKeyVerifier struct {
	tokens *apitokenusecase.Usecase
}

func (v apiKeyVerifier) VerifyAPIKey(ctx context.Context, secret string) (server.APIKeyIdentity, error) {
	identity, err := v.tokens.Authenticate(ctx, secret)
	if err != nil {
		return server.APIKeyIdentity{}, err
	}
	return server.APIKeyIdentity{
		UserID: formatID(identity.AdminID),
		RoleID: formatID(identity.RoleID),
	}, nil
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

type systemErrorRecorder struct {
	audit *auditusecase.Usecase
}

func (r systemErrorRecorder) RecordSystemError(ctx context.Context, input server.SystemErrorInput) error {
	_, err := r.audit.RecordSystemError(ctx, auditusecase.SystemErrorInput{
		Code:      input.Code,
		Message:   input.Message,
		Detail:    input.Detail,
		Method:    input.Method,
		Path:      input.Path,
		IP:        input.IP,
		UserAgent: input.UserAgent,
		RequestID: input.RequestID,
		UserID:    input.UserID,
	})
	return err
}

func newAuthHandler(i do.Injector) (*authhttp.Handler, error) {
	uc, err := do.Invoke[*authusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	cfg, err := do.Invoke[configs.Config](i)
	if err != nil {
		return nil, err
	}
	return authhttp.New(uc, cfg.HTTP.SecureCookies), nil
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
	accessUC, err := do.Invoke[*accessusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	importRunner, err := do.Invoke[settingsusecase.ImportRunner](i)
	if err != nil {
		return nil, err
	}
	return settingsusecase.New(store,
		settingsusecase.WithVersionCatalog(settingsVersionCatalog{
			access: accessUC,
		}),
		settingsusecase.WithImportRunner(importRunner),
	), nil
}

func newSettingsImportRunner(i do.Injector) (settingsusecase.ImportRunner, error) {
	_, db, err := startupMySQL(i)
	if err != nil {
		return nil, err
	}
	settingsStore, err := do.Invoke[*settingsmysql.Store](i)
	if err != nil {
		return nil, err
	}
	accessStore, err := do.Invoke[*accessmysql.Store](i)
	if err != nil {
		return nil, err
	}
	accessUC, err := do.Invoke[*accessusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return settingsImportRunner{
		db:       db,
		settings: settingsStore,
		access:   accessStore,
		accessUC: accessUC,
	}, nil
}

func newSettingsHandler(i do.Injector) (*settingshttp.Handler, error) {
	uc, err := do.Invoke[*settingsusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return settingshttp.New(uc, oprec.New(audit)), nil
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
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	cfg, err := do.Invoke[configs.Config](i)
	if err != nil {
		return nil, err
	}
	return filehttp.New(uc, oprec.New(audit), cfg.Admin.UploadDir), nil
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

func accessHandlerDeps(i do.Injector) (*accessusecase.Usecase, *auditusecase.Usecase, error) {
	uc, err := do.Invoke[*accessusecase.Usecase](i)
	if err != nil {
		return nil, nil, err
	}
	audit, err := do.Invoke[*auditusecase.Usecase](i)
	if err != nil {
		return nil, nil, err
	}
	return uc, audit, nil
}

type authLoginRecorder struct {
	audit *auditusecase.Usecase
}

type installationStateReader struct {
	setup *setupusecase.Usecase
}

func (r installationStateReader) Initialized(ctx context.Context) (bool, error) {
	state, err := r.setup.State(ctx)
	if err != nil {
		return false, err
	}
	return state.Initialized, nil
}

type setupTransactionRunner struct {
	db       *gorm.DB
	setup    *setupmysql.Store
	access   *accessmysql.Store
	identity *identitymysql.Store
	settings *settingsmysql.Store
}

func (r setupTransactionRunner) RunInitialization(ctx context.Context, fn func(context.Context, setupusecase.Transaction) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transaction := setupTransaction{
			setup:    r.setup.WithDB(tx),
			access:   r.access.WithDB(tx),
			identity: r.identity.WithDB(tx),
			settings: r.settings.WithDB(tx),
		}
		return fn(ctx, transaction)
	})
}

type setupTransaction struct {
	setup    *setupmysql.Store
	access   *accessmysql.Store
	identity *identitymysql.Store
	settings *settingsmysql.Store
}

func (t setupTransaction) RequireOpenInstallation(ctx context.Context) error {
	return t.setup.RequireOpenInstallation(ctx)
}

func (t setupTransaction) InstallRootAuthorization(ctx context.Context) (setupusecase.RootRole, error) {
	role, err := t.access.InstallRootAuthorization(ctx)
	if err != nil {
		return setupusecase.RootRole{}, err
	}
	return setupusecase.RootRole{ID: role.ID, Code: role.Code}, nil
}

func (t setupTransaction) CreateFirstAdministrator(ctx context.Context, input setupusecase.FirstAdministrator) error {
	admin, err := identitydomain.RestoreAdmin(
		0,
		input.Username,
		input.DisplayName,
		input.Email,
		input.PasswordHash,
		[]int64{input.RootRoleID},
		input.RootRoleID,
		true,
		time.Time{},
		time.Time{},
	)
	if err != nil {
		return err
	}
	_, err = t.identity.Create(ctx, admin)
	return err
}

func (t setupTransaction) InstallInitialSettings(ctx context.Context, input setupusecase.InitialSettings) error {
	return t.settings.InstallInitialSettings(ctx, input.SiteName)
}

func (t setupTransaction) CompleteInstallation(ctx context.Context) error {
	return t.setup.CompleteInstallation(ctx)
}

type accessAdminRoleReader struct {
	store *identitymysql.Store
}

type apiTokenAdminReader struct {
	store *identitymysql.Store
}

func (r apiTokenAdminReader) AdminSnapshot(ctx context.Context, adminID int64) (apitokenusecase.AdminSnapshot, error) {
	admin, err := r.store.FindByID(ctx, adminID)
	if err != nil {
		return apitokenusecase.AdminSnapshot{}, err
	}
	return apitokenusecase.AdminSnapshot{
		RoleIDs: admin.RoleIDs,
		Active:  admin.Active,
	}, nil
}

type apiTokenRolePolicy struct {
	store *accessmysql.Store
}

type settingsVersionCatalog struct {
	access *accessusecase.Usecase
}

// ExportVersionMenus implements settingsusecase.VersionCatalog for access menus.
// The bridge only maps types; export rules live in the access usecase.
func (c settingsVersionCatalog) ExportVersionMenus(ctx context.Context, ids []int64) ([]settingsusecase.VersionMenu, error) {
	nodes, err := c.access.ExportMenuTree(ctx, ids)
	if err != nil {
		return nil, err
	}
	return versionMenuNodes(nodes), nil
}

// settingsImportRunner runs settings imports in one database transaction. Menu
// writes go through a transaction-scoped copy of the access usecase so import
// rules stay in the owning usecase.
type settingsImportRunner struct {
	db       *gorm.DB
	settings *settingsmysql.Store
	access   *accessmysql.Store
	accessUC *accessusecase.Usecase
}

// RunImport implements settingsusecase.ImportRunner.
func (r settingsImportRunner) RunImport(ctx context.Context, fn func(context.Context, settingsusecase.ImportTransaction) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transaction := settingsImportTransaction{
			access:   r.accessUC.WithStore(r.access.WithDB(tx)),
			settings: r.settings.WithDB(tx),
		}
		return fn(ctx, transaction)
	})
}

// settingsImportTransaction implements settingsusecase.ImportTransaction. The
// bridge only binds transaction-scoped stores; import rules live in the
// owning usecases.
type settingsImportTransaction struct {
	access   *accessusecase.Usecase
	settings *settingsmysql.Store
}

// ImportMenus implements settingsusecase.ImportTransaction for access menus.
func (t settingsImportTransaction) ImportMenus(ctx context.Context, menus []settingsusecase.VersionMenu) error {
	return t.access.ImportMenuTree(ctx, menuTreeInputs(menus))
}

// ReplaceDictionary implements settingsusecase.ImportTransaction for dictionaries.
func (t settingsImportTransaction) ReplaceDictionary(ctx context.Context, dictionary settingsdomain.Dictionary) error {
	_, err := t.settings.ReplaceDictionary(ctx, dictionary)
	return err
}

// CreateVersion implements settingsusecase.ImportTransaction for release records.
func (t settingsImportTransaction) CreateVersion(ctx context.Context, version settingsdomain.SystemVersion) (settingsdomain.SystemVersion, error) {
	return t.settings.CreateVersion(ctx, version)
}

func versionMenuNodes(nodes []accessusecase.MenuNode) []settingsusecase.VersionMenu {
	out := make([]settingsusecase.VersionMenu, 0, len(nodes))
	for _, node := range nodes {
		item := versionMenuFromDomain(node.Menu)
		item.Children = versionMenuNodes(node.Children)
		out = append(out, item)
	}
	return out
}

func menuTreeInputs(menus []settingsusecase.VersionMenu) []accessusecase.MenuTreeInput {
	out := make([]accessusecase.MenuTreeInput, 0, len(menus))
	for _, menu := range menus {
		out = append(out, accessusecase.MenuTreeInput{
			Name:       menu.Name,
			Path:       menu.Path,
			Icon:       menu.Icon,
			Hidden:     menu.Hidden,
			Component:  menu.Component,
			Meta:       accessMenuMetaInput(menu.Meta),
			Permission: menu.Permission,
			Sort:       menu.Sort,
			Active:     menu.Active,
			Buttons:    versionMenuButtons(menu.Buttons),
			Children:   menuTreeInputs(menu.Children),
		})
	}
	return out
}

func versionMenuFromDomain(menu accessdomain.Menu) settingsusecase.VersionMenu {
	return settingsusecase.VersionMenu{
		Name:       menu.Name,
		Path:       menu.Path,
		Icon:       menu.Icon,
		Hidden:     menu.Hidden,
		Component:  menu.Component,
		Meta:       versionMenuMetaFromDomain(menu.Meta),
		Permission: menu.Permission,
		Sort:       menu.Sort,
		Active:     menu.Active,
		Buttons:    versionButtonsFromDomain(menu.Buttons),
	}
}

func versionMenuMetaFromDomain(meta accessdomain.MenuMeta) settingsusecase.VersionMenuMeta {
	return settingsusecase.VersionMenuMeta{
		ActiveName:     meta.ActiveName,
		KeepAlive:      meta.KeepAlive,
		DefaultMenu:    meta.DefaultMenu,
		CloseTab:       meta.CloseTab,
		TransitionType: meta.TransitionType,
	}
}

func accessMenuMetaInput(meta settingsusecase.VersionMenuMeta) accessusecase.MenuMetaInput {
	return accessusecase.MenuMetaInput{
		ActiveName:     meta.ActiveName,
		KeepAlive:      meta.KeepAlive,
		DefaultMenu:    meta.DefaultMenu,
		CloseTab:       meta.CloseTab,
		TransitionType: meta.TransitionType,
	}
}

func versionButtonsFromDomain(buttons []accessdomain.MenuButton) []settingsusecase.VersionButton {
	out := make([]settingsusecase.VersionButton, 0, len(buttons))
	for _, button := range buttons {
		out = append(out, settingsusecase.VersionButton{
			Name:        button.Name,
			Description: button.Description,
		})
	}
	return out
}

func versionMenuButtons(buttons []settingsusecase.VersionButton) []accessusecase.MenuButtonInput {
	out := make([]accessusecase.MenuButtonInput, 0, len(buttons))
	for _, button := range buttons {
		out = append(out, accessusecase.MenuButtonInput{
			Name:        button.Name,
			Description: button.Description,
		})
	}
	return out
}

// RoleView implements apitokenusecase.RolePolicy for access roles. The bridge
// only projects role state; policy decisions live in the apitoken usecase.
func (p apiTokenRolePolicy) RoleView(ctx context.Context, roleID int64) (apitokenusecase.RoleView, error) {
	role, err := p.store.FindRoleByID(ctx, roleID)
	if err != nil {
		return apitokenusecase.RoleView{}, err
	}
	return apitokenusecase.RoleView{IsSuper: role.IsSuperAdmin()}, nil
}

func (r accessAdminRoleReader) AdminRoleState(ctx context.Context, adminID int64) (accessusecase.AdminRoleState, error) {
	admin, err := r.store.FindByID(ctx, adminID)
	if err != nil {
		return accessusecase.AdminRoleState{}, err
	}
	return accessusecase.AdminRoleState{
		RoleIDs:      admin.RoleIDs,
		ActiveRoleID: admin.ActiveRoleID,
	}, nil
}

func (r accessAdminRoleReader) RoleAssigned(ctx context.Context, roleID int64) (bool, error) {
	return r.store.RoleAssigned(ctx, roleID)
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
