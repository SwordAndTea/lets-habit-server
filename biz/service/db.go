package service

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var globalDBExecutor *gorm.DB

func InitDB(dsn string) error {
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		return err
	}
	globalDBExecutor = db
	return nil
}

func GetDBExecutor() *gorm.DB {
	return globalDBExecutor
}
