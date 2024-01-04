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
	"log/slog"

	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/code"
	"github.com/NSObjects/echo-admin/internal/log"
	"github.com/go-sql-driver/mysql"
	"github.com/marmotedu/errors"
	"github.com/samber/lo"
)

type MenuHandler struct {
	q *query.Query
	//e *casbin.Enforcer
}

func NewMenuHandler(q *query.Query) *MenuHandler {
	return &MenuHandler{q: q}
}

func (m *MenuHandler) CreateMenu(ctx context.Context, menu param.Menu) error {
	if menu.PID != nil && *menu.PID != 0 {
		parent, err := m.q.Menu.WithContext(ctx).Where(m.q.Menu.ID.Eq(uint(*menu.PID))).First()
		if err != nil {
			return errors.WrapC(err, code.ErrParentMenuExisted, "父级菜单不存在")
		}
		if parent.Type == model.MenuTypeButton ||
			(parent.Type == model.MenuTypeMenu && menu.Type != model.MenuTypeButton) {
			return errors.WithCode(code.ErrNotAllowCreate, "父级菜单类型不正确")
		}
	}

	return m.q.Transaction(func(tx *query.Query) error {
		mm, err := menu.Model()
		if err != nil {
			return errors.WrapC(err, code.ErrDatabase, fmt.Sprintf("创建菜单失败 %v", menu))
		}

		if err = m.q.Menu.WithContext(ctx).Create(&mm); err != nil {
			return errors.WrapC(err, code.ErrDatabase, fmt.Sprintf("创建菜单失败 %v", menu))
		}
		return nil
	})
}

func (m *MenuHandler) ListMenu(ctx context.Context, q param.MenuParam) ([]param.MenuResp, int64, error) {
	condition := q.Condition()
	needChild := false
	countCondition := condition
	countCondition = append(countCondition, m.q.Menu.Pid.IsNotNull())
	if len(condition) == 0 {
		condition = append(condition, m.q.Menu.Pid.IsNull(), m.q.Menu.Layout.IsNull())
		needChild = true
	}

	menus, err := m.q.Menu.Offset(q.Offset()).
		Limit(q.Limit()).Where(condition...).WithContext(ctx).Find()

	if err != nil {
		return nil, 0, errors.WrapC(err, code.ErrDatabase, "查询菜单列表失败")
	}

	if needChild {
		for _, td := range menus {
			td.Children, err = m.GetAllMenu(td.ID)
			if err != nil {
				log.Error(err, slog.Int64("菜单ID", int64(td.ID)))
			}
		}
	}

	total, err := m.q.Menu.Where(countCondition...).WithContext(ctx).Count()
	if err != nil {
		return nil, 0, errors.WrapC(err, code.ErrDatabase, "统计菜单数量失败")
	}
	resp, err := param.MenuModelResp(menus)
	if err != nil {
		return nil, 0, errors.WrapC(err, code.ErrDecodingFailed, "菜单数据转换失败")
	}
	return resp, total, nil

}

func (m *MenuHandler) GetAllMenu(parentID uint) ([]*model.Menu, error) {
	if parentID == 0 {
		return nil, nil
	}
	departments, err := m.q.Menu.Where(m.q.Menu.Pid.Eq(int64(parentID))).
		Preload(m.q.Menu.Children).Preload(m.q.Menu.API).Find()
	if err != nil {
		return nil, err
	}
	for i, department := range departments {
		children, err := m.GetAllMenu(department.ID)
		if err != nil {
			return nil, err
		}
		departments[i].Children = children
	}

	return lo.Map(departments, func(item *model.Menu, index int) *model.Menu {
		return item
	}), nil
}

func (m *MenuHandler) UpdateMenu(ctx context.Context, id uint, menu param.Menu) error {
	return m.q.Transaction(func(tx *query.Query) error {
		selection, update := menu.Data()
		update.ID = id
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

		err = m.q.Menu.API.Model(&update).Clear()
		if err != nil {
			return errors.WrapC(err, code.ErrDatabase, fmt.Sprintf("清除菜单API失败 %v", menu))
		}

		if len(menu.API) <= 0 {
			return nil
		}

		apis := make([]*model.API, len(menu.API))
		for index, menuAPI := range menu.API {
			apis[index] = &model.API{
				Path:   menuAPI.URL,
				Method: string(menuAPI.Method),
				Name:   menuAPI.Name,
			}
		}

		first, err := m.q.Menu.Where(m.q.Menu.ID.Eq(id)).First()
		if err != nil {
			return errors.WrapC(err, code.ErrDatabase, fmt.Sprintf("创建菜单API失败 %v", menu))
		}

		if err = m.q.Menu.API.Model(first).Append(apis...); err != nil {
			return errors.WrapC(err, code.ErrDatabase, fmt.Sprintf("创建菜单API失败 %v", menu))
		}
		return nil
	})
}

func (m *MenuHandler) Delete(ctx context.Context, id uint) error {

	return m.q.Transaction(func(tx *query.Query) error {
		err := m.cleanChildMenu(ctx, int64(id))
		if err != nil {
			return errors.WrapC(err, code.ErrDatabase, fmt.Sprintf("删除菜单失败 %v", id))
		}
		if _, err = m.q.Menu.WithContext(ctx).
			Where(m.q.Menu.ID.Eq(id)).
			Delete(); err != nil {
			return errors.WrapC(err, code.ErrDatabase, fmt.Sprintf("删除菜单失败 %v", id))
		}
		return nil
	})
}

// cleanChildMenu
func (m *MenuHandler) cleanChildMenu(ctx context.Context, id int64) error {
	if id <= 0 {
		return nil
	}

	find, err := m.q.Menu.WithContext(ctx).Where(m.q.Menu.Pid.Eq(id)).Find()
	if err != nil {
		return err
	}
	for _, v := range find {
		if _, err = m.q.Menu.WithContext(ctx).
			Where(m.q.Menu.ID.Eq(v.ID)).
			Delete(); err != nil {
			log.Error(err)
		}

		if err = m.cleanChildMenu(ctx, int64(v.ID)); err != nil {
			continue
		}
	}
	return nil
}
