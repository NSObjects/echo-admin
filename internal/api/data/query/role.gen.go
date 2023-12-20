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

func newRole(db *gorm.DB, opts ...gen.DOOption) role {
	_role := role{}

	_role.roleDo.UseDB(db, opts...)
	_role.roleDo.UseModel(&model.Role{})

	tableName := _role.roleDo.TableName()
	_role.ALL = field.NewAsterisk(tableName)
	_role.ID = field.NewUint(tableName, "id")
	_role.Name = field.NewString(tableName, "name")
	_role.Order_ = field.NewInt(tableName, "order")
	_role.Mark = field.NewString(tableName, "mark")
	_role.Status = field.NewInt(tableName, "status")
	_role.CreatedAt = field.NewTime(tableName, "created_at")
	_role.UpdatedAt = field.NewTime(tableName, "updated_at")
	_role.DeletedAt = field.NewField(tableName, "deleted_at")
	_role.Menus = roleManyToManyMenus{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Menus", "model.Menu"),
		Children: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Menus.Children", "model.Menu"),
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
			RelationField: field.NewRelation("Menus.RoleMenus", "model.Role"),
			Menus: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Menus.RoleMenus.Menus", "model.Menu"),
			},
			User: struct {
				field.RelationField
				Role struct {
					field.RelationField
				}
			}{
				RelationField: field.NewRelation("Menus.RoleMenus.User", "model.User"),
				Role: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Menus.RoleMenus.User.Role", "model.Role"),
				},
			},
		},
	}

	_role.User = roleManyToManyUser{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("User", "model.User"),
	}

	_role.fillFieldMap()

	return _role
}

type role struct {
	roleDo

	ALL       field.Asterisk
	ID        field.Uint
	Name      field.String
	Order_    field.Int
	Mark      field.String
	Status    field.Int
	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Field
	Menus     roleManyToManyMenus

	User roleManyToManyUser

	fieldMap map[string]field.Expr
}

func (r role) Table(newTableName string) *role {
	r.roleDo.UseTable(newTableName)
	return r.updateTableName(newTableName)
}

func (r role) As(alias string) *role {
	r.roleDo.DO = *(r.roleDo.As(alias).(*gen.DO))
	return r.updateTableName(alias)
}

func (r *role) updateTableName(table string) *role {
	r.ALL = field.NewAsterisk(table)
	r.ID = field.NewUint(table, "id")
	r.Name = field.NewString(table, "name")
	r.Order_ = field.NewInt(table, "order")
	r.Mark = field.NewString(table, "mark")
	r.Status = field.NewInt(table, "status")
	r.CreatedAt = field.NewTime(table, "created_at")
	r.UpdatedAt = field.NewTime(table, "updated_at")
	r.DeletedAt = field.NewField(table, "deleted_at")

	r.fillFieldMap()

	return r
}

func (r *role) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := r.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (r *role) fillFieldMap() {
	r.fieldMap = make(map[string]field.Expr, 10)
	r.fieldMap["id"] = r.ID
	r.fieldMap["name"] = r.Name
	r.fieldMap["order"] = r.Order_
	r.fieldMap["mark"] = r.Mark
	r.fieldMap["status"] = r.Status
	r.fieldMap["created_at"] = r.CreatedAt
	r.fieldMap["updated_at"] = r.UpdatedAt
	r.fieldMap["deleted_at"] = r.DeletedAt

}

func (r role) clone(db *gorm.DB) role {
	r.roleDo.ReplaceConnPool(db.Statement.ConnPool)
	return r
}

func (r role) replaceDB(db *gorm.DB) role {
	r.roleDo.ReplaceDB(db)
	return r
}

type roleManyToManyMenus struct {
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

func (a roleManyToManyMenus) Where(conds ...field.Expr) *roleManyToManyMenus {
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

func (a roleManyToManyMenus) WithContext(ctx context.Context) *roleManyToManyMenus {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a roleManyToManyMenus) Session(session *gorm.Session) *roleManyToManyMenus {
	a.db = a.db.Session(session)
	return &a
}

func (a roleManyToManyMenus) Model(m *model.Role) *roleManyToManyMenusTx {
	return &roleManyToManyMenusTx{a.db.Model(m).Association(a.Name())}
}

type roleManyToManyMenusTx struct{ tx *gorm.Association }

func (a roleManyToManyMenusTx) Find() (result []*model.Menu, err error) {
	return result, a.tx.Find(&result)
}

func (a roleManyToManyMenusTx) Append(values ...*model.Menu) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a roleManyToManyMenusTx) Replace(values ...*model.Menu) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a roleManyToManyMenusTx) Delete(values ...*model.Menu) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a roleManyToManyMenusTx) Clear() error {
	return a.tx.Clear()
}

func (a roleManyToManyMenusTx) Count() int64 {
	return a.tx.Count()
}

type roleManyToManyUser struct {
	db *gorm.DB

	field.RelationField
}

func (a roleManyToManyUser) Where(conds ...field.Expr) *roleManyToManyUser {
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

func (a roleManyToManyUser) WithContext(ctx context.Context) *roleManyToManyUser {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a roleManyToManyUser) Session(session *gorm.Session) *roleManyToManyUser {
	a.db = a.db.Session(session)
	return &a
}

func (a roleManyToManyUser) Model(m *model.Role) *roleManyToManyUserTx {
	return &roleManyToManyUserTx{a.db.Model(m).Association(a.Name())}
}

type roleManyToManyUserTx struct{ tx *gorm.Association }

func (a roleManyToManyUserTx) Find() (result []*model.User, err error) {
	return result, a.tx.Find(&result)
}

func (a roleManyToManyUserTx) Append(values ...*model.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a roleManyToManyUserTx) Replace(values ...*model.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a roleManyToManyUserTx) Delete(values ...*model.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a roleManyToManyUserTx) Clear() error {
	return a.tx.Clear()
}

func (a roleManyToManyUserTx) Count() int64 {
	return a.tx.Count()
}

type roleDo struct{ gen.DO }

type IRoleDo interface {
	gen.SubQuery
	Debug() IRoleDo
	WithContext(ctx context.Context) IRoleDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IRoleDo
	WriteDB() IRoleDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IRoleDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IRoleDo
	Not(conds ...gen.Condition) IRoleDo
	Or(conds ...gen.Condition) IRoleDo
	Select(conds ...field.Expr) IRoleDo
	Where(conds ...gen.Condition) IRoleDo
	Order(conds ...field.Expr) IRoleDo
	Distinct(cols ...field.Expr) IRoleDo
	Omit(cols ...field.Expr) IRoleDo
	Join(table schema.Tabler, on ...field.Expr) IRoleDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IRoleDo
	RightJoin(table schema.Tabler, on ...field.Expr) IRoleDo
	Group(cols ...field.Expr) IRoleDo
	Having(conds ...gen.Condition) IRoleDo
	Limit(limit int) IRoleDo
	Offset(offset int) IRoleDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IRoleDo
	Unscoped() IRoleDo
	Create(values ...*model.Role) error
	CreateInBatches(values []*model.Role, batchSize int) error
	Save(values ...*model.Role) error
	First() (*model.Role, error)
	Take() (*model.Role, error)
	Last() (*model.Role, error)
	Find() ([]*model.Role, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.Role, err error)
	FindInBatches(result *[]*model.Role, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.Role) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IRoleDo
	Assign(attrs ...field.AssignExpr) IRoleDo
	Joins(fields ...field.RelationField) IRoleDo
	Preload(fields ...field.RelationField) IRoleDo
	FirstOrInit() (*model.Role, error)
	FirstOrCreate() (*model.Role, error)
	FindByPage(offset int, limit int) (result []*model.Role, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IRoleDo
	UnderlyingDB() *gorm.DB
	schema.Tabler

	GetById(id int) (result model.Role, err error)
	DeleteByID(id int64) (err error)
}

// GetById
// SELECT * FROM @@table WHERE id = @id
func (r roleDo) GetById(id int) (result model.Role, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, id)
	generateSQL.WriteString("SELECT * FROM role WHERE id = ? ")

	var executeSQL *gorm.DB
	executeSQL = r.UnderlyingDB().Raw(generateSQL.String(), params...).Take(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// DeleteByID
// DELETE FROM @@table WHERE id = @id
func (r roleDo) DeleteByID(id int64) (err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, id)
	generateSQL.WriteString("DELETE FROM role WHERE id = ? ")

	var executeSQL *gorm.DB
	executeSQL = r.UnderlyingDB().Exec(generateSQL.String(), params...) // ignore_security_alert
	err = executeSQL.Error

	return
}

func (r roleDo) Debug() IRoleDo {
	return r.withDO(r.DO.Debug())
}

func (r roleDo) WithContext(ctx context.Context) IRoleDo {
	return r.withDO(r.DO.WithContext(ctx))
}

func (r roleDo) ReadDB() IRoleDo {
	return r.Clauses(dbresolver.Read)
}

func (r roleDo) WriteDB() IRoleDo {
	return r.Clauses(dbresolver.Write)
}

func (r roleDo) Session(config *gorm.Session) IRoleDo {
	return r.withDO(r.DO.Session(config))
}

func (r roleDo) Clauses(conds ...clause.Expression) IRoleDo {
	return r.withDO(r.DO.Clauses(conds...))
}

func (r roleDo) Returning(value interface{}, columns ...string) IRoleDo {
	return r.withDO(r.DO.Returning(value, columns...))
}

func (r roleDo) Not(conds ...gen.Condition) IRoleDo {
	return r.withDO(r.DO.Not(conds...))
}

func (r roleDo) Or(conds ...gen.Condition) IRoleDo {
	return r.withDO(r.DO.Or(conds...))
}

func (r roleDo) Select(conds ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Select(conds...))
}

func (r roleDo) Where(conds ...gen.Condition) IRoleDo {
	return r.withDO(r.DO.Where(conds...))
}

func (r roleDo) Order(conds ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Order(conds...))
}

func (r roleDo) Distinct(cols ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Distinct(cols...))
}

func (r roleDo) Omit(cols ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Omit(cols...))
}

func (r roleDo) Join(table schema.Tabler, on ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Join(table, on...))
}

func (r roleDo) LeftJoin(table schema.Tabler, on ...field.Expr) IRoleDo {
	return r.withDO(r.DO.LeftJoin(table, on...))
}

func (r roleDo) RightJoin(table schema.Tabler, on ...field.Expr) IRoleDo {
	return r.withDO(r.DO.RightJoin(table, on...))
}

func (r roleDo) Group(cols ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Group(cols...))
}

func (r roleDo) Having(conds ...gen.Condition) IRoleDo {
	return r.withDO(r.DO.Having(conds...))
}

func (r roleDo) Limit(limit int) IRoleDo {
	return r.withDO(r.DO.Limit(limit))
}

func (r roleDo) Offset(offset int) IRoleDo {
	return r.withDO(r.DO.Offset(offset))
}

func (r roleDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IRoleDo {
	return r.withDO(r.DO.Scopes(funcs...))
}

func (r roleDo) Unscoped() IRoleDo {
	return r.withDO(r.DO.Unscoped())
}

func (r roleDo) Create(values ...*model.Role) error {
	if len(values) == 0 {
		return nil
	}
	return r.DO.Create(values)
}

func (r roleDo) CreateInBatches(values []*model.Role, batchSize int) error {
	return r.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (r roleDo) Save(values ...*model.Role) error {
	if len(values) == 0 {
		return nil
	}
	return r.DO.Save(values)
}

func (r roleDo) First() (*model.Role, error) {
	if result, err := r.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.Role), nil
	}
}

func (r roleDo) Take() (*model.Role, error) {
	if result, err := r.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.Role), nil
	}
}

func (r roleDo) Last() (*model.Role, error) {
	if result, err := r.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.Role), nil
	}
}

func (r roleDo) Find() ([]*model.Role, error) {
	result, err := r.DO.Find()
	return result.([]*model.Role), err
}

func (r roleDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.Role, err error) {
	buf := make([]*model.Role, 0, batchSize)
	err = r.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (r roleDo) FindInBatches(result *[]*model.Role, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return r.DO.FindInBatches(result, batchSize, fc)
}

func (r roleDo) Attrs(attrs ...field.AssignExpr) IRoleDo {
	return r.withDO(r.DO.Attrs(attrs...))
}

func (r roleDo) Assign(attrs ...field.AssignExpr) IRoleDo {
	return r.withDO(r.DO.Assign(attrs...))
}

func (r roleDo) Joins(fields ...field.RelationField) IRoleDo {
	for _, _f := range fields {
		r = *r.withDO(r.DO.Joins(_f))
	}
	return &r
}

func (r roleDo) Preload(fields ...field.RelationField) IRoleDo {
	for _, _f := range fields {
		r = *r.withDO(r.DO.Preload(_f))
	}
	return &r
}

func (r roleDo) FirstOrInit() (*model.Role, error) {
	if result, err := r.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.Role), nil
	}
}

func (r roleDo) FirstOrCreate() (*model.Role, error) {
	if result, err := r.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.Role), nil
	}
}

func (r roleDo) FindByPage(offset int, limit int) (result []*model.Role, count int64, err error) {
	result, err = r.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = r.Offset(-1).Limit(-1).Count()
	return
}

func (r roleDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = r.Count()
	if err != nil {
		return
	}

	err = r.Offset(offset).Limit(limit).Scan(result)
	return
}

func (r roleDo) Scan(result interface{}) (err error) {
	return r.DO.Scan(result)
}

func (r roleDo) Delete(models ...*model.Role) (result gen.ResultInfo, err error) {
	return r.DO.Delete(models)
}

func (r *roleDo) withDO(do gen.Dao) *roleDo {
	r.DO = *do.(*gen.DO)
	return r
}
