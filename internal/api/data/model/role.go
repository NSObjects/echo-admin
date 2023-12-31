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

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Name   string `json:"name" form:"name" query:"name"`
	Order  int    `json:"order" form:"order" query:"order"`
	Mark   string `json:"mark" form:"mark" query:"mark"`
	Status int    `gorm:"column:status"`
	Menus  []Menu `gorm:"many2many:role_menus;" json:"-"`
	//User      []User `gorm:"many2many:user_role;" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Role) TableName() string {
	return "role"
}
