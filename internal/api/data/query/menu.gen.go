// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"github.com/NSObjects/echo-admin/internal/api/data/model"
)

func newMenu(db *gorm.DB, opts ...gen.DOOption) menu {
	_menu := menu{}

	_menu.menuDo.UseDB(db, opts...)
	_menu.menuDo.UseModel(&model.Menu{})

	tableName := _menu.menuDo.TableName()
	_menu.ALL = field.NewAsterisk(tableName)
	_menu.ID = field.NewUint(tableName, "id")
	_menu.Name = field.NewString(tableName, "name")
	_menu.Path = field.NewString(tableName, "path")
	_menu.Component = field.NewString(tableName, "component")
	_menu.Redirect = field.NewString(tableName, "redirect")
	_menu.Layout = field.NewInt(tableName, "layout")
	_menu.Icon = field.NewString(tableName, "icon")
	_menu.Type = field.NewInt(tableName, "type")
	_menu.Remark = field.NewString(tableName, "remark")
	_menu.API = field.NewString(tableName, "api")
	_menu.Link = field.NewString(tableName, "link")
	_menu.Identifier = field.NewString(tableName, "identifier")
	_menu.Sort = field.NewInt(tableName, "sort")
	_menu.Hidden = field.NewInt(tableName, "hidden")
	_menu.Cache = field.NewInt(tableName, "cache")
	_menu.Fixed = field.NewInt(tableName, "fixed")
	_menu.Pid = field.NewInt64(tableName, "pid")
	_menu.CreatedAt = field.NewTime(tableName, "created_at")
	_menu.UpdatedAt = field.NewTime(tableName, "updated_at")
	_menu.DeletedAt = field.NewField(tableName, "deleted_at")
	_menu.Children = menuHasManyChildren{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Children", "model.Menu"),
		Children: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Children.Children", "model.Menu"),
		},
		RoleMenus: struct {
			field.RelationField
			Menus struct {
				field.RelationField
			}
			User struct {
				field.RelationField
				Role struct {
					field.RelationField
				}
			}
		}{
			RelationField: field.NewRelation("Children.RoleMenus", "model.Role"),
			Menus: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Children.RoleMenus.Menus", "model.Menu"),
			},
			User: struct {
				field.RelationField
				Role struct {
					field.RelationField
				}
			}{
				RelationField: field.NewRelation("Children.RoleMenus.User", "model.User"),
				Role: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Children.RoleMenus.User.Role", "model.Role"),
				},
			},
		},
	}

	_menu.RoleMenus = menuManyToManyRoleMenus{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("RoleMenus", "model.Role"),
	}

	_menu.fillFieldMap()

	return _menu
}

type menu struct {
	menuDo

	ALL        field.Asterisk
	ID         field.Uint
	Name       field.String
	Path       field.String
	Component  field.String
	Redirect   field.String
	Layout     field.Int
	Icon       field.String
	Type       field.Int
	Remark     field.String
	API        field.String
	Link       field.String
	Identifier field.String
	Sort       field.Int
	Hidden     field.Int
	Cache      field.Int
	Fixed      field.Int
	Pid        field.Int64
	CreatedAt  field.Time
	UpdatedAt  field.Time
	DeletedAt  field.Field
	Children   menuHasManyChildren

	RoleMenus menuManyToManyRoleMenus

	fieldMap map[string]field.Expr
}

func (m menu) Table(newTableName string) *menu {
	m.menuDo.UseTable(newTableName)
	return m.updateTableName(newTableName)
}

func (m menu) As(alias string) *menu {
	m.menuDo.DO = *(m.menuDo.As(alias).(*gen.DO))
	return m.updateTableName(alias)
}

func (m *menu) updateTableName(table string) *menu {
	m.ALL = field.NewAsterisk(table)
	m.ID = field.NewUint(table, "id")
	m.Name = field.NewString(table, "name")
	m.Path = field.NewString(table, "path")
	m.Component = field.NewString(table, "component")
	m.Redirect = field.NewString(table, "redirect")
	m.Layout = field.NewInt(table, "layout")
	m.Icon = field.NewString(table, "icon")
	m.Type = field.NewInt(table, "type")
	m.Remark = field.NewString(table, "remark")
	m.API = field.NewString(table, "api")
	m.Link = field.NewString(table, "link")
	m.Identifier = field.NewString(table, "identifier")
	m.Sort = field.NewInt(table, "sort")
	m.Hidden = field.NewInt(table, "hidden")
	m.Cache = field.NewInt(table, "cache")
	m.Fixed = field.NewInt(table, "fixed")
	m.Pid = field.NewInt64(table, "pid")
	m.CreatedAt = field.NewTime(table, "created_at")
	m.UpdatedAt = field.NewTime(table, "updated_at")
	m.DeletedAt = field.NewField(table, "deleted_at")

	m.fillFieldMap()

	return m
}

func (m *menu) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := m.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (m *menu) fillFieldMap() {
	m.fieldMap = make(map[string]field.Expr, 22)
	m.fieldMap["id"] = m.ID
	m.fieldMap["name"] = m.Name
	m.fieldMap["path"] = m.Path
	m.fieldMap["component"] = m.Component
	m.fieldMap["redirect"] = m.Redirect
	m.fieldMap["layout"] = m.Layout
	m.fieldMap["icon"] = m.Icon
	m.fieldMap["type"] = m.Type
	m.fieldMap["remark"] = m.Remark
	m.fieldMap["api"] = m.API
	m.fieldMap["link"] = m.Link
	m.fieldMap["identifier"] = m.Identifier
	m.fieldMap["sort"] = m.Sort
	m.fieldMap["hidden"] = m.Hidden
	m.fieldMap["cache"] = m.Cache
	m.fieldMap["fixed"] = m.Fixed
	m.fieldMap["pid"] = m.Pid
	m.fieldMap["created_at"] = m.CreatedAt
	m.fieldMap["updated_at"] = m.UpdatedAt
	m.fieldMap["deleted_at"] = m.DeletedAt

}

func (m menu) clone(db *gorm.DB) menu {
	m.menuDo.ReplaceConnPool(db.Statement.ConnPool)
	return m
}

func (m menu) replaceDB(db *gorm.DB) menu {
	m.menuDo.ReplaceDB(db)
	return m
}

type menuHasManyChildren struct {
	db *gorm.DB

	field.RelationField

	Children struct {
		field.RelationField
	}
	RoleMenus struct {
		field.RelationField
		Menus struct {
			field.RelationField
		}
		User struct {
			field.RelationField
			Role struct {
				field.RelationField
			}
		}
	}
}

func (a menuHasManyChildren) Where(conds ...field.Expr) *menuHasManyChildren {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a menuHasManyChildren) WithContext(ctx context.Context) *menuHasManyChildren {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a menuHasManyChildren) Session(session *gorm.Session) *menuHasManyChildren {
	a.db = a.db.Session(session)
	return &a
}

func (a menuHasManyChildren) Model(m *model.Menu) *menuHasManyChildrenTx {
	return &menuHasManyChildrenTx{a.db.Model(m).Association(a.Name())}
}

type menuHasManyChildrenTx struct{ tx *gorm.Association }

func (a menuHasManyChildrenTx) Find() (result []*model.Menu, err error) {
	return result, a.tx.Find(&result)
}

func (a menuHasManyChildrenTx) Append(values ...*model.Menu) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a menuHasManyChildrenTx) Replace(values ...*model.Menu) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a menuHasManyChildrenTx) Delete(values ...*model.Menu) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a menuHasManyChildrenTx) Clear() error {
	return a.tx.Clear()
}

func (a menuHasManyChildrenTx) Count() int64 {
	return a.tx.Count()
}

type menuManyToManyRoleMenus struct {
	db *gorm.DB

	field.RelationField
}

func (a menuManyToManyRoleMenus) Where(conds ...field.Expr) *menuManyToManyRoleMenus {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a menuManyToManyRoleMenus) WithContext(ctx context.Context) *menuManyToManyRoleMenus {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a menuManyToManyRoleMenus) Session(session *gorm.Session) *menuManyToManyRoleMenus {
	a.db = a.db.Session(session)
	return &a
}

func (a menuManyToManyRoleMenus) Model(m *model.Menu) *menuManyToManyRoleMenusTx {
	return &menuManyToManyRoleMenusTx{a.db.Model(m).Association(a.Name())}
}

type menuManyToManyRoleMenusTx struct{ tx *gorm.Association }

func (a menuManyToManyRoleMenusTx) Find() (result []*model.Role, err error) {
	return result, a.tx.Find(&result)
}

func (a menuManyToManyRoleMenusTx) Append(values ...*model.Role) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a menuManyToManyRoleMenusTx) Replace(values ...*model.Role) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a menuManyToManyRoleMenusTx) Delete(values ...*model.Role) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a menuManyToManyRoleMenusTx) Clear() error {
	return a.tx.Clear()
}

func (a menuManyToManyRoleMenusTx) Count() int64 {
	return a.tx.Count()
}

type menuDo struct{ gen.DO }

type IMenuDo interface {
	gen.SubQuery
	Debug() IMenuDo
	WithContext(ctx context.Context) IMenuDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IMenuDo
	WriteDB() IMenuDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IMenuDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IMenuDo
	Not(conds ...gen.Condition) IMenuDo
	Or(conds ...gen.Condition) IMenuDo
	Select(conds ...field.Expr) IMenuDo
	Where(conds ...gen.Condition) IMenuDo
	Order(conds ...field.Expr) IMenuDo
	Distinct(cols ...field.Expr) IMenuDo
	Omit(cols ...field.Expr) IMenuDo
	Join(table schema.Tabler, on ...field.Expr) IMenuDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IMenuDo
	RightJoin(table schema.Tabler, on ...field.Expr) IMenuDo
	Group(cols ...field.Expr) IMenuDo
	Having(conds ...gen.Condition) IMenuDo
	Limit(limit int) IMenuDo
	Offset(offset int) IMenuDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IMenuDo
	Unscoped() IMenuDo
	Create(values ...*model.Menu) error
	CreateInBatches(values []*model.Menu, batchSize int) error
	Save(values ...*model.Menu) error
	First() (*model.Menu, error)
	Take() (*model.Menu, error)
	Last() (*model.Menu, error)
	Find() ([]*model.Menu, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.Menu, err error)
	FindInBatches(result *[]*model.Menu, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.Menu) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IMenuDo
	Assign(attrs ...field.AssignExpr) IMenuDo
	Joins(fields ...field.RelationField) IMenuDo
	Preload(fields ...field.RelationField) IMenuDo
	FirstOrInit() (*model.Menu, error)
	FirstOrCreate() (*model.Menu, error)
	FindByPage(offset int, limit int) (result []*model.Menu, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IMenuDo
	UnderlyingDB() *gorm.DB
	schema.Tabler

	GetById(id int) (result model.Menu, err error)
	DeleteByID(id int64) (err error)
}

// GetById
// SELECT * FROM @@table WHERE id = @id
func (m menuDo) GetById(id int) (result model.Menu, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, id)
	generateSQL.WriteString("SELECT * FROM menu WHERE id = ? ")

	var executeSQL *gorm.DB
	executeSQL = m.UnderlyingDB().Raw(generateSQL.String(), params...).Take(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// DeleteByID
// DELETE FROM @@table WHERE id = @id
func (m menuDo) DeleteByID(id int64) (err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, id)
	generateSQL.WriteString("DELETE FROM menu WHERE id = ? ")

	var executeSQL *gorm.DB
	executeSQL = m.UnderlyingDB().Exec(generateSQL.String(), params...) // ignore_security_alert
	err = executeSQL.Error

	return
}

func (m menuDo) Debug() IMenuDo {
	return m.withDO(m.DO.Debug())
}

func (m menuDo) WithContext(ctx context.Context) IMenuDo {
	return m.withDO(m.DO.WithContext(ctx))
}

func (m menuDo) ReadDB() IMenuDo {
	return m.Clauses(dbresolver.Read)
}

func (m menuDo) WriteDB() IMenuDo {
	return m.Clauses(dbresolver.Write)
}

func (m menuDo) Session(config *gorm.Session) IMenuDo {
	return m.withDO(m.DO.Session(config))
}

func (m menuDo) Clauses(conds ...clause.Expression) IMenuDo {
	return m.withDO(m.DO.Clauses(conds...))
}

func (m menuDo) Returning(value interface{}, columns ...string) IMenuDo {
	return m.withDO(m.DO.Returning(value, columns...))
}

func (m menuDo) Not(conds ...gen.Condition) IMenuDo {
	return m.withDO(m.DO.Not(conds...))
}

func (m menuDo) Or(conds ...gen.Condition) IMenuDo {
	return m.withDO(m.DO.Or(conds...))
}

func (m menuDo) Select(conds ...field.Expr) IMenuDo {
	return m.withDO(m.DO.Select(conds...))
}

func (m menuDo) Where(conds ...gen.Condition) IMenuDo {
	return m.withDO(m.DO.Where(conds...))
}

func (m menuDo) Order(conds ...field.Expr) IMenuDo {
	return m.withDO(m.DO.Order(conds...))
}

func (m menuDo) Distinct(cols ...field.Expr) IMenuDo {
	return m.withDO(m.DO.Distinct(cols...))
}

func (m menuDo) Omit(cols ...field.Expr) IMenuDo {
	return m.withDO(m.DO.Omit(cols...))
}

func (m menuDo) Join(table schema.Tabler, on ...field.Expr) IMenuDo {
	return m.withDO(m.DO.Join(table, on...))
}

func (m menuDo) LeftJoin(table schema.Tabler, on ...field.Expr) IMenuDo {
	return m.withDO(m.DO.LeftJoin(table, on...))
}

func (m menuDo) RightJoin(table schema.Tabler, on ...field.Expr) IMenuDo {
	return m.withDO(m.DO.RightJoin(table, on...))
}

func (m menuDo) Group(cols ...field.Expr) IMenuDo {
	return m.withDO(m.DO.Group(cols...))
}

func (m menuDo) Having(conds ...gen.Condition) IMenuDo {
	return m.withDO(m.DO.Having(conds...))
}

func (m menuDo) Limit(limit int) IMenuDo {
	return m.withDO(m.DO.Limit(limit))
}

func (m menuDo) Offset(offset int) IMenuDo {
	return m.withDO(m.DO.Offset(offset))
}

func (m menuDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IMenuDo {
	return m.withDO(m.DO.Scopes(funcs...))
}

func (m menuDo) Unscoped() IMenuDo {
	return m.withDO(m.DO.Unscoped())
}

func (m menuDo) Create(values ...*model.Menu) error {
	if len(values) == 0 {
		return nil
	}
	return m.DO.Create(values)
}

func (m menuDo) CreateInBatches(values []*model.Menu, batchSize int) error {
	return m.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (m menuDo) Save(values ...*model.Menu) error {
	if len(values) == 0 {
		return nil
	}
	return m.DO.Save(values)
}

func (m menuDo) First() (*model.Menu, error) {
	if result, err := m.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.Menu), nil
	}
}

func (m menuDo) Take() (*model.Menu, error) {
	if result, err := m.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.Menu), nil
	}
}

func (m menuDo) Last() (*model.Menu, error) {
	if result, err := m.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.Menu), nil
	}
}

func (m menuDo) Find() ([]*model.Menu, error) {
	result, err := m.DO.Find()
	return result.([]*model.Menu), err
}

func (m menuDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.Menu, err error) {
	buf := make([]*model.Menu, 0, batchSize)
	err = m.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (m menuDo) FindInBatches(result *[]*model.Menu, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return m.DO.FindInBatches(result, batchSize, fc)
}

func (m menuDo) Attrs(attrs ...field.AssignExpr) IMenuDo {
	return m.withDO(m.DO.Attrs(attrs...))
}

func (m menuDo) Assign(attrs ...field.AssignExpr) IMenuDo {
	return m.withDO(m.DO.Assign(attrs...))
}

func (m menuDo) Joins(fields ...field.RelationField) IMenuDo {
	for _, _f := range fields {
		m = *m.withDO(m.DO.Joins(_f))
	}
	return &m
}

func (m menuDo) Preload(fields ...field.RelationField) IMenuDo {
	for _, _f := range fields {
		m = *m.withDO(m.DO.Preload(_f))
	}
	return &m
}

func (m menuDo) FirstOrInit() (*model.Menu, error) {
	if result, err := m.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.Menu), nil
	}
}

func (m menuDo) FirstOrCreate() (*model.Menu, error) {
	if result, err := m.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.Menu), nil
	}
}

func (m menuDo) FindByPage(offset int, limit int) (result []*model.Menu, count int64, err error) {
	result, err = m.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = m.Offset(-1).Limit(-1).Count()
	return
}

func (m menuDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = m.Count()
	if err != nil {
		return
	}

	err = m.Offset(offset).Limit(limit).Scan(result)
	return
}

func (m menuDo) Scan(result interface{}) (err error) {
	return m.DO.Scan(result)
}

func (m menuDo) Delete(models ...*model.Menu) (result gen.ResultInfo, err error) {
	return m.DO.Delete(models)
}

func (m *menuDo) withDO(do gen.Dao) *menuDo {
	m.DO = *do.(*gen.DO)
	return m
}
