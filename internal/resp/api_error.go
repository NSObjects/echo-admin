/*
 * Created by lintao on 2023/7/18 下午3:56
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package resp

import (
	"errors"
	"github.com/NSObjects/echo-admin/internal/code"
)

type Error struct {
	Err  error
	Code int
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func NewError(err error, code int) *Error {
	return &Error{
		Err:  err,
		Code: code,
	}
}

func NewParamError(err error) *Error {
	return &Error{
		Err:  err,
		Code: code.ErrValidation,
	}
}

func NewDBError(err error) *Error {
	return &Error{
		Err:  err,
		Code: code.ErrDatabase,
	}
}

func NewMsgError(str string) *Error {
	return &Error{
		Err:  errors.New(str),
		Code: code.ErrParentMenuExisted,
	}
}

func NewAuthError(err error) *Error {
	return &Error{
		Err:  err,
		Code: code.ErrInvalidAuthHeader,
	}
}
