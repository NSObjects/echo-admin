/*
 * Created by lintao on 2023/7/18 下午3:59
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package biz

import (
	"context"
	"time"

	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/code"
	"github.com/NSObjects/echo-admin/internal/log"
	"github.com/marmotedu/errors"
	"github.com/samber/lo"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

type UserHandler struct {
	q *query.Query
}

func NewUserHandler(q *query.Query) *UserHandler {
	return &UserHandler{q: q}
}

func (h *UserHandler) ListUser(p param.UserParam) ([]param.UserResponse, int64, error) {
	do := h.q.User
	var cd []gen.Condition

	if p.Phone != nil && *p.Phone != "" {
		cd = append(cd, do.Phone.Like(*p.Phone+"%"))
	}

	if p.Status != 0 {
		cd = append(cd, do.Status.Eq(p.Status))
	}

	if p.DepartmentId != nil && *p.DepartmentId != 0 {
		cd = append(cd, do.DepartmentID.Eq(*p.DepartmentId))
	}

	if p.CreateStart != nil && p.CreateEnd != nil && *p.CreateEnd > *p.CreateStart {
		start, err := time.Parse("2006-01-02 15:04:05", *p.CreateStart)
		if err != nil {
			return nil, 0, err
		}
		end, err := time.Parse("2006-01-02 15:04:05", *p.CreateEnd)
		if err != nil {
			return nil, 0, err
		}
		cd = append(cd, do.CreatedAt.Between(start, end))
	}

	ido := do.Where(cd...)
	if p.Key != nil && *p.Key != "" {
		ido = do.Or(do.Name.Like(*p.Key + "%")).Or(do.Account.Like(*p.Key + "%"))
	}

	users, err := ido.Limit(p.Limit()).Offset(p.Offset()).Preload(h.q.User.Role).Find()
	if err != nil {
		return nil, 0, err
	}

	resp := make([]param.UserResponse, len(users))
	for i, user := range users {
		resp[i] = param.UserResponse{
			Name:         user.Name,
			Phone:        user.Phone,
			Status:       user.Status,
			Avatar:       user.Avatar,
			Posts:        user.Posts,
			Email:        user.Email,
			Account:      user.Account,
			RoleID:       user.Role,
			DepartmentID: user.DepartmentID,
			ID:           user.ID,
			Sex:          user.Sex,
			CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	total, err := ido.Count()
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}

func (h *UserHandler) CreateUser(param param.UserBody) (err error) {
	return h.q.Transaction(func(tx *query.Query) error {
		selection, m := param.Data()
		if err = h.q.User.Select(selection...).Create(&m); err != nil {
			return errors.WrapC(err, code.ErrDatabase, "创建用户失败")
		}

		if param.RoleID != nil && len(*param.RoleID) > 0 {
			var value []*model.Role
			for _, menuID := range *param.RoleID {
				value = append(value, &model.Role{ID: menuID})
			}
			if err = h.q.User.Role.Model(&m).Append(value...); err != nil {
				return errors.WrapC(err, code.ErrDatabase, "添加用户角色失败")
			}
		}
		return nil
	})
}

func (h *UserHandler) DeleteUser(id int64) (err error) {

	return h.q.Transaction(func(tx *query.Query) error {
		if err = h.q.User.DeleteByID(id); err != nil {
			return errors.WrapC(err, code.ErrDatabase, "删除用户失败")
		}

		if err = h.q.User.Role.Model(&model.User{ID: uint(id)}).Clear(); err != nil {
			return errors.WrapC(err, code.ErrDatabase, "删除用户失败")
		}
		return nil
	})
}

func (h *UserHandler) UpdateUser(user param.UserBody, id int64) error {

	return h.q.Transaction(func(tx *query.Query) error {
		selection, m := user.Data()
		m.ID = uint(id)
		if _, err := h.q.User.Select(selection...).
			Where(h.q.User.ID.Eq(uint(id))).
			Updates(&m); err != nil {
			return errors.WrapC(err, code.ErrDatabase, "更新用户失败")
		}

		if err := h.q.User.Role.Model(&m).Clear(); err != nil {
			return errors.WrapC(err, code.ErrDatabase, "更新用户失败")
		}
		value := make([]*model.Role, len(*user.RoleID))
		for index, menuID := range *user.RoleID {
			value[index] = &model.Role{ID: menuID}
		}
		if err := h.q.User.Role.Model(&m).Append(value...); err != nil {
			return errors.WrapC(err, code.ErrDatabase, "更新用户失败")
		}
		return nil
	})
}

func (h *UserHandler) GetUserDetail(id int64) (param.UserResponse, error) {

	user, err := h.q.User.Preload(h.q.User.Role).GetById(int(id))
	if err != nil {
		return param.UserResponse{}, errors.WrapC(err, code.ErrDatabase, "查询用户详情失败")
	}

	return param.UserResponse{
		Name:         user.Name,
		Phone:        user.Phone,
		Status:       user.Status,
		Password:     user.Password,
		Avatar:       user.Avatar,
		Posts:        user.Posts,
		Email:        user.Email,
		Account:      user.Account,
		RoleID:       user.Role,
		DepartmentID: user.DepartmentID,
		ID:           user.ID,
	}, nil
}

func (h *UserHandler) ListUserMenu(ctx context.Context, id int64) ([]param.MenuResp, int, error) {
	user, err := h.q.User.Preload(h.q.User.Role).Preload(h.q.User.Role.Menus).GetById(int(id))
	if err != nil {
		return nil, 0, err
	}

	var menuIds []uint

	for _, v := range user.Role {
		//find, err := h.q.Role.Where(h.q.Role.ID.Eq(v.)).Find()
		//if err != nil {
		//	return nil, 0, err
		//}
		for _, menu := range v.Menus {
			menuIds = append(menuIds, menu.ID)
		}
	}
	menuIds = lo.Union(menuIds)
	condition := []gen.Condition{h.q.Menu.Pid.IsNull(), h.q.Menu.Layout.IsNull()}
	if len(menuIds) > 0 {
		condition = append(condition, h.q.Menu.ID.In(menuIds...))
	}

	menus, err := h.q.Menu.Where(condition...).
		Preload(field.Associations).WithContext(ctx).Find()

	if err != nil {
		return nil, 0, errors.WrapC(err, code.ErrDatabase, "查询菜单列表失败")
	}
	respMenu := make([]*model.Menu, len(menus))
	for index, td := range menus {
		td.Children, err = h.GetAllMenu(td.ID, menuIds)
		if err != nil {
			log.Error(err)
		}
		respMenu[index] = td
	}
	resp, err := param.MenuModelResp(respMenu)
	if err != nil {
		return nil, 0, err
	}
	return resp, len(menus), nil
}

func (h *UserHandler) GetAllMenu(parentID uint, menuID []uint) ([]*model.Menu, error) {
	if parentID == 0 {
		return nil, nil
	}

	condition := []gen.Condition{h.q.Menu.Pid.Eq(int64(parentID))}
	if len(menuID) > 0 {
		condition = append(condition, h.q.Menu.ID.In(menuID...))
	}
	menus, err := h.q.Menu.Where(condition...).Find()
	if err != nil {
		return nil, err
	}
	for i, menu := range menus {

		children, err := h.GetAllMenu(menu.ID, menuID)
		if err != nil {
			return nil, err
		}
		menus[i].Children = children
	}

	return lo.Map(menus, func(item *model.Menu, index int) *model.Menu {
		return item
	}), nil
}
