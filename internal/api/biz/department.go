/*
 *
 * department.go
 * biz
 *
 * Created by lintao on 2023/11/21 10:47
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package biz

import (
	"context"
	"database/sql"
	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"gorm.io/gen"
)

type DepartmentHandler struct {
	q *query.Query
}

func NewDepartmentHandler(q *query.Query) *DepartmentHandler {
	return &DepartmentHandler{q: q}
}

func (d *DepartmentHandler) Get(ctx context.Context, id uint) (*model.Department, error) {
	dep, err := d.q.Department.WithContext(ctx).
		Preload(d.q.Department.Principal).
		Where(d.q.Department.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}

	return dep, nil
}

func (d *DepartmentHandler) Create(ctx context.Context, department param.Department) error {
	var de model.Department
	de.Name = department.Name
	if department.ParentID != nil {
		de.Departments = []model.Department{
			{
				ID: *department.ParentID,
			},
		}
	}

	de.Email = department.Email
	de.Phone = department.Phone
	de.Status = department.Status
	de.Sort = department.Sort
	if department.PrincipalID != nil {
		de.Principal = &model.User{ID: *department.PrincipalID}
	}

	if err := d.q.Department.WithContext(ctx).Create(&de); err != nil {
		return err
	}
	return nil

}

func (d *DepartmentHandler) List(ctx context.Context, q param.DepartmentQuery) ([]*model.Department, int64, error) {
	var cd []gen.Condition
	if q.Name != "" {
		cd = append(cd, d.q.Department.Name.Eq(q.Name))
	}
	if q.Status != 0 {
		cd = append(cd, d.q.Department.Where(d.q.Department.Status.Eq(q.Status)))
	}

	deps, err := d.q.Department.WithContext(ctx).
		Preload(d.q.Department.Departments).
		Where(d.q.Department.ParentID.IsNull()).
		Where(cd...).Offset(q.Offset()).Limit(q.Limit()).Find()

	if err != nil {
		return nil, 0, err
	}

	total, err := d.q.Department.WithContext(ctx).Where(cd...).Count()
	if err != nil {
		return nil, 0, err
	}

	return deps, total, nil
}

func (d *DepartmentHandler) Delete(ctx context.Context, id uint) error {
	_, err := d.q.Department.WithContext(ctx).Where(d.q.Department.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}

	return nil
}

func (d *DepartmentHandler) Update(ctx context.Context, id uint, department param.Department) error {
	var update = make(map[string]interface{})
	if department.Name != "" {
		update["name"] = department.Name
	}

	if department.Email != "" {
		update["email"] = department.Email
	}
	if department.Phone != "" {
		update["phone"] = department.Phone
	}
	if department.Status != 0 {
		update["status"] = department.Status
	}
	if department.Sort != 0 {
		update["sort"] = department.Sort
	}
	if department.PrincipalID != nil {
		if *department.PrincipalID > 0 {
			update["principal_id"] = department.PrincipalID
		} else {
			update["principal_id"] = sql.NullInt64{}
		}
	}

	if department.ParentID != nil {
		if *department.ParentID > 0 {
			update["parent_id"] = department.ParentID
		} else {
			update["parent_id"] = sql.NullInt64{}
		}
	}

	_, err := d.q.Department.WithContext(ctx).Where(d.q.Department.ID.Eq(id)).Updates(update)
	if err != nil {
		return err
	}

	return nil
}
