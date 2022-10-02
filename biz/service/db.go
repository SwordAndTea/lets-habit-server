package service

import (
	"github.com/swordandtea/fhwh/biz/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var globalDBExecutor *gorm.DB

func InitDB() error {
	db, err := gorm.Open(mysql.Open(config.GlobalConfig.Mysql.DSN))
	if err != nil {
		return err
	}
	globalDBExecutor = db
	return nil
}

func GetDBExecutor() *gorm.DB {
	return globalDBExecutor
}
