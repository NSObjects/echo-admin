/*
 * Created by lintao on 2023/7/18 下午3:59
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package biz

import (
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/tools"
	"gorm.io/gen/field"

	"github.com/NSObjects/echo-admin/internal/api/data/model"
)

type UserHandler struct {
	q *query.Query
}

func NewUserHandler(q *query.Query) *UserHandler {
	return &UserHandler{q: q}
}

func (h *UserHandler) ListUser(u model.User, p param.APIQuery) ([]param.UserResponse, int64, error) {

	users, total, err := h.q.User.Where(field.Attrs(u)).FindByPage(p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, err
	}
	//users, total, err := h.repository.FindUser(u, p)
	//if err != nil {
	//	return nil, 0, err
	//}

	resp := make([]param.UserResponse, len(users))
	for i, user := range users {
		resp[i] = param.UserResponse{
			Name:     user.Name,
			Phone:    user.Phone,
			Status:   user.Status,
			Password: user.Password,
		}
	}

	return resp, total, nil
}

func (h *UserHandler) CreateUser(param model.User) (err error) {
	param.Password = tools.Sha25(param.Password)
	if err = h.q.User.Create(&param); err != nil {
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

func (h *UserHandler) UpdateUser(user model.User, id int64) error {

	if _, err := h.q.User.Where(h.q.User.ID.Eq(uint(id))).Updates(&user); err != nil {
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
		Name:     user.Name,
		Phone:    user.Phone,
		Status:   user.Status,
		Password: user.Password,
	}, nil
}
