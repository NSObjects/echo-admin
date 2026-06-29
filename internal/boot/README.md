# Boot

`internal/boot` is the composition root. It owns process startup, configured infrastructure resources, `samber/do` dependency injection, and business route mounting.

Business modules declare what they provide and what they mount:

```go
func accessModule() Module {
	return NewModule("access",
		Provide(newAccessStore),
		Provide(newAccessUsecase),
		Provide(newAccessHandler),
		Route(accesshttp.Register),
	)
}
```

Provider functions are ordinary `do.Provider[T]` functions. Runtime business storage uses the configured MySQL resource:

```go
func newAccessStore(i do.Injector) (*accessmysql.Store, error) {
	ctx, db, err := startupMySQL(i)
	if err != nil {
		return nil, err
	}
	return accessmysql.NewStore(ctx, db)
}
```

Business code lives under `internal/modules/<module>`. Platform runtime code lives under `internal/platform`. Boot is allowed to import adapters, infrastructure, server, and configs so usecase and domain packages stay clean.

Cross-module wiring belongs here. For example, auth defines the small reader, recorder, and login-session interfaces it needs, while boot passes the concrete identity/access/auth MySQL stores and bridges login records into the audit usecase. Access also receives a small identity reader so role delegation can use the current administrator’s assigned roles and active role without importing the identity adapter directly. Settings receives a version-catalog bridge here so version bundles can export/import access-owned menus and APIs without making settings import the access adapter. API token authentication is wired here through a small server-facing verifier adapter, so `internal/platform/server` can accept `X-API-Token` without importing token storage. Browser Login Session authentication is wired the same way: server receives only an opaque-token authenticator and does not import auth storage. Internal-error recording is wired through a small recorder interface, with boot bridging it to the audit usecase. The default administrator seed is also wired here: boot waits for the access store to seed the `super_admin` role, then seeds the identity store with the local `admin` account assigned to that role. Business modules should not import each other’s adapters directly.

Authorization uses Casbin RBAC inside the auth usecase. Boot wires auth to identity for users, access for roles, data-role visibility, gin-vue-admin-style menus, menu buttons, and managed API routes, audit for login records, and auth-owned Login Session storage. The auth module builds a Casbin `{subject, object, action}` policy snapshot from the current active role, then private HTTP handlers pass the registered method/path route into auth so ordinary roles must hold both the permission token and the matching API id. Access seeds the root role with every boot-time menu id, button id, API id, and existing role id for data authority, including API Token, current-user profile, system parameter, login-session logout routes, system-error resolution, and version import/export routes, before identity seeds the default `admin` account.
