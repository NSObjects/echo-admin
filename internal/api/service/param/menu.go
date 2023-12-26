/*
 *
 * menu.go
 * param
 *
 * Created by lintao on 2023/11/14 11:07
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package param

import (
	"time"

	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/data/query"

	"gorm.io/gen/field"
)

// Menu 菜单
// Type: 1:目录 2:菜单 3:按钮
// API: 接口规则
// Link: 外链地址
// Identify: 菜单标识
// Sort: 排序
// Hidden: 是否隐藏 1=是 2=否
// Cache: 是否缓存 1=是 2=否
// Fixed: 是否固定 1=是 2=否
// Name: 菜单名称
// Path: 路由地址
// Component: 组件路径
// PID: 父菜单ID
// Icon: 图标
// Remark: 备注
type Menu struct {
	Type      model.MenuType `json:"type" copier:"type"`
	API       *string        `json:"api,omitempty"`
	Cache     *int           `json:"cache,omitempty"`
	Component *string        `json:"component,omitempty"`
	Fixed     *int           `json:"fixed,omitempty"`
	Hidden    *int           `json:"hidden,omitempty"`
	Icon      *string        `json:"icon,omitempty"`
	Identify  *int           `json:"identify,omitempty"`
	Layout    *int           `json:"layout,omitempty"`
	Link      *string        `json:"link,omitempty"`
	Name      *string        `json:"name,omitempty"`
	Path      *string        `json:"path,omitempty"`
	PID       *int64         `json:"pid,omitempty"`
	Redirect  *string        `json:"redirect,omitempty"`
	Remark    *string        `json:"remark,omitempty"`
	Role      []int64        `json:"role,omitempty"`
	Sort      *int           `json:"sort,omitempty"`
	Status    *int64         `json:"status,omitempty"`
	Routes    []Menu         `json:"routes"`
}

type RoleMenu struct {
	MenuID  []int64 `json:"menu_id" form:"menu_id" query:"menu_id"`
	Creator string  `json:"creator" form:"creator" query:"creator"`
}

type MenuResp struct {
	ID        uint         `json:"id"`
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Component string       `json:"component"`
	Redirect  string       `json:"redirect"`
	Pid       int64        `json:"pid,omitempty"`
	Routes    []model.Menu `json:"routes"`
	CreatedAt time.Time    `json:"created_at" `
	UpdatedAt time.Time    `json:"updated_at" `
}

func (m Menu) Data() ([]field.Expr, model.Menu) {
	var filed []field.Expr
	var menu model.Menu
	if m.Name != nil {
		filed = append(filed, query.Q.Menu.Name)
		menu.Name = *m.Name
	}
	if m.Path != nil {
		filed = append(filed, query.Q.Menu.Path)
		menu.Path = *m.Path
	}
	if m.Component != nil {
		filed = append(filed, query.Q.Menu.Component)
		menu.Component = *m.Component
	}
	if m.Redirect != nil {
		filed = append(filed, query.Q.Menu.Redirect)
		menu.Redirect = *m.Redirect
	}
	if m.Layout != nil {
		filed = append(filed, query.Q.Menu.Layout)
		menu.Layout = *m.Layout
	}
	if m.PID != nil {
		filed = append(filed, query.Q.Menu.Pid)
		menu.Pid = *m.PID
	}
	if m.Icon != nil {
		filed = append(filed, query.Q.Menu.Icon)
		menu.Icon = *m.Icon
	}
	if m.Type != 0 {
		filed = append(filed, query.Q.Menu.Type)
		menu.Type = m.Type
	}
	if m.API != nil {
		filed = append(filed, query.Q.Menu.API)
		menu.API = *m.API
	}
	if m.Link != nil {
		filed = append(filed, query.Q.Menu.Link)
		menu.Link = *m.Link
	}
	if m.Remark != nil {
		filed = append(filed, query.Q.Menu.Remark)
		menu.Remark = *m.Remark
	}
	if m.Hidden != nil {
		filed = append(filed, query.Q.Menu.Hidden)
		menu.Hidden = *m.Hidden
	}
	if m.Cache != nil {
		filed = append(filed, query.Q.Menu.Cache)
		menu.Cache = *m.Cache
	}
	if m.Fixed != nil {
		filed = append(filed, query.Q.Menu.Fixed)
		menu.Fixed = *m.Fixed
	}
	if m.Sort != nil {
		filed = append(filed, query.Q.Menu.Sort)
		menu.Sort = *m.Sort
	}
	if m.Identify != nil {
		filed = append(filed, query.Q.Menu.Identifier)
		menu.Identifier = *m.Identify
	}
	return filed, menu
}
