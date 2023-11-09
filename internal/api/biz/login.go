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
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/NSObjects/echo-admin/internal/code"
	"github.com/NSObjects/echo-admin/tools"
	"github.com/golang-jwt/jwt/v5"
	"github.com/marmotedu/errors"
	"time"
)

var secret = []byte("tn)M^P<j,/6$Gr/Wrs")

type jwtCustomClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

type LoginHandler struct {
	repository data.UserRepository
}

func NewLoginHandler(r data.UserRepository) *LoginHandler {
	return &LoginHandler{repository: r}
}

func (h *LoginHandler) Login(ctx context.Context, account, password string) (response param.LoginResponse, err error) {
	byAccount, err := h.repository.GetUserByAccount(ctx, account)
	if err != nil {
		return param.LoginResponse{}, err
	}

	if byAccount.Password != tools.Sha25(password) {
		return param.LoginResponse{}, errors.WithCode(code.ErrPasswordIncorrect, "密码错误")
	}

	claims := &jwtCustomClaims{
		byAccount.Name,
		true,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString(secret)
	if err != nil {
		return param.LoginResponse{}, err
	}

	return param.LoginResponse{Token: t}, nil

}
