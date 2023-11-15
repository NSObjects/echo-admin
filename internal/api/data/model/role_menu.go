/*
 *
 * role_menu.go
 * model
 *
 * Created by lintao on 2023/11/15 15:49
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package model

import "gorm.io/gorm"

type RoleMenu struct {
	gorm.Model
	RoleId int64 `json:"role_id" form:"role_id" query:"role_id"`
	MenuId int64 `json:"menu_id" form:"menu_id" query:"menu_id"`
}
