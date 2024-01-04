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
	"github.com/NSObjects/echo-admin/internal/log"
	"github.com/jinzhu/copier"
	"github.com/marmotedu/errors"
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
	Type      model.MenuType `json:"type" copier:"Type"`
	API       []ChildAPI     `json:"apis" query:"apis" form:"apis" copier:"API" `
	Component *string        `json:"component,omitempty" copier:"Component"`
	Hidden    *int           `json:"hidden,omitempty" copier:"Hidden"`
	Icon      *string        `json:"icon,omitempty" copier:"Icon"`
	Layout    *int           `json:"layout,omitempty" copier:"Layout"`
	Link      *string        `json:"link,omitempty" copier:"Link"`
	Name      *string        `json:"name,omitempty" copier:"Name"`
	Path      *string        `json:"path,omitempty" copier:"Path"`
	PID       *int64         `json:"pid,omitempty" copier:"Pid"`
	Redirect  *string        `json:"redirect,omitempty" copier:"Redirect"`
	Remark    *string        `json:"remark,omitempty" copier:"Remark"`
	Role      []int64        `json:"role,omitempty" copier:"Role"`
	Sort      *int           `json:"sort,omitempty" copier:"Sort"`
	Status    *int64         `json:"status,omitempty" copier:"Status"`
	//Routes    []Menu         `json:"routes"`
}

type ChildAPI struct {
	Method Method `json:"method" query:"method" form:"method" copier:"Method"`
	URL    string `json:"url" query:"url" form:"url" copier:"URL"`
	Name   string `json:"name" form:"name" query:"name" copier:"Name"`
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
	Children  []MenuResp     `json:"children,omitempty"`
	Component string         `json:"component"`
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

func MentModel(v *model.Menu) (MenuResp, error) {
	var rp MenuResp
	err := copier.CopyWithOption(&rp, v, defaultOpts())
	if err != nil {
		return MenuResp{}, err
	}
	return rp, nil
}

func MenuModelResp(child []*model.Menu) ([]MenuResp, error) {
	if len(child) == 0 {
		return []MenuResp{}, nil
	}

	resp := make([]MenuResp, len(child))
	for index, v := range child {
		rp, err := MentModel(v)
		if err != nil {
			return nil, err
		}
		resp[index] = rp
	}
	return resp, nil
}

func (m Menu) Data() ([]field.Expr, model.Menu) {
	filed := m.Fields()
	menu, err := m.Model()
	if err != nil {
		log.Error(err)
	}

	return filed, menu
}

func (m Menu) Model() (model.Menu, error) {
	var menu model.Menu
	err := copier.CopyWithOption(&menu, &m, defaultOpts())
	if err != nil {
		log.Error(err)
		return model.Menu{}, err
	}
	return menu, nil
}

// Fields
// 数据库插入需要的字段，代码太长还得想办法优化
func (m Menu) Fields() []field.Expr {
	var filed []field.Expr
	if m.Name != nil {
		filed = append(filed, query.Q.Menu.Name)
	}
	if m.Path != nil {
		filed = append(filed, query.Q.Menu.Path)
	}
	if m.Component != nil {
		filed = append(filed, query.Q.Menu.Component)
	}
	if m.Redirect != nil {
		filed = append(filed, query.Q.Menu.Redirect)
	}
	if m.Layout != nil {
		filed = append(filed, query.Q.Menu.Layout)
	}

	if m.PID != nil {
		filed = append(filed, query.Q.Menu.Pid)
	}

	if m.Icon != nil {
		filed = append(filed, query.Q.Menu.Icon)
	}
	if m.Type != 0 {
		filed = append(filed, query.Q.Menu.Type)
	}

	if m.Link != nil {
		filed = append(filed, query.Q.Menu.Link)
	}
	if m.Remark != nil {
		filed = append(filed, query.Q.Menu.Remark)
	}
	if m.Hidden != nil {
		filed = append(filed, query.Q.Menu.Hidden)
	}

	if m.Sort != nil {
		filed = append(filed, query.Q.Menu.Sort)
	}

	return filed
}

func apiConverter(src interface{}) (dst interface{}, err error) {
	m2, ok := src.(ChildAPI)
	if !ok {
		return nil, errors.New("menu type error")
	}

	return model.API{Path: m2.URL, Method: string(m2.Method), Name: m2.Name}, err
}

func apiToModel(src interface{}) (dst interface{}, err error) {
	m2, ok := src.(model.API)
	if !ok {
		return nil, errors.New("menu type error")
	}

	return ChildAPI{URL: m2.Path, Method: Method(m2.Method), Name: m2.Name}, err
}

func defaultOpts() copier.Option {
	return copier.Option{IgnoreEmpty: true, DeepCopy: true, Converters: []copier.TypeConverter{
		{
			SrcType: ChildAPI{},
			DstType: model.API{},
			Fn:      apiConverter,
		},
		{
			SrcType: model.API{},
			DstType: ChildAPI{},
			Fn:      apiToModel,
		},
	}}
}
