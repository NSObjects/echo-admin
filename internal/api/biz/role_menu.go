/*
 *
 * role_menu.go
 * biz
 *
 * Created by lintao on 2023/11/16 10:52
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package biz

import (
	"context"

	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/query"
)

type RoleMenuHandler struct {
	q *query.Query
}

func NewRoleMenuHandler(q *query.Query) *RoleMenuHandler {
	return &RoleMenuHandler{q: q}
}

func (r *RoleMenuHandler) Create(ctx context.Context, roleId int, menuId int) error {
	if err := r.q.RoleMenu.WithContext(ctx).Create(&model.RoleMenu{
		RoleId: int64(roleId),
		MenuId: int64(menuId),
	}); err != nil {
		return err
	}
	return nil
}

func (r *RoleMenuHandler) BatchCreate(ctx context.Context, roleID int64, menuIDs []int64) error {
	roleMenus := make([]*model.RoleMenu, len(menuIDs))
	for index, v := range menuIDs {
		roleMenus[index] = &model.RoleMenu{
			RoleId: roleID,
			MenuId: v,
		}
	}

	err := r.q.RoleMenu.WithContext(ctx).Create(roleMenus...)
	if err != nil {
		return err
	}

	return nil
}

func (r *RoleMenuHandler) CleanRoleMenu(ctx context.Context, roleID int64) error {
	_, err := r.q.RoleMenu.WithContext(ctx).Where(r.q.RoleMenu.RoleId.Eq(roleID)).Delete()
	if err != nil {
		return err
	}

	return nil
}
