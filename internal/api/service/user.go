/*
 * Created by lintao on 2023/7/27 下午1:44
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package service

import (
	"github.com/NSObjects/echo-admin/internal/api/data"
	"github.com/NSObjects/echo-admin/internal/code"
	"github.com/golang-jwt/jwt/v5"
	"github.com/marmotedu/errors"
	"strconv"

	"github.com/NSObjects/echo-admin/internal/api/biz"
	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/resp"
	"github.com/labstack/echo/v4"
)

type userController struct {
	user *biz.UserHandler
}

func (u *userController) RegisterRouter(s *echo.Group, middlewareFunc ...echo.MiddlewareFunc) {
	s.GET("/users", u.getUser).Name = "用户查询"
	s.POST("/users", u.createUser).Name = "创建用户"
	s.DELETE("/users/:id", u.deleteUser).Name = "删除用户"
	s.PUT("/users/:id", u.updateUser).Name = "更新用户"
	s.GET("/users/:id", u.getUserDetail).Name = "获取某个用户信息"
	s.GET("/users/current", u.current).Name = "获取当前用户信息"
}

func NewUserController(u *biz.UserHandler) RegisterRouter {
	return &userController{
		user: u,
	}
}

func (u *userController) current(c echo.Context) (err error) {
	token := c.Get("user").(*jwt.Token)
	if token == nil {
		return errors.WithCode(code.ErrMissingHeader, "token is nil")
	}

	user := token.Claims.(*data.JwtCustomClaims)
	if user == nil {
		return errors.WithCode(code.ErrMissingHeader, "token is nil")
	}

	detail, err := u.user.GetUserDetail(user.ID)
	if err != nil {
		return err
	}

	return resp.OneDataResponse(detail, c)
}

func (u *userController) getUser(c echo.Context) (err error) {
	var user param.UserParam
	if err = BindAndValidate(&user, c); err != nil {
		return err
	}

	listUser, total, err := u.user.ListUser(user.User, user.APIQuery)
	if err != nil {
		return err
	}
	return resp.ListDataResponse(listUser, total, c)
}

func (u *userController) getUserDetail(c echo.Context) (err error) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	detail, err := u.user.GetUserDetail(id)
	if err != nil {
		return err
	}

	return resp.OneDataResponse(detail, c)
}

func (u *userController) createUser(c echo.Context) (err error) {
	var user param.UserCreateParam
	if err = BindAndValidate(&user, c); err != nil {
		return err
	}

	if err = u.user.CreateUser(user); err != nil {
		return err
	}

	return resp.OperateSuccess(c)
}

func (u *userController) updateUser(c echo.Context) (err error) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var user model.User
	if err = BindAndValidate(&user, c); err != nil {
		return err
	}

	if err = u.user.UpdateUser(user, id); err != nil {
		return err
	}

	return resp.OperateSuccess(c)

}

func (u *userController) deleteUser(c echo.Context) (err error) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err = u.user.DeleteUser(id); err != nil {
		return err
	}

	return resp.OperateSuccess(c)
}
