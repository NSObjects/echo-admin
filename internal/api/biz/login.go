/*
 *
 * login.go
 * biz
 *
 * Created by lintao on 2023/11/9 16:18
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package biz

import (
	"context"
	"github.com/NSObjects/echo-admin/internal/api/data"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/code"
	"github.com/NSObjects/echo-admin/internal/configs"
	"github.com/NSObjects/echo-admin/tools"
	"github.com/golang-jwt/jwt/v5"
	"github.com/marmotedu/errors"
	"time"
)

type LoginHandler struct {
	q   *query.Query
	cfg configs.Config
}

func NewLoginHandler(q *query.Query, cfg configs.Config) *LoginHandler {
	return &LoginHandler{q: q, cfg: cfg}
}

func (h *LoginHandler) Login(ctx context.Context, account, password string) (response param.LoginResponse, err error) {

	byAccount, err := h.q.User.WithContext(ctx).Where(h.q.User.Account.Eq(account)).First()
	if err != nil {
		return param.LoginResponse{}, err
	}

	if byAccount.Password != tools.Sha25(password) {
		return param.LoginResponse{}, errors.WithCode(code.ErrPasswordIncorrect, "密码错误")
	}

	claims := &data.JwtCustomClaims{
		Name:  byAccount.Name,
		Admin: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(h.cfg.JWT.Expire))),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(h.cfg.JWT.Secret))
	if err != nil {
		return param.LoginResponse{}, err
	}

	return param.LoginResponse{Token: t}, nil

}
