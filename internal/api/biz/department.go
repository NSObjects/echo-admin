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

	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/api/service/param"
	"github.com/google/martian/log"
	"github.com/samber/lo"
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
		//Preload(d.q.Department.Principal).
		Where(d.q.Department.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}

	return dep, nil
}

func (d *DepartmentHandler) Create(ctx context.Context, department param.Department) error {
	selection, m := department.Data()

	if err := d.q.Department.WithContext(ctx).Select(selection...).Create(&m); err != nil {
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
		//Preload(d.q.Department.Departments).
		Where(d.q.Department.ParentID.IsNull()).
		Where(cd...).Offset(q.Offset()).Limit(q.Limit()).Find()

	if err != nil {
		return nil, 0, err
	}

	for _, td := range deps {
		td.Departments, err = d.GetAllDepartments(td.ID)
		if err != nil {
			log.Errorf("get all departments error: %v", err)
		}
	}

	total, err := d.q.Department.WithContext(ctx).Where(cd...).Count()
	if err != nil {
		return nil, 0, err
	}

	return deps, total, nil
}

func (d *DepartmentHandler) GetAllDepartments(parentID uint) ([]model.Department, error) {
	if parentID == 0 {
		return nil, nil
	}
	departments, err := d.q.Department.Where(d.q.Department.ParentID.Eq(int64(parentID))).
		Preload(d.q.Department.Departments).Find()
	if err != nil {
		return nil, err
	}
	for i, department := range departments {
		children, err := d.GetAllDepartments(department.ID)
		if err != nil {
			return nil, err
		}
		departments[i].Departments = children
	}

	return lo.Map(departments, func(item *model.Department, index int) model.Department {
		return *item
	}), nil
}

func (d *DepartmentHandler) Delete(ctx context.Context, id uint) error {
	_, err := d.q.Department.WithContext(ctx).Where(d.q.Department.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}

	return nil
}

func (d *DepartmentHandler) Update(ctx context.Context, id uint, department param.Department) error {
	selection, m := department.Data()
	if _, err := d.q.Department.WithContext(ctx).Select(selection...).
		Where(d.q.Department.ID.Eq(id)).
		Updates(&m); err != nil {
		return err
	}

	return nil
}
