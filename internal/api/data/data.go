/*
 * Created by lintao on 2023/7/18 下午3:59
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package data

import (
	"github.com/NSObjects/echo-admin/internal/api/data/db"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"go.uber.org/fx"
)

var Model = fx.Options(
	fx.Provide(db.NewDataSource, NewQuery),
)

func NewQuery() *query.Query {
	return query.Q
}
