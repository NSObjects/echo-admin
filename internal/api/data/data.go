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
	"gorm.io/gorm/logger"
)

var Model = fx.Options(
	fx.Provide(db.NewDataSource, NewQuery, NewMysql),
)

func NewQuery(cfg configs.Config) *query.Query {
	if cfg.Mysql.Host == "" {
		panic("mysql config is empty")
	}

	return query.Q
}

func NewMysql(cfg configs.Config) *gorm.DB {
	if cfg.Mysql.Host == "" {
		panic("mysql config is empty")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.Mysql.User, cfg.Mysql.Password, cfg.Mysql.Host, cfg.Mysql.Port, cfg.Mysql.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		panic(err)
	}

	query.SetDefault(db)
	return db
}
