/*
 *
 * casbin.go
 * middlewares
 *
 * Created by lintao on 2024/1/4 15:06
 * Copyright © 2020-2024 LINTAO. All rights reserved.
 *
 */

package middlewares

import (
	"context"

	"github.com/NSObjects/echo-admin/internal/api/data"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/code"
	"github.com/casbin/casbin/v2"
	"github.com/golang-jwt/jwt/v5"
	casbin_mw "github.com/labstack/echo-contrib/casbin"
	"github.com/labstack/echo/v4"
	"github.com/marmotedu/errors"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func Casbin(enforce *casbin.Enforcer) echo.MiddlewareFunc {
	return casbin_mw.MiddlewareWithConfig(casbin_mw.Config{
		Enforcer: enforce,
		Skipper: func(c echo.Context) bool {
			skipper := []string{
				"/api/users/current",
				"/api/login/account",
				"/api/user/menus",
				"/api/login/out",
				"/api/api"}
			return lo.Contains(skipper, c.Path())
		},
		ErrorHandler: func(c echo.Context, internal error, proposedStatus int) error {
			return errors.WrapC(internal, code.ErrPermissionDenied, "权限不足")
		},
		UserGetter: func(c echo.Context) (string, error) {
			token, ok := c.Get("user").(*jwt.Token)
			if !ok {
				return "", errors.WrapC(errors.New("token is nil"), code.ErrSignatureInvalid, "JWT签名无效")
			}
			if token == nil {
				return "", nil
			}

			user := token.Claims.(*data.JwtCustomClaims)
			if user == nil {
				return "", nil
			}
			if user.Admin {
				return "root", nil
			}

			return cast.ToString(user.ID), nil
		},
		EnforceHandler: func(c echo.Context, user string) (bool, error) {
			if user == "root" {
				return true, nil
			}

			first, err := query.User.WithContext(context.Background()).
				Preload(query.User.Role).
				Where(query.User.ID.Eq(cast.ToUint(user))).First()
			if err != nil {
				return false, err
			}
			for _, v := range first.Role {
				allow, err := enforce.Enforce(user, cast.ToString(v.ID), c.Request().URL.Path, c.Request().Method)
				if err != nil {
					return false, err
				}
				if allow {
					return true, nil
				}
			}

			return false, nil
		},
	})
}
