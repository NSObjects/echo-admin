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
	"github.com/NSObjects/echo-admin/internal/api/data/query"

	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/code"

	"github.com/go-sql-driver/mysql"
	"github.com/marmotedu/errors"
	"gorm.io/gen/field"
)

type MenuHandler struct {
	q *query.Query
}

func NewMenuHandler(q *query.Query) *MenuHandler {
	return &MenuHandler{q: q}
}

func (m *MenuHandler) CreateMenu(ctx context.Context, menu param.Menu) (err error) {

	filed, mm := menu.Data()
	if err = m.q.Menu.WithContext(ctx).Select(filed...).Create(&mm); err != nil {
		return errors.WrapC(err, code.ErrDatabase, fmt.Sprintf("创建菜单失败 %v", menu))
	}

	return nil
}

func (m *MenuHandler) ListMenu(ctx context.Context, q param.APIQuery) ([]*model.Menu, int64, error) {
	menus, err := m.q.Menu.Offset(q.Offset()).
		Limit(q.Limit()).Where(m.q.Menu.Pid.IsNull(), m.q.Menu.Layout.IsNull()).
		Preload(field.Associations).WithContext(ctx).Find()

	if err != nil {
		return nil, 0, errors.WrapC(err, code.ErrDatabase, "查询菜单列表失败")
	}

	total, err := m.q.Menu.Where(m.q.Menu.Pid.IsNotNull()).WithContext(ctx).Count()
	if err != nil {
		return nil, 0, err
	}

	return menus, total, nil

}

func (m *MenuHandler) UpdateMenu(ctx context.Context, id uint, menu param.Menu) error {

	selection, update := menu.Data()

	_, err := m.q.Menu.WithContext(ctx).Select(selection...).Where(m.q.Menu.ID.Eq(id)).Updates(update)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1452 { //nolint:gomnd
				return errors.WrapC(err, code.ErrParentMenuExisted, "父级菜单不存在")
			}
		}

		return errors.WrapC(err, code.ErrDatabase, fmt.Sprintf("更新菜单失败 %v", menu))
	}

	return nil
}

func (m *MenuHandler) Delete(ctx context.Context, id uint) error {
	_, err := m.q.Menu.WithContext(ctx).Where(m.q.Menu.ID.Eq(id)).Delete()
	if err != nil {
		return errors.WrapC(err, code.ErrDatabase, fmt.Sprintf("删除菜单失败 %v", id))
	}

	return nil
}
