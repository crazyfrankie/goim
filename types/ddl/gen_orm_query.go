package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	db := connectDB("root:041126@tcp(localhost:3306)/goim?charset=utf8mb4&parseTime=True")

	genUser(db)
}

func genUser(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath:      "apps/user/domain/internal/dal/query",
		ModelPkgPath: "apps/user/domain/internal/dal/model",
		Mode:         gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.UseDB(db)

	g.ApplyBasic(g.GenerateModel("user"))

	g.Execute()
}

func connectDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(fmt.Errorf("connect db fail: %w", err))
	}
	return db
}
