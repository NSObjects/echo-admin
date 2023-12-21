/*
 *
 * department.go
 * param
 *
 * Created by lintao on 2023/11/21 10:50
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package param

import (
	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"gorm.io/gen/field"
)

type DepartmentQuery struct {
	APIQuery
	Name   string `json:"name"`
	Status int    `json:"status"`
}

type Department struct {
	Name      *string `json:"name"`
	ParentID  *int64  `json:"parent_id"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Status    *int    `json:"status"`
	Sort      *int    `json:"sort"`
	Principal *string `json:"principal" `
}

func (r Department) Data() ([]field.Expr, model.Department) {
	var filed []field.Expr
	var value model.Department

	if r.ParentID != nil && *r.ParentID != 0 {
		filed = append(filed, query.Department.ParentID)
		value.ParentID = *r.ParentID
	}

	if r.Email != nil && *r.Email != "" {
		filed = append(filed, query.Department.Email)
		value.Email = *r.Email
	}

	if r.Name != nil && *r.Name != "" {
		filed = append(filed, query.Department.Name)
		value.Name = *r.Name
	}

	if r.Status != nil && *r.Status != 0 {
		filed = append(filed, query.Department.Status)
		value.Status = *r.Status
	}

	if r.Principal != nil && *r.Principal != "" {
		filed = append(filed, query.Department.Principal)
		value.Principal = *r.Principal
	}

	if r.Sort != nil && *r.Sort != 0 {
		filed = append(filed, query.Department.Sort)
		value.Sort = *r.Sort
	}

	if r.ParentID != nil && *r.ParentID != 0 {
		filed = append(filed, query.Department.ParentID)
		value.ParentID = *r.ParentID
	}

	return filed, value
}
