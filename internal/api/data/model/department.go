/*
 *
 * department.go
 * model
 *
 * Created by lintao on 2023/11/21 10:31
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package model

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `json:"name" gorm:"default:null"`
	Status      int            `json:"status" gorm:"default:null"`
	Sort        int            `json:"sort" gorm:"default:null"`
	PrincipalID uint           `json:"principal_id" gorm:"default:null"`
	Principal   *User          `json:"principal,omitempty" gorm:"default:null"`
	ParentID    int64          `json:"parent_id,omitempty" gorm:"default:null"`
	Departments []Department   `json:"departments,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Phone       string         `json:"phone" gorm:"default:null"`
	Email       string         `json:"email" gorm:"default:null"`
	CreatedAt   time.Time      `json:"created_at" `
	UpdatedAt   time.Time      `json:"updated_at" `
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Department) TableName() string {
	return "department"
}
