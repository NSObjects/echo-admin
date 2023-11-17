/*
 * Created by lintao on 2023/7/18 下午3:59
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package data

import (
	"fmt"
	"github.com/NSObjects/echo-admin/internal/api/data/db"
	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/configs"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Model = fx.Options(
	fx.Provide(db.NewDataSource, NewQuery),
)

func NewQuery(cfg configs.Config) *query.Query {
	if cfg.Mysql.Host == "" {
		panic("mysql config is empty")
	}
	query.SetDefault(NewMysql(cfg.Mysql))
	return query.Q
}
func NewMysql(cfg configs.MysqlConfig) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}
