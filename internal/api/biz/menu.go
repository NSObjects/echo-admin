/*
 *
 * menu.go
 * biz
 *
 * Created by lintao on 2023/11/14 10:58
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package biz

import (
	"context"
	"fmt"
	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/code"
	"gorm.io/gen/field"

	"github.com/NSObjects/echo-admin/query"
	"github.com/marmotedu/errors"
)

type MenuHandler struct {
	q *query.Menu
}

func NewMenuHandler(q *query.Query) *MenuHandler {
	return &MenuHandler{q: q}
}

func (m *MenuHandler) CreateMenu(ctx context.Context, menu param.Menu) (err error) {

	mm := model.Menu{}

	mm.Name = menu.Name
	mm.Path = menu.Path
	mm.Component = menu.Component
	mm.Redirect = menu.Redirect
	mm.ParentID = menu.ParentId

	if err = m.q.Menu.WithContext(ctx).Create(&mm); err != nil {
		return errors.WrapC(err, code.ErrDatabase, fmt.Sprintf("创建菜单失败 %v", menu))
	}

	return nil
}

func (m *MenuHandler) ListMenu(ctx context.Context, q param.APIQuery) ([]*model.Menu, int64, error) {
	menus, err := m.q.Menu.Offset(q.Offset()).
		Limit(q.Limit()).Where(m.q.Menu.ParentID.IsNull()).
		Preload(field.Associations).WithContext(ctx).Find()

	if err != nil {
		return nil, 0, errors.WrapC(err, code.ErrDatabase, "查询菜单列表失败")
	}

	total, err := m.q.Menu.Where(m.q.Menu.ParentID.IsNotNull()).WithContext(ctx).Count()
	if err != nil {
		return nil, 0, err
	}

	return menus, total, nil

}
