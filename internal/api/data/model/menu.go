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
	ID        uint           `gorm:"primaryKey" json:"id" copier:"ID"`
	Name      string         `json:"name" gorm:"default:null" copier:"Name"`
	Path      string         `json:"path" gorm:"default:null" copier:"Path"`
	Component string         `json:"component" gorm:"default:null" copier:"Component"`
	Redirect  string         `json:"redirect" gorm:"default:null" copier:"Redirect"`
	Layout    int            `json:"layout" gorm:"default:null" copier:"Layout"`
	Icon      string         `json:"icon" gorm:"default:null" copier:"Icon"`
	Type      MenuType       `json:"type" gorm:"default:null" copier:"Type"`
	Remark    string         `gorm:"column:remark" copier:"Remark"`
	Link      string         `json:"link" copier:"Link"`
	Sort      int            `json:"sort"  copier:"Sort"`
	Hidden    int            `json:"hidden" copier:"Hidden"`
	Pid       *int64         `gorm:"column:pid" json:"pid,omitempty" copier:"Pid"`
	Children  []*Menu        `json:"children,omitempty" gorm:"foreignKey:Pid;references:ID" copier:"Children"`
	API       []API          `gorm:"many2many:menu_apis" copier:"API"`
	CreatedAt time.Time      `json:"created_at" copier:"CreatedAt"`
	UpdatedAt time.Time      `json:"updated_at" copier:"UpdatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Menu) TableName() string {
	return "menu"
}
