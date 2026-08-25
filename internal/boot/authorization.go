package boot

import (
	"context"

	"github.com/samber/do/v2"
)

type routeAuthorizer interface {
	AuthorizeRoute(context.Context, string, string) error
}

func resolveRouteAuthorizer(injector do.Injector) (routeAuthorizer, error) {
	authorizer, err := do.InvokeAs[routeAuthorizer](injector)
	if err != nil {
		if optionalServiceMissing(err) {
			return nil, nil
		}
		return nil, err
	}
	return authorizer, nil
}
