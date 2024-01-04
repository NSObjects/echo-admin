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

	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/code"
	"github.com/NSObjects/echo-admin/internal/log"
	"github.com/casbin/casbin/v2"
	"github.com/marmotedu/errors"
	"github.com/samber/lo"
	"github.com/spf13/cast"
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
	err := r.q.Role.WithContext(ctx).Select(selection...).Create(&m)
	if err != nil {
		return errors.WrapC(err, code.ErrDatabase, "创建角色失败")
	}
	if role.Menus != nil && len(role.Menus) > 0 {
		var value []*model.Menu
		for _, menuID := range role.Menus {
			value = append(value, &model.Menu{ID: menuID})
		}
		if err = r.q.Role.Menus.Model(&m).Append(value...); err != nil {
			return err
		}
	}

	policies := r.Policies(ctx, role.Menus, m.ID)
	if len(policies) > 0 {
		if _, err = r.e.AddPolicies(policies); err != nil {
			return err
		}
	}

	return nil
}

func (r *RoleHandler) Policies(ctx context.Context, menusID []uint, roleId uint) [][]string {
	if len(menusID) == 0 {
		return nil
	}
	find, err := r.q.Menu.Where(r.q.Menu.ID.In(menusID...)).WithContext(ctx).Find()
	if err != nil {
		log.Error(err)
		return nil
	}
	rules := make([][]string, len(find))
	for index, v := range find {
		for _, api := range v.API {
			rules[index] = []string{cast.ToString(roleId), cast.ToString(api.ID), api.Path, api.Method}
		}

	}
	return rules
}

func (r *RoleHandler) Update(ctx context.Context, id uint, role *param.Role) error {

	selection, m := role.Data()
	m.ID = id
	_, err := r.q.Role.WithContext(ctx).Select(selection...).Where(r.q.Role.ID.Eq(id)).Updates(&m)
	if err != nil {
		return err
	}
	if err = r.q.Role.Menus.Model(&m).Clear(); err != nil {
		return err
	}
	value := make([]*model.Menu, len(role.Menus))
	for index, menuID := range role.Menus {
		value[index] = &model.Menu{ID: menuID}
	}
	m.ID = id
	if err = r.q.Role.Menus.Model(&m).Append(value...); err != nil {
		return err
	}

	_, err = r.e.RemoveFilteredPolicy(0, cast.ToString(id))
	if err != nil {
		return err
	}

	policies := r.Policies(ctx, role.Menus, id)
	if len(policies) > 0 {
		if _, err = r.e.AddNamedPolicy(cast.ToString(m.ID), policies); err != nil {
			return err
		}
	}

	return nil
}

func (r *RoleHandler) Delete(ctx context.Context, id uint) error {
	_, err := r.q.Role.WithContext(ctx).Where(r.q.Role.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if err = r.q.Role.Menus.Model(&model.Role{ID: id}).Clear(); err != nil {
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
