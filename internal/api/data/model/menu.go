/*
 *
 * menu.go
 * db
 *
 * Created by lintao on 2023/11/10 16:35
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package model

import (
	"time"

	"gorm.io/gorm"
)

type MenuType int

const (
	MenuTypeDir MenuType = iota
	MenuTypeMenu
	MenuTypeButton
)

type Menu struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name" gorm:"default:null"`
	Path      string         `json:"path" gorm:"default:null"`
	Component string         `json:"component" gorm:"default:null"`
	Redirect  string         `json:"redirect" gorm:"default:null"`
	Layout    bool           `json:"layout" gorm:"default:null"`
	Icon      string         `json:"icon" gorm:"default:null"`
	Type      MenuType       `json:"type" gorm:"default:null"`
	ParentID  int64          `json:"parent_id,omitempty" gorm:"default:null"`
	Remark    string         `gorm:"column:remark"`
	Children  []Menu         `json:"children,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	RoleMenus []Role         `gorm:"many2many:role_menus;" json:"-"`
	CreatedAt time.Time      `json:"created_at" `
	UpdatedAt time.Time      `json:"updated_at" `
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Menu) TableName() string {
	return "menu"
}
