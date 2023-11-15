/*
 *
 * menu.go
 * param
 *
 * Created by lintao on 2023/11/14 11:07
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package param

import (
	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/query"
	"gorm.io/gen/field"
	"time"
)

type Menu struct {
	Name      *string `json:"name"`
	Path      *string `json:"path"`
	Component *string `json:"component"`
	Redirect  *string `json:"redirect"`
	ParentID  *int64  `json:"parent_id,omitempty"`
	Layout    *bool   `json:"layout,omitempty" `
	Routes    []Menu  `json:"routes"`
}

type MenuResp struct {
	ID        uint         `json:"id"`
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Component string       `json:"component"`
	Redirect  string       `json:"redirect"`
	ParentId  int64        `json:"parent_id,omitempty"`
	Routes    []model.Menu `json:"routes"`
	CreatedAt time.Time    `json:"created_at" `
	UpdatedAt time.Time    `json:"updated_at" `
}

func (m Menu) Data() ([]field.Expr, model.Menu) {

	var filed []field.Expr
	var model model.Menu
	if m.Name != nil {
		filed = append(filed, query.Q.Menu.Name)
		model.Name = *m.Name
	}
	if m.Path != nil {
		filed = append(filed, query.Q.Menu.Path)
		model.Path = *m.Path
	}
	if m.Component != nil {
		filed = append(filed, query.Q.Menu.Component)
		model.Component = *m.Component
	}
	if m.Redirect != nil {
		filed = append(filed, query.Q.Menu.Redirect)
		model.Redirect = *m.Redirect
	}
	if m.Layout != nil {
		filed = append(filed, query.Q.Menu.Layout)
		model.Layout = *m.Layout
	}
	if m.ParentID != nil {
		filed = append(filed, query.Q.Menu.ParentID)
		model.ParentID = *m.ParentID
	}

	return filed, model
}
