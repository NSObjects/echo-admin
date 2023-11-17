/*
 * Created by lintao on 2023/7/26 下午3:02
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package db

import (
	"fmt"

	"github.com/NSObjects/echo-admin/internal/api/data/query"
	"github.com/NSObjects/echo-admin/internal/configs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Basic CRUD
//GetUserByID(id int64) (user model.User, err error)
//FindUser(param model.User, query param.APIQuery) (user []model.User, total int64, err error)
//DeleteUserByID(id int64) (err error)
//UpdateUser(param model.User, id int64) (err error)
//CreateUser(param model.User) (id int64, err error)
//GetUserByAccount(ctx context.Context, account string) (model.User, error)

func NewMysql(cfg configs.MysqlConfig) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	query.SetDefault(db)
	return db
}
