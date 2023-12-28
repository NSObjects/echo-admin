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
	"github.com/casbin/casbin/v2"

	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/samber/lo"
	"gorm.io/gen"
)

type RoleHandler struct {
	q *query.Query
	e *casbin.Enforcer
}

func NewRoleHandler(q *query.Query, e *casbin.Enforcer) *RoleHandler {
	return &RoleHandler{q: q, e: e}
}

func (r *RoleHandler) List(ctx context.Context, q param.RoleQuery) ([]param.RoleResp, int64, error) {
	var cd []gen.Condition

	if q.Name != "" {
		cd = append(cd, r.q.Role.Name.Like(q.Name+"%"))
	}

	if q.State != 0 {
		cd = append(cd, r.q.Role.Status.Eq(q.State))
	}

	roles, err := r.q.Role.Preload(r.q.Role.Menus).
		WithContext(ctx).Where(cd...).
		Limit(q.Limit()).
		Offset(q.Offset()).
		Find()
	if err != nil {
		return nil, 0, err
	}

	total, err := r.q.Role.WithContext(ctx).Where(cd...).Count()
	if err != nil {
		return nil, 0, err
	}
	res := make([]param.RoleResp, len(roles))
	for index, v := range roles {
		res[index] = param.RoleResp{
			Id:       v.ID,
			Name:     v.Name,
			Sort:     v.Order,
			Status:   v.Status,
			Mark:     v.Mark,
			CreateAt: v.CreatedAt.Format("2006-01-02 15:01:05"),
			Menus: lo.Map(v.Menus, func(item model.Menu, index int) int64 {
				return int64(item.ID)
			}),
		}
	}

	return res, total, nil
}

func (r *RoleHandler) Create(ctx context.Context, role *param.Role) error {
	selection, m := role.Data()

	return r.q.Role.WithContext(ctx).Select(selection...).Create(&m)
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
