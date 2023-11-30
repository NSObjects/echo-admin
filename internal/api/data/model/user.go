/*
 * Created by lintao on 2023/7/18 下午3:59
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Name         string `json:"name,omitempty" form:"name" query:"name"`
	Phone        string `json:"phone,omitempty" form:"phone" query:"phone"`
	Status       int64  `json:"status,omitempty" form:"status" query:"status"`
	Account      string `json:"account,omitempty" form:"account" query:"account"`
	Password     string `json:"password,omitempty" form:"password" query:"password"`
	DepartmentID uint   `json:"department_id,omitempty" form:"department_id" query:"department_id"`
	Role         []Role `gorm:"many2many:user_role;" json:"role,omitempty"`
	Avatar       string
	CreatedAt    time.Time      `json:"created_at" `
	UpdatedAt    time.Time      `json:"updated_at" `
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "user"
}
