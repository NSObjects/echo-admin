/*
 *
 * role.go
 * service
 *
 * Created by lintao on 2023/11/15 15:40
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package service

import (
	"github.com/NSObjects/echo-admin/internal/api/biz"
	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/resp"
	"github.com/labstack/echo/v4"
)

type RoleController struct {
	h *biz.RoleHandler
}

func (r *RoleController) RegisterRouter(s *echo.Group, middlewareFunc ...echo.MiddlewareFunc) {
	s.GET("/roles", r.List).Name = "角色列表"
	s.POST("/roles", r.Create).Name = "创建角色"
	s.PUT("/roles/:id", r.Update).Name = "更新角色"
	s.DELETE("/roles/:id", r.Delete).Name = "删除角色"
	s.PUT("/roles/:id/menus", r.UpdateRoleMenus).Name = "更新角色菜单"

}

func NewRoleController(h *biz.RoleHandler) RegisterRouter {
	return &RoleController{h: h}
}

func (r *RoleController) List(c echo.Context) error {
	var query param.RoleQuery
	if err := BindAndValidate(&query, c); err != nil {
		return err
	}

	list, total, err := r.h.List(c.Request().Context(), query)
	if err != nil {
		return err
	}

	return resp.ListDataResponse(list, total, c)
}

func (r *RoleController) Create(c echo.Context) error {
	var role model.Role
	err := BindAndValidate(&role, c)
	if err != nil {
		return err
	}

	err = r.h.Create(c.Request().Context(), &role)
	if err != nil {
		return err
	}

	return resp.OperateSuccess(c)
}

func (r *RoleController) Update(c echo.Context) error {
	panic("implement me")
}

func (r *RoleController) Delete(c echo.Context) error {
	panic("implement me")
}

func (r *RoleController) UpdateRoleMenus(c echo.Context) error {
	panic("implement me")
}
