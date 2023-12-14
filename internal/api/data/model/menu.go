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
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `json:"name" gorm:"default:null"`
	Path       string         `json:"path" gorm:"default:null"`
	Component  string         `json:"component" gorm:"default:null"`
	Redirect   string         `json:"redirect" gorm:"default:null"`
	Layout     int            `json:"layout" gorm:"default:null"`
	Icon       string         `json:"icon" gorm:"default:null"`
	Type       MenuType       `json:"type" gorm:"default:null"`
	Remark     string         `gorm:"column:remark"`
	API        string         `json:"api"  copier:"api"`
	Link       string         `json:"link"  copier:"link"`
	Identifier string         `json:"identifier" copier:"identifier" `
	Sort       int            `json:"sort"  copier:"sort"`
	Hidden     int            `json:"hidden" copier:"hidden"`
	Cache      int            `json:"cache"  copier:"cache"`
	Fixed      int            `json:"fixed"  copier:"fixed"`
	Pid        int64          `json:"pid,omitempty"  copier:"pid"`
	Children   []Menu         `json:"children,omitempty" gorm:"foreignKey:Pid;references:ID"`
	RoleMenus  []Role         `gorm:"many2many:role_menus;" json:"-"`
	CreatedAt  time.Time      `json:"created_at" `
	UpdatedAt  time.Time      `json:"updated_at" `
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Menu) TableName() string {
	return "menu"
}
