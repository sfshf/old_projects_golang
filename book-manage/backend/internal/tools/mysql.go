package tools

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	db  *gorm.DB
	err error
)

func InitMysql(dbPath string) error {
	fmt.Println("InitMysql: ", dbPath)
	db, err = gorm.Open(mysql.Open(dbPath), &gorm.Config{
		PrepareStmt: false,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}

func MysqlDB() *gorm.DB {
	if db == nil {
		panic(any("db not inited"))
	}

	return db
}
