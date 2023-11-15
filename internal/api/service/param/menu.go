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
	"time"
)

type Menu struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Redirect  string `json:"redirect"`
	ParentId  int64  `json:"parent_id,omitempty"`
	Routes    []Menu `json:"routes"`
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
