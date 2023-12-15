/*
 * Created by lintao on 2023/7/26 下午2:39
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package param

import (
	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/tools"
	"gorm.io/gen/field"
)

type UserParam struct {
	APIQuery
	model.User
}

type UserResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Status       int    `json:"status"`
	Account      string `json:"account"`
	DepartmentID uint   `json:"department_id"`
	RoleID       uint   `json:"role_id"`
	Sex          int    `json:"sex"`
	Posts        string `json:"posts"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	Password     string `json:"password"`
	CreatedAt    string `json:"created_at"`
}

type UserBody struct {
	// 账号
	Account *string `json:"account,omitempty"`
	// 头像
	Avatar *string `json:"avatar,omitempty"`
	// 部门Id
	DepartmentID *uint `json:"department_id,omitempty"`
	// 邮箱
	Email *string `json:"email,omitempty"`
	ID    *uint   `json:"id,omitempty"`
	// 昵称
	Name *string `json:"name,omitempty"`
	// 密码
	Password *string `json:"password,omitempty"`
	// 手机号码
	Phone *string `json:"phone,omitempty"`
	// 岗位
	Posts *string `json:"posts,omitempty"`
	// 角色id
	RoleID *uint `json:"role_id,omitempty"`
	// 性别
	Sex *int `json:"sex,omitempty"`
	// 状态
	Status *int `json:"status,omitempty"`
}

func (u UserBody) Data() ([]field.Expr, model.User) {
	var filed []field.Expr
	var model model.User
	if u.Name != nil {
		filed = append(filed, query.Q.User.Name)
		model.Name = *u.Name
	}
	if u.Account != nil {
		filed = append(filed, query.Q.User.Account)
		model.Account = *u.Account
	}

	if u.Avatar != nil {
		filed = append(filed, query.Q.User.Avatar)
		model.Avatar = *u.Avatar
	}

	if u.Password != nil {
		filed = append(filed, query.Q.User.Password)
		model.Password = tools.Sha25(*u.Password)
	}

	if u.Phone != nil {
		filed = append(filed, query.Q.User.Phone)
		model.Phone = *u.Phone
	}

	if u.RoleID != nil {
		filed = append(filed, query.Q.User.RoleID)
		model.RoleID = *u.RoleID
	}

	if u.DepartmentID != nil {
		filed = append(filed, query.Q.User.DepartmentID)
		model.DepartmentID = *u.DepartmentID
	}

	if u.Status != nil {
		filed = append(filed, query.Q.User.Status)
		model.Status = *u.Status
	}

	if u.Sex != nil {
		filed = append(filed, query.Q.User.Sex)
		model.Sex = *u.Sex
	}

	if u.Posts != nil {
		filed = append(filed, query.Q.User.Posts)
		model.Posts = *u.Posts
	}
	if u.Email != nil {
		filed = append(filed, query.Q.User.Email)
		model.Email = *u.Email
	}

	return filed, model
}
