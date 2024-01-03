/*
 *
 * api.go
 * model
 *
 * Created by lintao on 2024/1/2 11:21
 * Copyright © 2020-2024 LINTAO. All rights reserved.
 *
 */

package model

type API struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(50);not null" json:"name"`
	Path        string `gorm:"column:path;type:varchar(50);not null;uniqueIndex:udx_name_path" json:"path"`
	Method      string `gorm:"column:method;type:varchar(50);not null;uniqueIndex:udx_name_path" json:"method"`
	Description string `gorm:"column:description;type:varchar(50);not null" json:"description"`
	APIGroup    string `gorm:"column:api_group;type:varchar(50);not null" json:"api_group"`
	CreatedAt   string `gorm:"column:created_at;type:varchar(50);not null" json:"created_at"`
	UpdatedAt   string `gorm:"column:updated_at;type:varchar(50);not null" json:"updated_at"`
}

func (API) TableName() string {
	return "api"
}
