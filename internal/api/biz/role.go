/*
 *
 * role.go
 * biz
 *
 * Created by lintao on 2023/11/15 15:51
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package biz

import (
	"context"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"time"

	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/service/param"

	"gorm.io/gen"
)

type RoleHandler struct {
	q *query.Query
}

func NewRoleHandler(q *query.Query) *RoleHandler {
	return &RoleHandler{q: q}
}

func (r *RoleHandler) List(ctx context.Context, q param.RoleQuery) ([]*model.Role, int64, error) {
	var cd []gen.Condition
	if q.Name != "" {
		cd = append(cd, r.q.Role.Name.Eq(q.Name))
	}
	if q.Identify != "" {
		cd = append(cd, r.q.Role.Where(r.q.Role.Identify.Eq(q.Identify)))
	}

	if q.State != 0 {
		cd = append(cd, r.q.Role.Where(r.q.Role.Where(r.q.Role.State.Eq(q.State))))
	}

	if q.StartDate > q.EndDate {
		cd = append(cd, r.q.Role.Where(r.q.Role.CreatedAt.Between(time.Unix(q.StartDate, 0), time.Unix(q.EndDate, 0))))

	}

	roles, err := r.q.Role.WithContext(ctx).Where(cd...).Limit(q.Limit()).Offset(q.Offset()).Find()
	if err != nil {
		return nil, 0, err
	}

	total, err := r.q.Role.WithContext(ctx).Where(cd...).Count()
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *RoleHandler) Create(ctx context.Context, role *model.Role) error {
	return r.q.Role.WithContext(ctx).Create(role)
}

func (r *RoleHandler) Update(ctx context.Context, id uint, role *model.Role) error {
	_, err := r.q.Role.WithContext(ctx).Where(r.q.Role.ID.Eq(id)).Updates(role)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoleHandler) Delete(ctx context.Context, id uint) error {
	_, err := r.q.Role.WithContext(ctx).Where(r.q.Role.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (r *RoleHandler) Get(ctx context.Context, id uint) (*model.Role, error) {
	role, err := r.q.Role.WithContext(ctx).Where(r.q.Role.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleHandler) UpdateRoleMenu(ctx context.Context, roleID int64, menuIDs []int64) error {

	first, err := r.q.Role.WithContext(ctx).Where(r.q.Role.ID.Eq(uint(roleID))).First()
	if err != nil {
		return err
	}

	err = r.q.Role.Menus.Model(first).Clear()
	if err != nil {
		return err
	}

	roleMenus := make([]*model.Menu, len(menuIDs))
	for index, v := range menuIDs {
		roleMenus[index] = &model.Menu{ID: uint(v)}
	}
	err = r.q.Role.Menus.Model(first).Append(roleMenus...)
	if err != nil {
		return err
	}

	return nil
}
