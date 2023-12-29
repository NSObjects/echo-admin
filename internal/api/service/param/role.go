/*
 *
 * role.go
 * param
 *
 * Created by lintao on 2023/11/15 16:20
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package param

import (
	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"gorm.io/gen/field"
)

type RoleQuery struct {
	Name  string `json:"name" `
	State int    `json:"state" `
	APIQuery
}

type RoleResp struct {
	Id       uint    `json:"id"`
	Name     string  `json:"name"`
	Sort     int     `json:"sort"`
	Status   int     `json:"status"`
	Mark     string  `json:"mark"`
	CreateAt string  `json:"create_at"`
	Menus    []int64 `json:"menus"`
}

// Role 角色
// Mark 备注
// Menus 关联菜单
// Name 角色名称
// Order 排序
// Status 状态 1=启用 其他禁用
type Role struct {
	Mark   *string `json:"mark,omitempty"`
	Menus  []uint  `json:"menus,omitempty"`
	Name   *string `json:"name,omitempty"`
	Order  *int    `json:"order,omitempty"`
	Status *int    `json:"status,omitempty"`
}

func (r Role) Data() ([]field.Expr, model.Role) {
	var filed []field.Expr
	var value model.Role

	if r.Mark != nil && *r.Mark != "" {
		filed = append(filed, query.Role.Mark)
		value.Mark = *r.Mark
	}

	if r.Order != nil && *r.Order != 0 {
		filed = append(filed, query.Role.Order_)
		value.Order = *r.Order
	}

	if r.Name != nil && *r.Name != "" {
		filed = append(filed, query.Role.Name)
		value.Name = *r.Name
	}

	if r.Status != nil && *r.Status != 0 {
		filed = append(filed, query.Role.Status)
		value.Status = *r.Status
	}

	//if r.Menus != nil && len(r.Menus) > 0 {
	//	for _, m := range r.Menus {
	//		value.Menus = append(value.Menus, model.Menu{ID: uint(m)})
	//	}
	//
	//	filed = append(filed, query.Role.Menus.Field())
	//}

	return filed, value
}
