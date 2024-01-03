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
	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/samber/lo"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

type Method string

const (
	Delete Method = "DELETE"
	Get    Method = "GET"
	Post   Method = "POST"
	Put    Method = "PUT"
)

type MenuParam struct {
	APIQuery
	Component string `json:"component" query:"component" form:"component"`
	Name      string `json:"name" query:"name" form:"name"`
	Path      string `json:"path" query:"path" form:"path"`
	Type      int    `json:"type" query:"type" form:"type"`
}

func (m MenuParam) Condition() []gen.Condition {
	var condition []gen.Condition
	if m.Name != "" {
		condition = append(condition, query.Q.Menu.Name.Like("%"+m.Name+"%"))
	}
	if m.Path != "" {
		condition = append(condition, query.Q.Menu.Path.Like("%"+m.Path+"%"))
	}
	if m.Component != "" {
		condition = append(condition, query.Q.Menu.Component.Like("%"+m.Component+"%"))
	}
	if m.Type != 0 {
		condition = append(condition, query.Q.Menu.Type.Eq(m.Type))
	}

	return condition
}

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
	API       []ChildAPI     `json:"apis" query:"apis" form:"apis"`
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

// ChildAPI api接口
type ChildAPI struct {
	Method Method `json:"method" query:"method" form:"method"`
	URL    string `json:"url" query:"url" form:"url"`
	Name   string `json:"name" form:"name" query:"name"`
}

type RoleMenu struct {
	MenuID  []int64 `json:"menu_id" form:"menu_id" query:"menu_id"`
	Creator string  `json:"creator" form:"creator" query:"creator"`
}

// MenuResp 菜单
// API api接口
// Cache 是否缓存 1=是 2=否
// Children 子菜单
// Component 组件路径
// Fixed 是否固定 1=是 2=否
// Hidden 是否隐藏 1=是 2=否
// Icon 图标
// ID 菜单id
// Identify 菜单标识符
// Label 菜单名称
// Layout 布局
// Link 外链地址
// Name 菜单名称
// Path 路由路径
// PID 父菜单id
// Redirect 重定向
// Remark 备注
// Role 角色id列表
// Sort 排序
// Status 状态 1=启用 2=禁用
// Type 类型 1=目录 2=菜单 3=按钮
type MenuResp struct {
	API       []ChildAPI     `json:"apis,omitempty"`
	Cache     int            `json:"cache,omitempty"`
	Children  []MenuResp     `json:"children,omitempty"`
	Component string         `json:"component"`
	Fixed     int            `json:"fixed,omitempty"`
	Hidden    int            `json:"hidden,omitempty"`
	Icon      string         `json:"icon,omitempty"`
	ID        uint           `json:"id,omitempty"`
	Identify  string         `json:"identify,omitempty"`
	Label     string         `json:"label,omitempty"`
	Layout    int            `json:"layout,omitempty"`
	Link      string         `json:"link,omitempty"`
	Name      string         `json:"name"`
	Path      string         `json:"path"`
	PID       *int64         `json:"pid"`
	Redirect  string         `json:"redirect,omitempty"`
	Remark    string         `json:"remark,omitempty"`
	Role      []int64        `json:"role,omitempty"`
	Sort      int            `json:"sort,omitempty"`
	Status    int64          `json:"status,omitempty"`
	Type      model.MenuType `json:"type"`
	Value     int64          `json:"value,omitempty"`
}

func MentModel(v *model.Menu) MenuResp {

	rp := MenuResp{
		API: lo.Map(v.API, func(item model.API, index int) ChildAPI {
			return ChildAPI{
				Method: Method(item.Method),
				URL:    item.Path,
				Name:   item.Name,
			}
		}),
		Cache:     v.Cache,
		Children:  MenuModelResp(v.Children),
		Component: v.Component,
		Fixed:     v.Fixed,
		Hidden:    v.Hidden,
		Icon:      v.Icon,
		ID:        v.ID,
		Layout:    v.Layout,
		Link:      v.Link,
		Name:      v.Name,
		Path:      v.Path,
		PID:       v.Pid,
		Redirect:  v.Redirect,
		Remark:    v.Remark,
		Sort:      v.Sort,
		Type:      v.Type,
	}
	return rp
}

func MenuModelResp(child []*model.Menu) []MenuResp {
	if len(child) == 0 {
		return []MenuResp{}
	}

	resp := make([]MenuResp, len(child))
	for index, v := range child {
		rp := MentModel(v)
		rp.Children = MenuModelResp(v.Children)
		resp[index] = rp
	}
	return resp
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
		menu.Pid = m.PID
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
		//filed = append(filed, query.Q.Menu.API.Field())
		for _, v := range m.API {
			menu.API = append(menu.API, model.API{
				Path:   v.URL,
				Method: string(v.Method),
			})
		}
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
