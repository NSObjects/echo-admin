/*
 *
 * menu.go
 * service
 *
 * Created by lintao on 2023/11/10 13:52
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package service

import (
	"github.com/NSObjects/echo-admin/internal/api/biz"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/resp"
	//nolint:goimports
	"github.com/labstack/echo/v4"
)

type MenuController struct {
	h *biz.MenuHandler
}

func (m *MenuController) RegisterRouter(s *echo.Group, middlewareFunc ...echo.MiddlewareFunc) {
	s.POST("/menus", m.create).Name = "创建菜单"
	s.GET("/menus", m.list).Name = "菜单列表"
}

func NewMenuController(h *biz.MenuHandler) *MenuController {
	return &MenuController{h: h}
}

func (m *MenuController) create(ctx echo.Context) (err error) {
	var menu param.Menu
	if err = BindAndValidate(&menu, ctx); err != nil {
		return err
	}

	if err = m.h.CreateMenu(ctx.Request().Context(), menu); err != nil {
		return err
	}
	return resp.OperateSuccess(ctx)
}

func (m *MenuController) list(ctx echo.Context) error {
	var q param.APIQuery

	if err := BindAndValidate(&q, ctx); err != nil {
		return err
	}

	menu, total, err := m.h.ListMenu(ctx.Request().Context(), q)
	if err != nil {
		return err
	}

	return resp.ListDataResponse(menu, total, ctx)
}
