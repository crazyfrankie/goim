package main

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("MYSQL_DSN")
	db := connectDB(dsn)

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
