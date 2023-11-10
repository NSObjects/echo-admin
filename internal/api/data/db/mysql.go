/*
 * Created by lintao on 2023/7/26 下午3:02
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package db

import (
	"fmt"
	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"gorm.io/gen"

	"github.com/NSObjects/echo-admin/internal/configs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Querier Dynamic SQL
type Querier interface {
	// SELECT * FROM @@table WHERE id = @id
	GetById(id int) (gen.T, error)
	// DELEL * FROM @@table WHERE id = @id
	DeleteByID(id int64) error
}

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
	g := gen.NewGenerator(gen.Config{
		OutPath: "query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})

	g.UseDB(db) // reuse your gorm db

	// Generate basic type-safe DAO API for struct `model.User` following conventions
	g.ApplyBasic(model.User{})

	// Generate Type Safe API with Dynamic SQL defined on Querier interface for `model.User` and `model.Company`
	g.ApplyInterface(func(Querier) {}, model.User{}, model.Menu{})

	// Generate the code
	g.Execute()
	return db
}
