/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/NSObjects/echo-admin/internal/api/data/model"
	"github.com/NSObjects/echo-admin/internal/configs"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"

	"github.com/spf13/cobra"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := configs.InitConfig(cfgFile)
		if err != nil {
			panic(err)
		}
		GenMysql(configs.Mysql)
		fmt.Println("gen called")
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
}

// Querier Dynamic SQL
type Querier interface {
	// GetById
	// SELECT * FROM @@table WHERE id = @id
	GetById(id int) (gen.T, error)

	// DeleteByID
	// DELETE * FROM @@table WHERE id = @id
	DeleteByID(id int64) error
}

func GenMysql(cfg configs.MysqlConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&model.User{}, &model.Menu{}, &model.Role{}, &model.RoleMenu{})
	if err != nil {
		panic(err)
	}
	g := gen.NewGenerator(gen.Config{
		OutPath: "query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode

	})
	g.UseDB(db)

	// Generate basic type-safe DAO API for struct `model.User` following conventions
	g.ApplyBasic(model.User{})
	// Generate Type Safe API with Dynamic SQL defined on Querier interface for `model.User` and `model.Company`
	g.ApplyInterface(func(Querier) {}, model.User{}, model.Menu{}, model.Role{}, model.RoleMenu{})

	// Generate the code
	g.Execute()
}
