// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameRoleMenu = "role_menus"

// RoleMenu mapped from table <role_menus>
type RoleMenu struct {
	RoleID int64 `gorm:"column:role_id;type:bigint unsigned;primaryKey" json:"role_id"`
	MenuID int64 `gorm:"column:menu_id;type:bigint unsigned;primaryKey;index:fk_role_menus_menu,priority:1" json:"menu_id"`
}

// TableName RoleMenu's table name
func (*RoleMenu) TableName() string {
	return TableNameRoleMenu
}
