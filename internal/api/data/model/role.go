/*
 *
 * role.go
 * model
 *
 * Created by lintao on 2023/11/15 15:40
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package model

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name     string `json:"name" form:"name" query:"name"`
	Order    int    `json:"order" form:"order" query:"order"`
	Identify string `json:"identify" form:"identify" query:"identify"`
	State    int    `json:"state" `
	Menus    []Menu `gorm:"many2many:role_menus;"`
	User     []User `gorm:"many2many:user_role;"`
}

func (Role) TableName() string {
	return "role"
}
