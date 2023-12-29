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
r = sub, menu,obj, act

[policy_definition]
p = sub,menu, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && g(r.menu, p.menu) && keyMatch2(r.obj, p.obj) && r.act == p.act || r.sub == "root"
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
