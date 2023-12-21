/*
 * Created by lintao on 2023/7/18 下午3:59
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package biz

import (
	"time"

	"gorm.io/gen"

	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
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

	users, err := ido.Limit(p.Limit()).Offset(p.Offset()).Find()
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
			RoleID:       user.RoleID,
			DepartmentID: user.DepartmentID,
			ID:           user.ID,
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

	selection, model := param.Data()
	if err = h.q.User.Select(selection...).Create(&model); err != nil {
		return err
	}
	return nil
}

func (h *UserHandler) DeleteUser(id int64) (err error) {

	if err = h.q.User.DeleteByID(id); err != nil {
		return err
	}
	return err
}

func (h *UserHandler) UpdateUser(user param.UserBody, id int64) error {
	selection, model := user.Data()
	if _, err := h.q.User.Select(selection...).
		Where(h.q.User.ID.Eq(uint(id))).
		Updates(&model); err != nil {
		return err
	}
	return nil
}

func (h *UserHandler) GetUserDetail(id int64) (param.UserResponse, error) {

	user, err := h.q.User.GetById(int(id))
	if err != nil {
		return param.UserResponse{}, err
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
		RoleID:       user.RoleID,
		DepartmentID: user.DepartmentID,
		ID:           user.ID,
	}, nil
}
