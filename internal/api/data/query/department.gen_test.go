// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm/clause"
)

func init() {
	InitializeDB()
	err := db.AutoMigrate(&model.Department{})
	if err != nil {
		fmt.Printf("Error: AutoMigrate(&model.Department{}) fail: %s", err)
	}
}

func Test_departmentQuery(t *testing.T) {
	department := newDepartment(db)
	department = *department.As(department.TableName())
	_do := department.WithContext(context.Background()).Debug()

	primaryKey := field.NewString(department.TableName(), clause.PrimaryKey)
	_, err := _do.Unscoped().Where(primaryKey.IsNotNull()).Delete()
	if err != nil {
		t.Error("clean table <department> fail:", err)
		return
	}

	_, ok := department.GetFieldByName("")
	if ok {
		t.Error("GetFieldByName(\"\") from department success")
	}

	err = _do.Create(&model.Department{})
	if err != nil {
		t.Error("create item in table <department> fail:", err)
	}

	err = _do.Save(&model.Department{})
	if err != nil {
		t.Error("create item in table <department> fail:", err)
	}

	err = _do.CreateInBatches([]*model.Department{{}, {}}, 10)
	if err != nil {
		t.Error("create item in table <department> fail:", err)
	}

	_, err = _do.Select(department.ALL).Take()
	if err != nil {
		t.Error("Take() on table <department> fail:", err)
	}

	_, err = _do.First()
	if err != nil {
		t.Error("First() on table <department> fail:", err)
	}

	_, err = _do.Last()
	if err != nil {
		t.Error("First() on table <department> fail:", err)
	}

	_, err = _do.Where(primaryKey.IsNotNull()).FindInBatch(10, func(tx gen.Dao, batch int) error { return nil })
	if err != nil {
		t.Error("FindInBatch() on table <department> fail:", err)
	}

	err = _do.Where(primaryKey.IsNotNull()).FindInBatches(&[]*model.Department{}, 10, func(tx gen.Dao, batch int) error { return nil })
	if err != nil {
		t.Error("FindInBatches() on table <department> fail:", err)
	}

	_, err = _do.Select(department.ALL).Where(primaryKey.IsNotNull()).Order(primaryKey.Desc()).Find()
	if err != nil {
		t.Error("Find() on table <department> fail:", err)
	}

	_, err = _do.Distinct(primaryKey).Take()
	if err != nil {
		t.Error("select Distinct() on table <department> fail:", err)
	}

	_, err = _do.Select(department.ALL).Omit(primaryKey).Take()
	if err != nil {
		t.Error("Omit() on table <department> fail:", err)
	}

	_, err = _do.Group(primaryKey).Find()
	if err != nil {
		t.Error("Group() on table <department> fail:", err)
	}

	_, err = _do.Scopes(func(dao gen.Dao) gen.Dao { return dao.Where(primaryKey.IsNotNull()) }).Find()
	if err != nil {
		t.Error("Scopes() on table <department> fail:", err)
	}

	_, _, err = _do.FindByPage(0, 1)
	if err != nil {
		t.Error("FindByPage() on table <department> fail:", err)
	}

	_, err = _do.ScanByPage(&model.Department{}, 0, 1)
	if err != nil {
		t.Error("ScanByPage() on table <department> fail:", err)
	}

	_, err = _do.Attrs(primaryKey).Assign(primaryKey).FirstOrInit()
	if err != nil {
		t.Error("FirstOrInit() on table <department> fail:", err)
	}

	_, err = _do.Attrs(primaryKey).Assign(primaryKey).FirstOrCreate()
	if err != nil {
		t.Error("FirstOrCreate() on table <department> fail:", err)
	}

	var _a _another
	var _aPK = field.NewString(_a.TableName(), "id")

	err = _do.Join(&_a, primaryKey.EqCol(_aPK)).Scan(map[string]interface{}{})
	if err != nil {
		t.Error("Join() on table <department> fail:", err)
	}

	err = _do.LeftJoin(&_a, primaryKey.EqCol(_aPK)).Scan(map[string]interface{}{})
	if err != nil {
		t.Error("LeftJoin() on table <department> fail:", err)
	}

	_, err = _do.Not().Or().Clauses().Take()
	if err != nil {
		t.Error("Not/Or/Clauses on table <department> fail:", err)
	}
}

var DepartmentGetByIdTestCase = []TestCase{}

func Test_department_GetById(t *testing.T) {
	department := newDepartment(db)
	do := department.WithContext(context.Background()).Debug()

	for i, tt := range DepartmentGetByIdTestCase {
		t.Run("GetById_"+strconv.Itoa(i), func(t *testing.T) {
			res1, res2 := do.GetById(tt.Input.Args[0].(int))
			assert(t, "GetById", res1, tt.Expectation.Ret[0])
			assert(t, "GetById", res2, tt.Expectation.Ret[1])
		})
	}
}

var DepartmentDeleteByIDTestCase = []TestCase{}

func Test_department_DeleteByID(t *testing.T) {
	department := newDepartment(db)
	do := department.WithContext(context.Background()).Debug()

	for i, tt := range DepartmentDeleteByIDTestCase {
		t.Run("DeleteByID_"+strconv.Itoa(i), func(t *testing.T) {
			res1 := do.DeleteByID(tt.Input.Args[0].(int64))
			assert(t, "DeleteByID", res1, tt.Expectation.Ret[0])
		})
	}
}
