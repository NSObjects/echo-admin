// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameAPI = "api"

// API mapped from table <api>
type API struct {
	ID          int64  `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement:true" json:"id"`
	Name        string `gorm:"column:name;type:varchar(50);not null" json:"name"`
	Path        string `gorm:"column:path;type:varchar(50);not null;uniqueIndex:udx_name_path,priority:1" json:"path"`
	Method      string `gorm:"column:method;type:varchar(50);not null;uniqueIndex:udx_name_path,priority:2" json:"method"`
	Description string `gorm:"column:description;type:varchar(50);not null" json:"description"`
	APIGroup    string `gorm:"column:api_group;type:varchar(50);not null" json:"api_group"`
	CreatedAt   string `gorm:"column:created_at;type:varchar(50);not null" json:"created_at"`
	UpdatedAt   string `gorm:"column:updated_at;type:varchar(50);not null" json:"updated_at"`
}

// TableName API's table name
func (*API) TableName() string {
	return TableNameAPI
}
