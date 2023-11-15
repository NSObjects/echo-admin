/*
 *
 * login.go
 * service
 *
 * Created by lintao on 2023/11/9 16:15
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package service

import (
	"github.com/NSObjects/echo-admin/internal/api/biz"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/resp"
	"github.com/labstack/echo/v4"
)

type loginController struct {
	l *biz.LoginHandler
}

func (l *loginController) RegisterRouter(s *echo.Group, middlewareFunc ...echo.MiddlewareFunc) {
	s.POST("/login", l.login).Name = "用户登录"
}

func NewLoginController(l *biz.LoginHandler) RegisterRouter {
	return &loginController{l: l}
}

func (l *loginController) login(c echo.Context) error {
	var loginParam param.Login
	err := BindAndValidate(&loginParam, c)
	if err != nil {
		return err
	}

	login, err := l.l.Login(c.Request().Context(), loginParam.Account, loginParam.Password)
	if err != nil {
		return err
	}

	return resp.OneDataResponse(login, c)
}