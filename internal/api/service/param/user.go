/*
 * Created by lintao on 2023/7/26 下午2:39
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package param

import (
	"github.com/NSObjects/echo-admin/internal/api/data/model"
)

type UserParam struct {
	APIQuery
	model.User
}

type UserResponse struct {
	ID        uint   `json:"id" form:"id" query:"id"`
	Name      string `json:"name" form:"name" query:"name"`
	Phone     string `json:"phone" form:"phone" query:"phone"`
	Status    int64  `json:"status" form:"status" query:"status"`
	Avatar    string `json:"avatar" form:"avatar" query:"avatar"`
	Password  string `json:"password" form:"password" query:"password"`
	CreatedAt string `json:"created_at" form:"created_at" query:"created_at"`
}

type UserCreateParam struct {
	Account      string `json:"account"`
	Avatar       string `json:"avatar,omitempty"`
	DepartmentID uint   `json:"department_id,omitempty"`
	Name         string `json:"name"`
	Password     string `json:"password"`
	Phone        string `json:"phone"`
	RoleID       uint   `json:"role_id,omitempty"`
	Status       int64  `json:"status,omitempty"`
}
