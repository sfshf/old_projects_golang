package batch

import (
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// Dynamic SQL
type Querier interface {
	// SELECT * FROM @@table WHERE name = @name{{if role !=""}} AND role = @role{{end}}
	FilterWithNameAndRole(name, role string) ([]gen.T, error)
}

func GenerateModels(dbPath string) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "/Users/lxm/Documents/fullstack/Go/word/internal/app/dao",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})

	// dbPath := "root:waf12KFkwo2@tcp(127.0.0.1:3306)/word?charset=utf8&interpolateParams=True"
	gormdb, _ := gorm.Open(mysql.Open(dbPath))
	g.UseDB(gormdb) // reuse your gorm db

	//   // Generate basic type-safe DAO API for struct `model.User` following conventions
	//   g.ApplyBasic(model.User{})

	// Generate Type Safe API with Dynamic SQL defined on Querier interface for `model.User` and `model.Company`
	//   g.ApplyInterface(func(Querier){}, model.User{}, model.Company{})

	// Generate the code
	g.GenerateAllTable()
	g.Execute()
}
