package boot

import (
	"context"
	"strconv"

	"github.com/samber/do/v2"

	accessusecase "github.com/NSObjects/echo-admin/internal/modules/access/usecase"
	authusecase "github.com/NSObjects/echo-admin/internal/modules/auth/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
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

type routeAuthorizerAdapter struct {
	access *accessusecase.Usecase
}

type authAuthorizationReader struct {
	access *accessusecase.Usecase
}

func newRouteAuthorizer(i do.Injector) (routeAuthorizer, error) {
	access, err := do.Invoke[*accessusecase.Usecase](i)
	if err != nil {
		return nil, err
	}
	return routeAuthorizerAdapter{access: access}, nil
}

func (a routeAuthorizerAdapter) AuthorizeRoute(ctx context.Context, method, path string) error {
	if a.access == nil {
		return apperr.New(apperr.ErrInternalServer, "route authorizer is not configured")
	}
	adminID, err := strconv.ParseInt(requestctx.GetUserID(ctx), 10, 64)
	if err != nil || adminID <= 0 {
		return apperr.NewUnauthorized()
	}
	roleID, err := strconv.ParseInt(requestctx.GetRoleID(ctx), 10, 64)
	if err != nil || roleID <= 0 {
		return apperr.NewUnauthorized()
	}
	return a.access.AuthorizeRoute(ctx, accessusecase.AuthorizationSubject{
		AdminID:      adminID,
		ActiveRoleID: roleID,
	}, method, path)
}

func (r authAuthorizationReader) CurrentAuthorization(ctx context.Context, subject authusecase.AuthorizationSubject) (authusecase.AuthorizationView, error) {
	view, err := r.access.CurrentAuthorization(ctx, accessusecase.AuthorizationSubject{
		AdminID:      subject.AdminID,
		ActiveRoleID: subject.ActiveRoleID,
	})
	if err != nil {
		return authusecase.AuthorizationView{}, err
	}
	return authusecase.AuthorizationView{
		ActiveRole:  authRole(view.ActiveRole),
		Roles:       authRoles(view.Roles),
		Permissions: view.Permissions,
		Menus:       authMenus(view.Menus),
		DefaultPath: view.DefaultPath,
	}, nil
}

func authRoles(roles []accessusecase.Role) []authusecase.Role {
	out := make([]authusecase.Role, 0, len(roles))
	for _, role := range roles {
		out = append(out, authRole(role))
	}
	return out
}

func authRole(role accessusecase.Role) authusecase.Role {
	return authusecase.Role{
		ID:          role.ID,
		ParentID:    role.ParentID,
		Code:        role.Code,
		Name:        role.Name,
		Permissions: role.Permissions,
		MenuIDs:     role.MenuIDs,
		APIIDs:      role.APIIDs,
		ButtonIDs:   role.ButtonIDs,
		DataRoleIDs: role.DataRoleIDs,
		DefaultPath: role.DefaultPath,
		Active:      role.Active,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func authMenus(menus []accessusecase.Menu) []authusecase.Menu {
	out := make([]authusecase.Menu, 0, len(menus))
	for _, menu := range menus {
		buttons := make([]authusecase.Button, 0, len(menu.Buttons))
		for _, button := range menu.Buttons {
			buttons = append(buttons, authusecase.Button{
				ID:          button.ID,
				MenuID:      button.MenuID,
				Name:        button.Name,
				Description: button.Description,
				CreatedAt:   button.CreatedAt,
				UpdatedAt:   button.UpdatedAt,
			})
		}
		out = append(out, authusecase.Menu{
			ID:        menu.ID,
			ParentID:  menu.ParentID,
			Name:      menu.Name,
			Path:      menu.Path,
			Icon:      menu.Icon,
			Hidden:    menu.Hidden,
			Component: menu.Component,
			Meta: authusecase.MenuMeta{
				ActiveName:     menu.Meta.ActiveName,
				KeepAlive:      menu.Meta.KeepAlive,
				DefaultMenu:    menu.Meta.DefaultMenu,
				CloseTab:       menu.Meta.CloseTab,
				TransitionType: menu.Meta.TransitionType,
			},
			Permission: menu.Permission,
			Sort:       menu.Sort,
			Active:     menu.Active,
			Buttons:    buttons,
			CreatedAt:  menu.CreatedAt,
			UpdatedAt:  menu.UpdatedAt,
		})
	}
	return out
}
