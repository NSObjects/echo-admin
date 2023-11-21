/*
 *
 * casbin.go
 * db
 *
 * Created by lintao on 2023/11/21 14:29
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package biz

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

func NewCasbin(db *gorm.DB) *casbin.Enforcer {
	a, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		panic(err)
	}

	m, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act || r.sub == "root"
`)
	if err != nil {
		panic(err)
	}

	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		panic(err)
	}

	return e
}
