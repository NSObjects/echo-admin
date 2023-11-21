/*
 *
 * deparment.go
 * service
 *
 * Created by lintao on 2023/11/21 10:08
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package service

import (
	"strconv"

	"github.com/NSObjects/echo-admin/internal/api/biz"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/resp"
	"github.com/labstack/echo/v4"
)

type departmentController struct {
	h *biz.DepartmentHandler
}

func NewDepartmentController(h *biz.DepartmentHandler) RegisterRouter {
	return &departmentController{h: h}
}

func (d *departmentController) RegisterRouter(s *echo.Group, middlewareFunc ...echo.MiddlewareFunc) {
	s.GET("/departments/:id", d.get)
	s.GET("/departments", d.list)
	s.DELETE("/departments/:id", d.delete)
	s.PUT("/departments/:id", d.update)
	s.POST("/departments", d.create)
}

func (d *departmentController) create(c echo.Context) error {
	var de param.Department
	err := BindAndValidate(&de, c)
	if err != nil {
		return err
	}

	if err = d.h.Create(c.Request().Context(), de); err != nil {
		return err
	}

	return resp.OperateSuccess(c)

}

func (d *departmentController) update(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var de param.Department
	if err := BindAndValidate(&de, c); err != nil {
		return err
	}

	if err := d.h.Update(c.Request().Context(), uint(id), de); err != nil {
		return err
	}

	return resp.OperateSuccess(c)
}

func (d *departmentController) delete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := d.h.Delete(c.Request().Context(), uint(id)); err != nil {
		return err
	}

	return resp.OperateSuccess(c)
}

func (d *departmentController) list(c echo.Context) error {
	var de param.DepartmentQuery
	if err := BindAndValidate(&de, c); err != nil {
		return err
	}

	list, total, err := d.h.List(c.Request().Context(), de)
	if err != nil {
		return err
	}

	return resp.ListDataResponse(list, total, c)
}

func (d *departmentController) get(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	department, err := d.h.Get(c.Request().Context(), uint(id))
	if err != nil {
		return err
	}

	return resp.OneDataResponse(department, c)
}
