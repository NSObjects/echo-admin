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

Cross-module wiring belongs here. For example, auth defines the small reader and recorder interfaces it needs, while boot passes the concrete identity/access MySQL stores and bridges login records into the audit usecase. Business modules should not import each other’s adapters directly.

Authorization uses Casbin RBAC inside the auth usecase. Boot wires auth to identity for users, access for roles/menus, and audit for login records; the auth module then builds a Casbin `{subject, object, action}` policy snapshot from current MySQL-backed role assignments for each authorization check.
